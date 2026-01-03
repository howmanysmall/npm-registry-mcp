package license_test

import (
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/license"
)

func TestAssessRisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		license  string
		expected license.RiskLevel
	}{
		{"MIT", license.RiskLow},
		{"Apache-2.0", license.RiskLow},
		{"BSD-3-Clause", license.RiskLow},
		{"ISC", license.RiskLow},
		{"LGPL-3.0", license.RiskMedium},
		{"MPL-2.0", license.RiskMedium},
		{"GPL-3.0", license.RiskHigh},
		{"AGPL-3.0", license.RiskHigh},
		{"SSPL-1.0", license.RiskCritical},
		{"UNLICENSED", license.RiskCritical},
		{"", license.RiskCritical},
		{"Custom-License", license.RiskCritical},

		// Normalization edge cases
		{"mit", license.RiskLow},               // lowercase
		{"  MIT  ", license.RiskLow},           // whitespace
		{"GPL-3.0-or-later", license.RiskHigh}, // -or-later variant
	}

	for _, tt := range tests {
		t.Run(tt.license, func(t *testing.T) {
			t.Parallel()

			result := license.AssessRisk(tt.license)
			if result.Level != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.Level)
			}
		})
	}
}
