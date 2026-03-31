// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

var defaultBIAmbiguousRAPExtensions = []string{
	"p3d",
	"wrp",
	"rvmat",
	"bisurf",
}

// BIAmbiguousRAPExtensions returns default extension set for fast mode with
// forced magic probing for BI formats that can appear as source or binarized.
func BIAmbiguousRAPExtensions() []string {
	out := make([]string, len(defaultBIAmbiguousRAPExtensions))
	copy(out, defaultBIAmbiguousRAPExtensions)

	return out
}

// BIAmbiguousRAPOverrides returns extension overrides for the common
// "fast + forced magic for ambiguous BI formats" scenario.
func BIAmbiguousRAPOverrides() map[string]AnalyzePlan {
	extensions := BIAmbiguousRAPExtensions()
	plans := make(map[string]AnalyzePlan, len(extensions))
	for _, ext := range extensions {
		plans[ext] = AnalyzePlan{
			Match:    AnalyzeMatchExtensionMagic,
			Validate: AnalyzeValidateNone,
		}
	}

	return plans
}

// BIAmbiguousRAPOptions builds Analyze options for the common scenario:
// fast by default and forced magic probing for p3d/wrp/rvmat/bisurf.
func BIAmbiguousRAPOptions(path string, prefix []byte) AnalyzeOptions {
	return AnalyzeOptions{
		Path:             path,
		Prefix:           prefix,
		DefaultPlan:      PlanFast(),
		PlansByExtension: BIAmbiguousRAPOverrides(),
	}
}
