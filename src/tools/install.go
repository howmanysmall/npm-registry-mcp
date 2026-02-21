package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
	"github.com/howmanysmall/npm-registry-mcp/src/github"
	"github.com/howmanysmall/npm-registry-mcp/src/health"
	"github.com/howmanysmall/npm-registry-mcp/src/license"
	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusUnknown   = "unknown"
	supportIncluded = "included"
)

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

type installData struct {
	pkg               *npm.PackageResponse
	weeklyDownloads   int
	prevWeekDownloads int
	commitCount       int
	openIssues        int
	vulnCount         int
	tsSupport         string
	tsStatus          string
}

// NewInstallHandler creates a new install handler
func NewInstallHandler(npmClient *npm.Client, ghClient *github.Client, appCache *cache.Cache) InstallHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input InstallInput) (*mcp.CallToolResult, InstallOutput, error) {
		// Check cache
		cacheKey := fmt.Sprintf("install:%s", input.Package)
		if appCache != nil {
			if cached, found := cache.Get[InstallOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		data, err := fetchInstallData(ctx, npmClient, ghClient, input.Package)
		if err != nil {
			return nil, InstallOutput{}, err
		}

		latestVersion := data.pkg.DistTags["latest"]
		latestMeta := data.pkg.Versions[latestVersion]

		lastPublish := parseLastPublish(data.pkg, latestVersion)
		licenseRisk := license.AssessRisk(string(data.pkg.License))

		healthResult := health.CalculateScore(health.Input{
			LastPublish:       lastPublish,
			WeeklyDownloads:   data.weeklyDownloads,
			PrevWeekDownloads: data.prevWeekDownloads,
			DirectDeps:        len(latestMeta.Dependencies),
			OutdatedDeps:      0,
			CommitCount90d:    data.commitCount,
			OpenIssues:        data.openIssues,
			MaintainerCount:   len(data.pkg.Maintainers),
			VulnCount:         data.vulnCount,
			HasTypes:          data.tsSupport == supportIncluded,
			UnpackedSize:      latestMeta.Dist.UnpackedSize,
		})

		output := buildInstallOutput(data.pkg, latestMeta, healthResult, lastPublish, data.weeklyDownloads, data.prevWeekDownloads, data.tsSupport, data.tsStatus, licenseRisk)
		output.Security.Vulnerabilities = data.vulnCount

		if data.vulnCount > 0 {
			output.Security.Status = "vulnerable"
		} else {
			output.Security.Status = "secure"
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

func fetchInstallData(ctx context.Context, npmClient *npm.Client, ghClient *github.Client, pkgName string) (*installData, error) {
	data := &installData{}

	var pkgErr error

	var wg sync.WaitGroup

	// Fetch package metadata and download stats in parallel
	wg.Add(2)

	go func() {
		defer wg.Done()

		data.pkg, pkgErr = npmClient.GetPackage(ctx, pkgName)
	}()

	go func() {
		defer wg.Done()

		data.weeklyDownloads, data.prevWeekDownloads = getDownloadStats(ctx, npmClient, pkgName)
	}()

	wg.Wait()

	if pkgErr != nil {
		return nil, fmt.Errorf("failed to get package: %w", pkgErr)
	}

	latestVersion := data.pkg.DistTags["latest"]
	latestMeta := data.pkg.Versions[latestVersion]

	// Fetch GitHub stats, vulnerabilities, and types (depend on pkg)
	wg.Add(3)

	go func() {
		defer wg.Done()

		data.commitCount, data.openIssues = getGitHubStats(ctx, ghClient, data.pkg.Repository)
	}()

	go func() {
		defer wg.Done()

		advisories, err := npmClient.GetAdvisories(ctx, map[string][]string{
			pkgName: {latestVersion},
		})
		if err == nil {
			if pkgAdvisories, ok := advisories[pkgName]; ok {
				data.vulnCount = len(pkgAdvisories)
			}
		}
	}()

	go func() {
		defer wg.Done()

		data.tsSupport, data.tsStatus = getTypeScriptSupport(ctx, npmClient, pkgName, latestMeta)
	}()

	wg.Wait()

	return data, nil
}

func getDownloadStats(ctx context.Context, client *npm.Client, pkgName string) (int, int) {
	var (
		weeklyDownloads   int
		prevWeekDownloads int
		wg                sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()

		if dl, err := client.GetDownloads(ctx, pkgName, "last-week"); err == nil {
			weeklyDownloads = dl.Downloads
		}
	}()

	go func() {
		defer wg.Done()

		if dl, err := client.GetDownloads(ctx, pkgName, "last-month"); err == nil {
			prevWeekDownloads = dl.Downloads / 4
		}
	}()

	wg.Wait()

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

	var (
		openIssues  int
		commitCount int
		wg          sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()

		if repoInfo, err := client.GetRepository(ctx, owner, repoName); err == nil {
			openIssues = repoInfo.OpenIssuesCount
		}
	}()

	go func() {
		defer wg.Done()

		if commits, err := client.GetCommits(ctx, owner, repoName, 90); err == nil {
			commitCount = len(commits)
		}
	}()

	wg.Wait()

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
			SPDX: string(pkg.License),
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

// SecurityInput is the input for the check-package-security tool
type SecurityInput struct {
	Package string `json:"package" jsonschema:"NPM package name"`
	Version string `json:"version,omitempty" jsonschema:"specific version (defaults to latest)"`
}

// SecurityOutput is the output for the check-package-security tool
type SecurityOutput struct {
	Package         string         `json:"package"`
	Version         string         `json:"version"`
	Vulnerabilities []npm.Advisory `json:"vulnerabilities"`
	Summary         string         `json:"summary"`
}

// SecurityHandler is the handler type for check-package-security
type SecurityHandler = func(context.Context, *mcp.CallToolRequest, SecurityInput) (*mcp.CallToolResult, SecurityOutput, error)

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(client *npm.Client, appCache *cache.Cache) SecurityHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SecurityInput) (*mcp.CallToolResult, SecurityOutput, error) {
		version := input.Version
		if version == "" {
			pkg, err := client.GetAbbreviatedPackage(ctx, input.Package)
			if err != nil {
				return nil, SecurityOutput{}, err
			}

			version = pkg.DistTags["latest"]
		}

		// Check cache
		cacheKey := fmt.Sprintf("security:%s:%s", input.Package, version)
		if appCache != nil {
			if cached, found := cache.Get[SecurityOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		advisories, err := client.GetAdvisories(ctx, map[string][]string{
			input.Package: {version},
		})
		if err != nil {
			return nil, SecurityOutput{}, err
		}

		pkgAdvisories := advisories[input.Package]
		if pkgAdvisories == nil {
			pkgAdvisories = []npm.Advisory{}
		}

		summary := "No known vulnerabilities found."
		if len(pkgAdvisories) > 0 {
			summary = fmt.Sprintf("Found %d known vulnerabilities.", len(pkgAdvisories))
		}

		output := SecurityOutput{
			Package:         input.Package,
			Version:         version,
			Vulnerabilities: pkgAdvisories,
			Summary:         summary,
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

// SecurityTool returns the tool definition for check-package-security
func SecurityTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "check-package-security",
		Description: "Deep security audit for an NPM package, listing all known vulnerabilities",
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

func getTypeScriptSupport(ctx context.Context, client *npm.Client, pkgName string, meta npm.PackageVersion) (support, status string) {
	if meta.Types != "" || meta.Typings != "" {
		return supportIncluded, "bundled"
	}

	// Check for @types package
	// Scoped packages like @foo/bar are under @types/foo__bar
	typesPkg := "@types/" + pkgName
	if strings.HasPrefix(pkgName, "@") {
		typesPkg = "@types/" + strings.ReplaceAll(pkgName[1:], "/", "__")
	}

	if _, err := client.GetAbbreviatedPackage(ctx, typesPkg); err == nil {
		return supportIncluded, "@types available"
	}

	return statusUnknown, "none"
}

func getDownloadTrend(current, previous int) string {
	if previous <= 0 {
		return "stable"
	}

	change := float64(current-previous) / float64(previous) * 100
	if change >= 10 {
		return "growing"
	}

	if change <= -10 {
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
