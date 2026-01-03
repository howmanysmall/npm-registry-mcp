package health_test

import (
	"testing"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/health"
)

func TestCalculateScore(t *testing.T) {
	t.Parallel()

	// Recent, active package
	input := health.Input{
		LastPublish:       time.Now().AddDate(0, 0, -7), // 7 days ago
		WeeklyDownloads:   1000000,
		PrevWeekDownloads: 900000,
		DirectDeps:        5,
		OutdatedDeps:      0,
		CommitCount90d:    20,
		OpenIssues:        10,
		MaintainerCount:   3,
	}

	result := health.CalculateScore(input)

	if result.Score < 70 {
		t.Errorf("expected score >= 70 for healthy package, got %d", result.Score)
	}

	if result.Verdict != health.VerdictYes {
		t.Errorf("expected verdict 'yes', got %s", result.Verdict)
	}
}

func TestCalculateScore_Stale(t *testing.T) {
	t.Parallel()

	// Old, abandoned package
	input := health.Input{
		LastPublish:       time.Now().AddDate(-3, 0, 0), // 3 years ago
		WeeklyDownloads:   100,
		PrevWeekDownloads: 150,
		DirectDeps:        20,
		OutdatedDeps:      15,
		CommitCount90d:    0,
		OpenIssues:        100,
		MaintainerCount:   1,
	}

	result := health.CalculateScore(input)

	if result.Score >= 40 {
		t.Errorf("expected score < 40 for stale package, got %d", result.Score)
	}

	if result.Verdict != health.VerdictNo {
		t.Errorf("expected verdict 'no', got %s", result.Verdict)
	}
}

func TestCalculateScore_WithVulnerabilities(t *testing.T) {
	t.Parallel()

	// Package with good metrics but has vulnerabilities
	input := health.Input{
		LastPublish:       time.Now().AddDate(0, 0, -7),
		WeeklyDownloads:   1000000,
		PrevWeekDownloads: 900000,
		DirectDeps:        5,
		OutdatedDeps:      0,
		CommitCount90d:    20,
		MaintainerCount:   3,
		VulnCount:         1,
	}

	result := health.CalculateScore(input)

	if result.Verdict != health.VerdictNo {
		t.Errorf("expected verdict 'no' for package with vulnerabilities, got %s", result.Verdict)
	}
}

func TestCalculateScore_Caution(t *testing.T) {
	t.Parallel()

	// Package with moderate metrics
	input := health.Input{
		LastPublish:       time.Now().AddDate(0, -4, 0), // 4 months ago
		WeeklyDownloads:   50000,
		PrevWeekDownloads: 55000, // slight decline
		DirectDeps:        10,
		OutdatedDeps:      2,
		CommitCount90d:    5,
		MaintainerCount:   1,
	}

	result := health.CalculateScore(input)

	if result.Verdict != health.VerdictCaution {
		t.Errorf("expected verdict 'caution', got %s", result.Verdict)
	}
}
