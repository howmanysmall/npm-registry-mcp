package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
	"github.com/howmanysmall/npm-registry-mcp/src/github"
	"github.com/howmanysmall/npm-registry-mcp/src/health"
	"github.com/howmanysmall/npm-registry-mcp/src/license"
	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const statusUnknown = "unknown"

// InstallInput is the input for the should-i-install tool
type InstallInput struct {
	Package string `json:"package" jsonschema:"NPM package name to evaluate"`
}

// InstallOutput is the comprehensive health check output
type InstallOutput struct {
	Package      string          `json:"package"`
	Version      string          `json:"version"`
	Verdict      string          `json:"verdict"`
	Score        int             `json:"score"`
	Maintenance  MaintenanceInfo `json:"maintenance"`
	Dependencies DependencyInfo  `json:"dependencies"`
	Security     SecurityInfo    `json:"security"`
	Popularity   PopularityInfo  `json:"popularity"`
	Size         SizeInfo        `json:"size"`
	TypeScript   TypeScriptInfo  `json:"typescript"`
	License      LicenseInfo     `json:"license"`
	Warnings     []string        `json:"warnings,omitempty"`
}

// MaintenanceInfo contains maintenance status
type MaintenanceInfo struct {
	LastPublish string `json:"lastPublish"`
	Status      string `json:"status"`
}

// DependencyInfo contains dependency analysis
type DependencyInfo struct {
	Direct     int    `json:"direct"`
	Transitive int    `json:"transitive"`
	Outdated   int    `json:"outdated"`
	Status     string `json:"status"`
}

// SecurityInfo contains security status
type SecurityInfo struct {
	Vulnerabilities int    `json:"vulnerabilities"`
	Status          string `json:"status"`
}

// PopularityInfo contains popularity metrics
type PopularityInfo struct {
	WeeklyDownloads int    `json:"weeklyDownloads"`
	Trend           string `json:"trend"`
	Status          string `json:"status"`
}

// SizeInfo contains size information
type SizeInfo struct {
	Unpacked string `json:"unpacked"`
}

// TypeScriptInfo contains TypeScript support info
type TypeScriptInfo struct {
	Support string `json:"support"`
	Status  string `json:"status"`
}

// LicenseInfo contains license information
type LicenseInfo struct {
	SPDX string `json:"spdx"`
	Risk string `json:"risk"`
}

// InstallHandler is the handler type for should-i-install
type InstallHandler = func(context.Context, *mcp.CallToolRequest, InstallInput) (*mcp.CallToolResult, InstallOutput, error)

