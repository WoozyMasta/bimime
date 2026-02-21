// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// Source describes how detection result was obtained.
type Source string

// DetectMode controls how aggressively payload content is inspected.
type DetectMode string

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

// Detection modes.
const (
	// DetectModeFast resolves only by path hint (filename/extension).
	DetectModeFast DetectMode = "fast"
	// DetectModeNormal resolves by path hint and magic when content is required.
	DetectModeNormal DetectMode = "normal"
	// DetectModeStrict resolves by path hint and magic, then validates consistency.
	DetectModeStrict DetectMode = "strict"
)

// Strict-mode issues.
const (
	// AnalyzeIssueMagicMismatch means extension/path hint conflicts with magic.
	AnalyzeIssueMagicMismatch AnalyzeIssue = "magic_mismatch"
	// AnalyzeIssueTextExpected means detected text payload looks binary.
	AnalyzeIssueTextExpected AnalyzeIssue = "text_expected"
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

// AnalyzeOptions controls Analyze/AnalyzeReader/AnalyzeFile behavior.
type AnalyzeOptions struct {
	// Mode selects detection mode. Empty value defaults to DetectModeNormal.
	Mode DetectMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	// PrefixSize limits bytes read from reader/file for magic and text checks.
	// Zero or negative value uses the package default.
	PrefixSize int `json:"prefix_size,omitempty" yaml:"prefix_size,omitempty"`
}

// AnalyzeResult stores classification and validation outcome.
type AnalyzeResult struct {
	// Mode is effective mode after option normalization.
	Mode DetectMode `json:"mode" yaml:"mode"`
	// Issues contains strict-mode validation issues.
	Issues []AnalyzeIssue `json:"issues,omitempty" yaml:"issues,omitempty"`
	// Probe contains extension/magic matches and resolved final type.
	Probe ProbeResult `json:"probe" yaml:"probe"`
	// Valid is true when strict validation passed or was not requested.
	Valid bool `json:"valid" yaml:"valid"`
	// CheckedMagic reports whether strict mode validated extension against magic.
	CheckedMagic bool `json:"checked_magic,omitempty" yaml:"checked_magic,omitempty"`
	// CheckedText reports whether strict mode validated text-like payload.
	CheckedText bool `json:"checked_text,omitempty" yaml:"checked_text,omitempty"`
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
