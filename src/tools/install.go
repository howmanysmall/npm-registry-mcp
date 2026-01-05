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
		pkg, err := npmClient.GetPackage(ctx, input.Package)
		if err != nil {
			return nil, InstallOutput{}, fmt.Errorf("failed to get package: %w", err)
		}

		latestVersion := pkg.DistTags["latest"]
		latestMeta := pkg.Versions[latestVersion]

		weeklyDownloads, prevWeekDownloads := getDownloadStats(ctx, npmClient, input.Package)
		commitCount, openIssues := getGitHubStats(ctx, ghClient, pkg.Repository)

		lastPublish := parseLastPublish(pkg, latestVersion)
		tsSupport, tsStatus := getTypeScriptSupport(latestMeta)
		licenseRisk := license.AssessRisk(pkg.License)

		healthResult := health.CalculateScore(health.Input{
			LastPublish:       lastPublish,
			WeeklyDownloads:   weeklyDownloads,
			PrevWeekDownloads: prevWeekDownloads,
			DirectDeps:        len(latestMeta.Dependencies),
			OutdatedDeps:      0,
			CommitCount90d:    commitCount,
			OpenIssues:        openIssues,
			MaintainerCount:   len(pkg.Maintainers),
			VulnCount:         0,
		})

		output := buildInstallOutput(pkg, latestMeta, healthResult, lastPublish, weeklyDownloads, prevWeekDownloads, tsSupport, tsStatus, licenseRisk)

		return nil, output, nil
	}
}

func getDownloadStats(ctx context.Context, client *npm.Client, pkgName string) (int, int) {
	weeklyDownloads := 0
	prevWeekDownloads := 0

	if dl, err := client.GetDownloads(ctx, pkgName, "last-week"); err == nil {
		weeklyDownloads = dl.Downloads
	}

	if dl, err := client.GetDownloads(ctx, pkgName, "last-month"); err == nil {
		prevWeekDownloads = dl.Downloads / 4
	}

	return weeklyDownloads, prevWeekDownloads
}

func getGitHubStats(ctx context.Context, client *github.Client, repo *npm.Repository) (int, int) {
	if repo == nil || client == nil {
		return 0, 0
	}

	owner, repoName := parseGitHubURL(repo.URL)
	if owner == "" || repoName == "" {
		return 0, 0
	}

	openIssues := 0
	commitCount := 0

	if repoInfo, err := client.GetRepository(ctx, owner, repoName); err == nil {
		openIssues = repoInfo.OpenIssuesCount
	}

	if commits, err := client.GetCommits(ctx, owner, repoName, 90); err == nil {
		commitCount = len(commits)
	}

	return commitCount, openIssues
}

func parseLastPublish(pkg *npm.PackageResponse, version string) time.Time {
	if t, ok := pkg.Time[version]; ok {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

func buildInstallOutput(
	pkg *npm.PackageResponse,
	meta npm.PackageVersion,
	healthResult health.Result,
	lastPublish time.Time,
	weeklyDownloads, prevWeekDownloads int,
	tsSupport, tsStatus string,
	licenseRisk license.RiskAssessment,
) InstallOutput {
	sizeStr := statusUnknown
	if meta.Dist.UnpackedSize > 0 {
		sizeStr = formatBytes(meta.Dist.UnpackedSize)
	}

	output := InstallOutput{
		Package: pkg.Name,
		Version: meta.Version,
		Verdict: string(healthResult.Verdict),
		Score:   healthResult.Score,

		Maintenance: MaintenanceInfo{
			LastPublish: formatTimeAgo(lastPublish),
			Status:      getMaintenanceStatus(healthResult.Factors),
		},

		Dependencies: DependencyInfo{Direct: len(meta.Dependencies), Transitive: 0, Outdated: 0, Status: statusUnknown},

		Security: SecurityInfo{Vulnerabilities: 0, Status: statusUnknown},

		Popularity: PopularityInfo{
			WeeklyDownloads: weeklyDownloads,
			Trend:           getDownloadTrend(weeklyDownloads, prevWeekDownloads),
			Status:          getPopularityStatus(weeklyDownloads),
		},

		Size: SizeInfo{Unpacked: sizeStr},

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

	if licenseRisk.Level == license.RiskHigh || licenseRisk.Level == license.RiskCritical {
		output.Warnings = append(output.Warnings, licenseRisk.Description)
	}

	return output
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
