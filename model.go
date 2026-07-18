// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// TypeID is the stable identifier used to refer to a registered file type.
type TypeID string

// FamilyID groups type variants that can refine one another.
type FamilyID string

// Trait describes an objective property of a file type.
type Trait uint32

const (
	// TraitText marks a type whose payload is text.
	TraitText Trait = 1 << iota
	// TraitBinary marks a type whose payload is binary.
	TraitBinary
	// TraitSource marks source or editable content.
	TraitSource
	// TraitCompiled marks compiled content.
	TraitCompiled
	// TraitContainer marks a container format.
	TraitContainer
	// TraitArchive marks an archive format.
	TraitArchive
	// TraitImage marks an image format.
	TraitImage
	// TraitAudio marks an audio format.
	TraitAudio
	// TraitVideo marks a video format.
	TraitVideo
)

// Type describes one known file type from the built-in registry.
type Type struct {
	// ID is the stable identifier for this type.
	ID TypeID `json:"id" yaml:"id"`
	// Family identifies compatible generic and specialized classifications.
	Family FamilyID `json:"family,omitempty" yaml:"family,omitempty"`
	// MIME is the canonical media type.
	MIME string `json:"mime" yaml:"mime"`
	// Description is the detailed human-readable type description.
	Description string `json:"description" yaml:"description"`
	// ShortDescription is a compact description for narrow UI columns.
	ShortDescription string `json:"short_description,omitempty" yaml:"short_description,omitempty"`
	// Extensions lists mapped extensions without a leading dot.
	Extensions []string `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	// Traits contains neutral, objective properties of the type.
	Traits Trait `json:"traits" yaml:"traits"`
}

// Has reports whether the type has the requested trait.
func (t Type) Has(trait Trait) bool {
	return t.Traits&trait != 0
}

// UnknownID identifies the fallback type used when classification finds no match.
const UnknownID TypeID = "unknown"

// Profile controls the cost and evidence budget for classification.
type Profile uint8

const (
	// ProfileFast uses path evidence only and never plans content reads.
	ProfileFast Profile = iota
	// ProfileNormal reads content only for useful path-specific refinements.
	ProfileNormal
	// ProfileThorough enables bounded global discovery and path-specific probes.
	ProfileThorough
)

// Relation describes how path and content evidence relate.
type Relation uint8

const (
	// RelationUnknown means neither path nor content produced a match.
	RelationUnknown Relation = iota
	// RelationPathOnly means only path evidence matched.
	RelationPathOnly
	// RelationContentOnly means only content evidence matched.
	RelationContentOnly
	// RelationConfirmed means path and content identify the same type.
	RelationConfirmed
	// RelationRefined means one compatible signal identifies a more specific type.
	RelationRefined
	// RelationConflict means path and content identify incompatible types.
	RelationConflict
)

// ContentSource identifies the kind of content probe that produced evidence.
type ContentSource uint8

const (
	// ContentSourceNone means no content probe produced a match.
	ContentSourceNone ContentSource = iota
	// ContentSourceMagic means a static signature produced the match.
	ContentSourceMagic
	// ContentSourceStructure means a structured binary probe produced the match.
	ContentSourceStructure
	// ContentSourcePattern means a path-scoped content pattern produced the match.
	ContentSourcePattern
)

// ResultFlag records bounded-probe conditions that affect interpretation.
type ResultFlag uint16

const (
	// FlagIncomplete means the supplied prefix was shorter than the planned probe.
	FlagIncomplete ResultFlag = 1 << iota
	// FlagExpectedContentMissing means a complete expected verifier did not match.
	FlagExpectedContentMissing
)

// Result contains a classification and the evidence used to resolve it.
type Result struct {
	// Type is the final selected type, or UnknownID when no type is known.
	Type TypeID `json:"type" yaml:"type"`
	// ByPath is the type suggested by the logical name, if any.
	ByPath TypeID `json:"by_path,omitempty" yaml:"by_path,omitempty"`
	// ByContent is the type suggested by the supplied prefix, if any.
	ByContent TypeID `json:"by_content,omitempty" yaml:"by_content,omitempty"`
	// Relation describes compatibility between path and content evidence.
	Relation Relation `json:"relation" yaml:"relation"`
	// ContentSource identifies the probe that produced ByContent.
	ContentSource ContentSource `json:"content_source" yaml:"content_source"`
	// Flags records incomplete or missing expected content evidence.
	Flags ResultFlag `json:"flags,omitempty" yaml:"flags,omitempty"`
	// RequiredPrefix is the total prefix length planned for the name and profile.
	RequiredPrefix int `json:"required_prefix" yaml:"required_prefix"`
}

// Has reports whether the result contains the requested flag.
func (r Result) Has(flag ResultFlag) bool {
	return r.Flags&flag != 0
}

// UnknownType returns the fallback descriptor for unclassified content.
func UnknownType() Type {
	return Type{
		ID:          UnknownID,
		MIME:        "application/octet-stream",
		Description: "Unknown or generic binary payload",
		Traits:      TraitBinary,
	}
}
