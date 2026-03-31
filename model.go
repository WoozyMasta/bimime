// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// Source describes how detection result was obtained.
type Source string

// AnalyzeMatchMode controls how detection match is performed.
type AnalyzeMatchMode uint8

// AnalyzeValidateMode controls whether strict validations are applied.
type AnalyzeValidateMode uint8

// AnalyzeIssue describes a strict-mode validation problem.
type AnalyzeIssue string

// Detection sources.
const (
	// SourceUnknown means no known extension/magic was matched.
	SourceUnknown Source = "unknown"
	// SourceExtension means only path hint matched (well-known filename or extension).
	SourceExtension Source = "extension"
	// SourceMagic means only magic bytes matched.
	SourceMagic Source = "magic"
	// SourceMagicAndExtension means both extension and magic matched.
	SourceMagicAndExtension Source = "magic+extension"
)

// Analyze match modes.
const (
	// AnalyzeMatchDefault falls back to extension+magic-as-needed behavior.
	AnalyzeMatchDefault AnalyzeMatchMode = iota
	// AnalyzeMatchExtension resolves only by path hint (filename/extension).
	AnalyzeMatchExtension
	// AnalyzeMatchExtensionMagicNeeded resolves by extension+magic, reading
	// content only when needed by extension heuristics.
	AnalyzeMatchExtensionMagicNeeded
	// AnalyzeMatchExtensionMagic resolves by extension+magic and expects magic
	// probing to be available for selected targets.
	AnalyzeMatchExtensionMagic
)

// Analyze validation modes.
const (
	// AnalyzeValidateDefault falls back to no strict validation.
	AnalyzeValidateDefault AnalyzeValidateMode = iota
	// AnalyzeValidateNone disables strict validation checks.
	AnalyzeValidateNone
	// AnalyzeValidateStrict enables strict consistency and text checks.
	AnalyzeValidateStrict
)

// Strict-mode issues.
const (
	// AnalyzeIssueMagicMismatch means extension/path hint conflicts with magic.
	AnalyzeIssueMagicMismatch AnalyzeIssue = "magic_mismatch"
	// AnalyzeIssueTextExpected means detected text payload looks binary.
	AnalyzeIssueTextExpected AnalyzeIssue = "text_expected"
	// AnalyzeIssueContentPatternMismatch means payload does not match expected
	// content markers for resolved type in strict mode.
	AnalyzeIssueContentPatternMismatch AnalyzeIssue = "content_pattern_mismatch"
)

// Type describes one known file type from registry.
type Type struct {
	// ID is stable internal identifier (e.g. "bi.rap").
	ID string `json:"id" yaml:"id"`
	// MIME is canonical MIME value for the type.
	MIME string `json:"mime" yaml:"mime"`
	// Description is detailed human-readable type description.
	Description string `json:"description" yaml:"description"`
	// ShortDescription is compact description for UIs with narrow columns.
	ShortDescription string `json:"short_description,omitempty" yaml:"short_description,omitempty"`
	// Extensions lists mapped extensions without a leading dot.
	Extensions []string `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	// Binary reports whether payload should be treated as binary by default.
	Binary bool `json:"binary" yaml:"binary"`
}

// ProbeResult contains extension and magic matches with resolved final type.
type ProbeResult struct {
	// Source shows which signals produced the result.
	Source Source `json:"source" yaml:"source"`
	// Extension is normalized extension used for extension lookup.
	Extension string `json:"extension,omitempty" yaml:"extension,omitempty"`
	// Resolved is final selected type (magic has priority over extension).
	Resolved Type `json:"resolved" yaml:"resolved"`
	// ByMagic is magic-based type when matched.
	ByMagic Type `json:"by_magic" yaml:"by_magic"`
	// ByExtension is path-hint type when matched (well-known filename or extension).
	ByExtension Type `json:"by_extension" yaml:"by_extension"`
}

// AnalyzePlan describes how one file should be matched and validated.
type AnalyzePlan struct {
	// Match controls extension-only vs extension+magic probing behavior.
	Match AnalyzeMatchMode `json:"match,omitempty" yaml:"match,omitempty"`
	// Validate controls whether strict validation checks are applied.
	Validate AnalyzeValidateMode `json:"validate,omitempty" yaml:"validate,omitempty"`
}

// AnalyzeOptions controls Analyze/AnalyzeReader/AnalyzeFile behavior.
type AnalyzeOptions struct {
	// PlansByExtension maps extension (without dot) to per-extension plan.
	PlansByExtension map[string]AnalyzePlan `json:"plans_by_extension,omitempty" yaml:"plans_by_extension,omitempty"`
	// Path is filesystem path or filename hint used for extension matching.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Prefix is optional payload prefix already available to caller.
	Prefix []byte `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	// PrefixSize limits bytes read from reader/file for magic and text checks.
	// Zero or negative value uses the package default.
	PrefixSize int `json:"prefix_size,omitempty" yaml:"prefix_size,omitempty"`
	// DefaultPlan is used when no extension-specific override is configured.
	DefaultPlan AnalyzePlan `json:"default_plan" yaml:"default_plan"`
}

// AnalyzeResult stores classification and validation outcome.
type AnalyzeResult struct {
	// Issues contains strict-mode validation issues.
	Issues []AnalyzeIssue `json:"issues,omitempty" yaml:"issues,omitempty"`
	// Probe contains extension/magic matches and resolved final type.
	Probe ProbeResult `json:"probe" yaml:"probe"`
	// Plan is effective plan after option normalization and extension overrides.
	Plan AnalyzePlan `json:"plan" yaml:"plan"`
	// Valid is true when strict validation passed or was not requested.
	Valid bool `json:"valid" yaml:"valid"`
	// CheckedMagic reports whether strict mode validated extension against magic.
	CheckedMagic bool `json:"checked_magic,omitempty" yaml:"checked_magic,omitempty"`
	// CheckedText reports whether strict mode validated text-like payload.
	CheckedText bool `json:"checked_text,omitempty" yaml:"checked_text,omitempty"`
	// CheckedContentPattern reports whether strict mode validated type-specific
	// content regex markers.
	CheckedContentPattern bool `json:"checked_content_pattern,omitempty" yaml:"checked_content_pattern,omitempty"`
	// LooksText reports quick text-likeness heuristic result when CheckedText is true.
	LooksText bool `json:"looks_text,omitempty" yaml:"looks_text,omitempty"`
}

// UnknownType is returned when registry has no match for extension/magic.
var UnknownType = Type{
	ID:          "unknown",
	MIME:        "application/octet-stream",
	Description: "Unknown or generic binary payload",
	Binary:      true,
}
