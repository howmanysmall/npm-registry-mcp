// Package license provides license risk assessment functionality.
package license

import "strings"

// RiskLevel represents the risk level of a license
type RiskLevel string

// Risk level constants for license assessment.
const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// RiskAssessment is the result of assessing a license
type RiskAssessment struct {
	License     string    `json:"license"`
	Level       RiskLevel `json:"level"`
	Description string    `json:"description"`
	Flags       []string  `json:"flags,omitempty"`
}

// Permissive licenses (low risk)
var permissiveLicenses = map[string]bool{
	"MIT":          true,
	"Apache-2.0":   true,
	"BSD-2-Clause": true,
	"BSD-3-Clause": true,
	"ISC":          true,
	"0BSD":         true,
	"Unlicense":    true,
	"WTFPL":        true,
	"CC0-1.0":      true,
	"Zlib":         true,
	"BSL-1.0":      true, // Boost Software License
}

// Weak copyleft licenses (medium risk)
var weakCopyleftLicenses = map[string]bool{
	"LGPL-2.0": true,
	"LGPL-2.1": true,
	"LGPL-3.0": true,
	"MPL-2.0":  true,
	"EPL-1.0":  true,
	"EPL-2.0":  true,
	"CDDL-1.0": true,
	"CDDL-1.1": true,
}

// Strong copyleft licenses (high risk)
var strongCopyleftLicenses = map[string]bool{
	"GPL-2.0":       true,
	"GPL-3.0":       true,
	"AGPL-3.0":      true,
	"GPL-2.0-only":  true,
	"GPL-3.0-only":  true,
	"AGPL-3.0-only": true,
}

// Problematic licenses (critical risk)
var problematicLicenses = map[string]bool{
	"SSPL-1.0":       true,
	"BUSL-1.1":       true, // Business Source License
	"Commons-Clause": true,
	"UNLICENSED":     true,
}

// AssessRisk assesses the risk level of a license
func AssessRisk(spdxID string) RiskAssessment {
	normalized := normalizeLicense(spdxID)

	// Check each category
	if isPermissive(normalized) {
		return RiskAssessment{
			License:     spdxID,
			Level:       RiskLow,
			Description: "Permissive license - generally safe for any use",
		}
	}

	if isWeakCopyleft(normalized) {
		return RiskAssessment{
			License:     spdxID,
			Level:       RiskMedium,
			Description: "Weak copyleft - some restrictions on modifications",
			Flags:       []string{"Modifications may need to be shared under same license"},
		}
	}

	if isStrongCopyleft(normalized) {
		flags := []string{
			"Copyleft - derivative works must use same license",
			"May be incompatible with proprietary projects",
		}
		if strings.Contains(normalized, "AGPL") {
			flags = append(flags, "Network use triggers license requirements")
		}

		return RiskAssessment{
			License:     spdxID,
			Level:       RiskHigh,
			Description: "Strong copyleft - significant restrictions",
			Flags:       flags,
		}
	}

	if isProblematic(normalized) {
		return RiskAssessment{
			License:     spdxID,
			Level:       RiskCritical,
			Description: "Problematic license - review with legal",
			Flags:       []string{"May have commercial use restrictions", "Not OSI approved"},
		}
	}

	// Unknown or missing license
	if spdxID == "" {
		return RiskAssessment{
			License:     "(none)",
			Level:       RiskCritical,
			Description: "No license specified",
			Flags:       []string{"All rights reserved by default", "Cannot use without explicit permission"},
		}
	}

	return RiskAssessment{
		License:     spdxID,
		Level:       RiskCritical,
		Description: "Unknown license - review manually",
		Flags:       []string{"License not recognized", "Review terms before use"},
	}
}

func normalizeLicense(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	// Handle common variations
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	return s
}

func isPermissive(normalized string) bool {
	for license := range permissiveLicenses {
		if strings.EqualFold(normalized, license) {
			return true
		}
	}

	return false
}

func isWeakCopyleft(normalized string) bool {
	for license := range weakCopyleftLicenses {
		if strings.EqualFold(normalized, license) {
			return true
		}

		// Handle -or-later variants
		if strings.EqualFold(normalized, license+"-OR-LATER") {
			return true
		}
	}

	return false
}

func isStrongCopyleft(normalized string) bool {
	for license := range strongCopyleftLicenses {
		if strings.EqualFold(normalized, license) {
			return true
		}

		// Handle -or-later variants
		if strings.EqualFold(normalized, license+"-OR-LATER") {
			return true
		}
	}

	return false
}

func isProblematic(normalized string) bool {
	for license := range problematicLicenses {
		if strings.EqualFold(normalized, license) {
			return true
		}
	}

	return false
}