// NewInstallHandler creates a new install handler
func NewInstallHandler(npmClient *npm.Client, ghClient *github.Client, _ *cache.Cache) InstallHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input InstallInput) (*mcp.CallToolResult, InstallOutput, error) {
		// Get package info
		pkg, err := npmClient.GetPackage(ctx, input.Package)
		if err != nil {
			return nil, InstallOutput{}, fmt.Errorf("failed to get package: %w", err)
		}

		latestVersion := pkg.DistTags["latest"]
		latestMeta := pkg.Versions[latestVersion]

		// Get download stats
		weeklyDownloads := 0
		prevWeekDownloads := 0

		if dl, err := npmClient.GetDownloads(ctx, input.Package, "last-week"); err == nil {
			weeklyDownloads = dl.Downloads
		}

		if dl, err := npmClient.GetDownloads(ctx, input.Package, "last-month"); err == nil {
			// Approximate previous week from monthly
			prevWeekDownloads = dl.Downloads / 4
		}

		// Parse last publish time
		var lastPublish time.Time
		if t, ok := pkg.Time[latestVersion]; ok {
			if parsed, err := time.Parse(time.RFC3339, t); err == nil {
				lastPublish = parsed
			}
		}

		// Get GitHub info if available
		commitCount := 0
		openIssues := 0

		if pkg.Repository != nil && ghClient != nil {
			owner, repo := parseGitHubURL(pkg.Repository.URL)
			if owner != "" && repo != "" {
				if repoInfo, err := ghClient.GetRepository(ctx, owner, repo); err == nil {
					openIssues = repoInfo.OpenIssuesCount
				}

				if commits, err := ghClient.GetCommits(ctx, owner, repo, 90); err == nil {
					commitCount = len(commits)
				}
			}
		}

		// Check TypeScript support
		tsSupport, tsStatus := getTypeScriptSupport(latestMeta)

		// Assess license risk
		licenseRisk := license.AssessRisk(pkg.License)

		// Calculate health score
		healthInput := health.Input{
			LastPublish:       lastPublish,
			WeeklyDownloads:   weeklyDownloads,
			PrevWeekDownloads: prevWeekDownloads,
			DirectDeps:        len(latestMeta.Dependencies),
			OutdatedDeps:      0, // Would need additional API call
			CommitCount90d:    commitCount,
			OpenIssues:        openIssues,
			MaintainerCount:   len(pkg.Maintainers),
			VulnCount:         0, // Would need security API
		}

		healthResult := health.CalculateScore(healthInput)

		// Format size
		sizeStr := statusUnknown
		if latestMeta.Dist.UnpackedSize > 0 {
			sizeStr = formatBytes(latestMeta.Dist.UnpackedSize)
		}

		// Format download trend
		trend := getDownloadTrend(weeklyDownloads, prevWeekDownloads)

		// Determine popularity status
		popStatus := getPopularityStatus(weeklyDownloads)

		// Build output
		output := InstallOutput{
			Package: pkg.Name,
			Version: latestVersion,
			Verdict: string(healthResult.Verdict),
			Score:   healthResult.Score,

			Maintenance: MaintenanceInfo{
				LastPublish: formatTimeAgo(lastPublish),
				Status:      getMaintenanceStatus(healthResult.Factors),
			},

			Dependencies: DependencyInfo{
				Direct:     len(latestMeta.Dependencies),
				Transitive: 0, // Would need full tree resolution
				Outdated:   0,
				Status:     statusUnknown,
			},

			Security: SecurityInfo{
				Vulnerabilities: 0,
				Status:          statusUnknown,
			},

			Popularity: PopularityInfo{
				WeeklyDownloads: weeklyDownloads,
				Trend:           trend,
				Status:          popStatus,
			},

			Size: SizeInfo{
				Unpacked: sizeStr,
			},

			TypeScript: TypeScriptInfo{
				Support: tsSupport,
				Status:  tsStatus,
			},

			License: LicenseInfo{
				SPDX: pkg.License,
				Risk: string(licenseRisk.Level),
			},

			Warnings: healthResult.Warnings,
		}

		// Add license warnings if applicable
		if licenseRisk.Level == license.RiskHigh || licenseRisk.Level == license.RiskCritical {
			output.Warnings = append(output.Warnings, licenseRisk.Description)
		}

		return nil, output, nil
	}
}

// InstallTool returns the tool definition for should-i-install
func InstallTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "should-i-install",
		Description: "Comprehensive health check for an NPM package - evaluates maintenance, security, popularity, and license risk",
	}
}

var githubURLRegex = regexp.MustCompile(`github\.com[/:]([^/]+)/([^/.]+)`)

func parseGitHubURL(url string) (owner, repo string) {
	matches := githubURLRegex.FindStringSubmatch(url)
	if len(matches) >= 3 {
		return matches[1], strings.TrimSuffix(matches[2], ".git")
	}

	return "", ""
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return statusUnknown
	}

	days := int(time.Since(t).Hours() / 24)

	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "yesterday"
	case days < 30:
		return fmt.Sprintf("%d days ago", days)
	case days < 365:
		months := days / 30
		if months == 1 {
			return "1 month ago"
		}

		return fmt.Sprintf("%d months ago", months)
	default:
		years := days / 365
		if years == 1 {
			return "1 year ago"
		}

		return fmt.Sprintf("%d years ago", years)
	}
}

func getTypeScriptSupport(meta npm.PackageVersion) (support, status string) {
	if meta.Types != "" || meta.Typings != "" {
		return "included", "bundled"
	}
	// Could check for @types package, but skip for now
	return statusUnknown, "check @types"
}

func getDownloadTrend(current, previous int) string {
	if previous <= 0 {
		return "stable"
	}

	change := float64(current-previous) / float64(previous) * 100
	if change >= 10 {
		return "growing"
	} else if change <= -10 {
		return "declining"
	}

	return "stable"
}

func getPopularityStatus(weeklyDownloads int) string {
	switch {
	case weeklyDownloads >= 1000000:
		return "very popular"
	case weeklyDownloads >= 100000:
		return "popular"
	case weeklyDownloads >= 10000:
		return "moderate"
	case weeklyDownloads >= 1000:
		return "niche"
	default:
		return "low usage"
	}
}

// getMaintenanceStatus safely extracts maintenance status from health factors
func getMaintenanceStatus(factors []health.Factor) string {
	for _, f := range factors {
		if f.Name == "Last Publish" {
			return f.Status
		}
	}

	return statusUnknown
}
