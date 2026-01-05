// Package health provides health scoring for NPM packages based on maintenance,
// activity, and other quality signals.
package health

import (
	"fmt"
	"time"
)

// Verdict is the overall recommendation
type Verdict string

// Verdict constants represent the possible recommendations.
const (
	VerdictYes     Verdict = "yes"
	VerdictCaution Verdict = "caution"
	VerdictNo      Verdict = "no"
)

// Input contains data for health scoring
type Input struct {
	LastPublish       time.Time
	WeeklyDownloads   int
	PrevWeekDownloads int
	DirectDeps        int
	TransitiveDeps    int
	OutdatedDeps      int
	CommitCount90d    int
	OpenIssues        int
	MaintainerCount   int
	HasTypes          bool
	UnpackedSize      int64
	VulnCount         int
}

// Factor is a single scoring factor
type Factor struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Score  int    `json:"score"`
	Status string `json:"status"`
}

// Result is the health assessment result
type Result struct {
	Score    int      `json:"score"`
	Verdict  Verdict  `json:"verdict"`
	Factors  []Factor `json:"factors"`
	Warnings []string `json:"warnings"`
}

func determineVerdict(
	score int,
	warnings int,
	vulnCount int,
) Verdict {
	if score < 40 || vulnCount > 0 {
		return VerdictNo
	}

	if score < 70 || warnings > 0 {
		return VerdictCaution
	}

	return VerdictYes
}

func computeFactors(input Input) (
	factors []Factor,
	warnings []string,
	totalScore int,
	totalWeight int,
) {
	var (
		s int
		f Factor
	)

	s, f = scoreLastPublish(input.LastPublish)
	factors = append(factors, f)
	totalScore += s * 25
	totalWeight += 25

	if s < 50 {
		warnings = append(warnings, fmt.Sprintf("Last published %s", f.Value))
	}

	s, f = scoreDownloadTrend(input.WeeklyDownloads, input.PrevWeekDownloads)
	factors = append(factors, f)
	totalScore += s * 20
	totalWeight += 20

	s, f = scoreOutdatedDeps(input.DirectDeps, input.OutdatedDeps)
	factors = append(factors, f)
	totalScore += s * 20
	totalWeight += 20

	if s < 50 {
		warnings = append(warnings, fmt.Sprintf("%d outdated dependencies", input.OutdatedDeps))
	}

	s, f = scoreCommitActivity(input.CommitCount90d)
	factors = append(factors, f)
	totalScore += s * 15
	totalWeight += 15

	s, f = scoreMaintainers(input.MaintainerCount)
	factors = append(factors, f)
	totalScore += s * 10
	totalWeight += 10

	if input.MaintainerCount == 1 {
		warnings = append(warnings, "Bus factor: 1")
	}

	s, f = scoreVulnerabilities(input.VulnCount)
	factors = append(factors, f)
	totalScore += s * 10
	totalWeight += 10

	if input.VulnCount > 0 {
		warnings = append(warnings, fmt.Sprintf("%d known vulnerabilities", input.VulnCount))
	}

	return factors, warnings, totalScore, totalWeight
}

// CalculateScore calculates the health score
func CalculateScore(input Input) Result {
	factors, warnings, totalScore, totalWeight := computeFactors(input)

	finalScore := 0
	if totalWeight > 0 {
		finalScore = totalScore / totalWeight
	}

	verdict := determineVerdict(finalScore, len(warnings), input.VulnCount)

	return Result{
		Score:    finalScore,
		Verdict:  verdict,
		Factors:  factors,
		Warnings: warnings,
	}
}

func scoreLastPublish(t time.Time) (int, Factor) {
	days := int(time.Since(t).Hours() / 24)

	var score int

	var status string

	switch {
	case days <= 30:
		score = 100
		status = "active"
	case days <= 90:
		score = 80
		status = "recent"
	case days <= 180:
		score = 60
		status = "aging"
	case days <= 365:
		score = 40
		status = "stale"
	case days <= 730:
		score = 20
		status = "very stale"
	default:
		score = 0
		status = "abandoned"
	}

	value := formatDuration(days)

	return score, Factor{
		Name:   "Last Publish",
		Value:  value,
		Score:  score,
		Status: status,
	}
}

func scoreDownloadTrend(current, previous int) (int, Factor) {
	if previous == 0 {
		return 50, Factor{
			Name:   "Download Trend",
			Value:  "no history",
			Score:  50,
			Status: "unknown",
		}
	}

	change := float64(current-previous) / float64(previous) * 100

	var score int

	var status string

	switch {
	case change >= 10:
		score = 100
		status = "growing"
	case change >= 0:
		score = 80
		status = "stable"
	case change >= -10:
		score = 60
		status = "slight decline"
	case change >= -25:
		score = 40
		status = "declining"
	default:
		score = 20
		status = "steep decline"
	}

	return score, Factor{
		Name:   "Download Trend",
		Value:  fmt.Sprintf("%.1f%%", change),
		Score:  score,
		Status: status,
	}
}

func scoreOutdatedDeps(total, outdated int) (int, Factor) {
	if total == 0 {
		return 100, Factor{
			Name:   "Dependencies",
			Value:  "none",
			Score:  100,
			Status: "clean",
		}
	}

	pct := float64(outdated) / float64(total) * 100

	var score int

	var status string

	switch {
	case pct == 0:
		score = 100
		status = "up to date"
	case pct <= 10:
		score = 80
		status = "mostly current"
	case pct <= 25:
		score = 60
		status = "some outdated"
	case pct <= 50:
		score = 40
		status = "many outdated"
	default:
		score = 20
		status = "severely outdated"
	}

	return score, Factor{
		Name:   "Dependencies",
		Value:  fmt.Sprintf("%d/%d outdated", outdated, total),
		Score:  score,
		Status: status,
	}
}

func scoreCommitActivity(count int) (int, Factor) {
	var score int

	var status string

	switch {
	case count >= 20:
		score = 100
		status = "very active"
	case count >= 10:
		score = 80
		status = "active"
	case count >= 5:
		score = 60
		status = "moderate"
	case count >= 1:
		score = 40
		status = "low"
	default:
		score = 20
		status = "inactive"
	}

	return score, Factor{
		Name:   "Commit Activity (90d)",
		Value:  fmt.Sprintf("%d commits", count),
		Score:  score,
		Status: status,
	}
}

func scoreMaintainers(count int) (int, Factor) {
	var score int

	var status string

	switch {
	case count >= 3:
		score = 100
		status = "healthy team"
	case count == 2:
		score = 70
		status = "small team"
	case count == 1:
		score = 40
		status = "single maintainer"
	default:
		score = 0
		status = "no maintainers"
	}

	return score, Factor{
		Name:   "Maintainers",
		Value:  fmt.Sprintf("%d", count),
		Score:  score,
		Status: status,
	}
}

func scoreVulnerabilities(count int) (int, Factor) {
	var score int

	var status string

	switch {
	case count == 0:
		score = 100
		status = "none known"
	case count <= 2:
		score = 40
		status = "some issues"
	default:
		score = 0
		status = "critical"
	}

	return score, Factor{
		Name:   "Vulnerabilities",
		Value:  fmt.Sprintf("%d", count),
		Score:  score,
		Status: status,
	}
}

func formatDuration(days int) string {
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
