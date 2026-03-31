// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// BIAmbiguousRAPExtensions is the default extension set for fast mode with
// forced magic probing for BI formats that can appear as source or binarized.
var BIAmbiguousRAPExtensions = []string{
	"p3d",
	"wrp",
	"rvmat",
	"bisurf",
}

// BIAmbiguousRAPOverrides returns extension overrides for the common
// "fast + forced magic for ambiguous BI formats" scenario.
func BIAmbiguousRAPOverrides() map[string]AnalyzePlan {
	plans := make(map[string]AnalyzePlan, len(BIAmbiguousRAPExtensions))
	for _, ext := range BIAmbiguousRAPExtensions {
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
