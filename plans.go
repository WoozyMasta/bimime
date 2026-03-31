// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// PlanFast returns extension-only matching with no strict validation.
func PlanFast() AnalyzePlan {
	return AnalyzePlan{
		Match:    AnalyzeMatchExtension,
		Validate: AnalyzeValidateNone,
	}
}

// PlanNormal returns extension+magic-as-needed matching with no strict checks.
func PlanNormal() AnalyzePlan {
	return AnalyzePlan{
		Match:    AnalyzeMatchExtensionMagicNeeded,
		Validate: AnalyzeValidateNone,
	}
}

// PlanStrict returns extension+magic-as-needed matching with strict checks.
func PlanStrict() AnalyzePlan {
	return AnalyzePlan{
		Match:    AnalyzeMatchExtensionMagicNeeded,
		Validate: AnalyzeValidateStrict,
	}
}
