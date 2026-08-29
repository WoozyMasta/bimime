// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"fmt"
	"io"
	"os"
	"slices"
)

const (
	// thoroughPrefixSize covers the largest global binary probe, including DX10 EDDS.
	thoroughPrefixSize = 156
	// semanticPrefixSize bounds path-scoped text pattern probes.
	semanticPrefixSize = 4096
	// maxConsecutiveEmptyReads prevents broken readers from spinning forever.
	maxConsecutiveEmptyReads = 100
)

// RequiredPrefix returns the total prefix length needed for the selected name and profile.
// Fast always returns zero.
func RequiredPrefix(name string, profile Profile) int {
	validateProfile(profile)

	switch profile {
	case ProfileFast:
		return 0

	case ProfileNormal:
		if record, ok := detectByPathRecord(name); ok {
			if capability, ok := builtinRegistry.normalProbes[record.typ.ID]; ok {
				return capability.requiredPrefix
			}
		}
		return 0

	case ProfileThorough:
		required := builtinRegistry.thoroughPrefix
		if record, ok := detectByPathRecord(name); ok {
			for _, refinement := range pathPatternRefinements {
				if record.typ.ID == refinement.pathType {
					required = max(required, refinement.requiredPrefix)
					break
				}
			}
		}
		return required

	default:
		panic("unreachable profile")
	}
}

// Analyze classifies a logical name and caller-owned prefix without performing I/O.
func Analyze(name string, prefix []byte, profile Profile) Result {
	required := RequiredPrefix(name, profile)
	if len(prefix) > required {
		prefix = prefix[:required]
	}

	result := Result{Type: UnknownID, RequiredPrefix: required}
	if len(prefix) < required {
		result.Flags |= FlagIncomplete
	}

	pathRecord, hasPath := detectByPathRecord(name)
	if hasPath {
		result.ByPath = pathRecord.typ.ID
		result.Type = pathRecord.typ.ID
		result.Relation = RelationPathOnly
	}

	if profile == ProfileFast || len(prefix) == 0 {
		if !hasPath {
			result.Relation = RelationUnknown
		}
		return result
	}

	if profile == ProfileNormal {
		if !hasPath || required == 0 {
			return result
		}

		contentRecord, hasContent := detectNormalProbeRecord(prefix, pathRecord.typ.ID)
		contentSource := ContentSourceMagic
		if hasContent {
			contentRecord = specializeRAPRecordForExtension(contentRecord, string(pathRecord.typ.ID))
			if contentRecord.typ.Family != pathRecord.typ.Family {
				hasContent = false
			}
		}
		if hasContent {
			result.ByContent = contentRecord.typ.ID
			result.ContentSource = contentSource
		}

		resolveEvidence(&result, pathRecord, hasPath, contentRecord, hasContent)
		return result
	}

	contentRecord, contentSource, hasContent := detectByMagicRecord(prefix)
	if hasPath && !hasContent {
		if refined, ok := matchPathPattern(pathRecord, prefix); ok {
			contentRecord = refined
			hasContent = true
			contentSource = ContentSourcePattern
		}
	}
	if hasPath && hasContent {
		contentRecord = specializeRAPRecordForExtension(contentRecord, string(pathRecord.typ.ID))
	}

	if hasContent {
		result.ByContent = contentRecord.typ.ID
		result.ContentSource = contentSource
	}
	resolveEvidence(&result, pathRecord, hasPath, contentRecord, hasContent)

	if profile == ProfileThorough && hasPath && !hasContent && len(prefix) >= required &&
		hasContentVerifier(pathRecord) {
		result.Flags |= FlagExpectedContentMissing
	}

	return result
}

// Probe performs thorough in-memory classification without reading external data.
func Probe(name string, prefix []byte) Result {
	return Analyze(name, prefix, ProfileThorough)
}

// AnalyzeReader reads only the prefix planned for the selected profile.
func AnalyzeReader(name string, reader io.Reader, profile Profile) (Result, error) {
	required := RequiredPrefix(name, profile)
	if required == 0 {
		return Analyze(name, nil, profile), nil
	}
	if reader == nil {
		return Result{}, ErrNilReader
	}

	prefix, err := readPrefix(reader, required)
	if err != nil {
		return Result{}, fmt.Errorf("read content prefix: %w", err)
	}

	return Analyze(name, prefix, profile), nil
}

// AnalyzeFile opens the exact supplied path and reads only its planned prefix.
func AnalyzeFile(path string, profile Profile) (Result, error) {
	required := RequiredPrefix(path, profile)
	if required == 0 {
		return Analyze(path, nil, profile), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return AnalyzeReader(path, file, profile)
}

// validateProfile rejects enum values that are outside the public profile set.
func validateProfile(profile Profile) {
	if profile > ProfileThorough {
		panic(fmt.Sprintf("bimime: invalid profile %d", profile))
	}
}

// readPrefix allocates the planned bound once and detects readers that stop making progress.
func readPrefix(reader io.Reader, limit int) ([]byte, error) {
	buf := make([]byte, limit)
	read := 0
	emptyReads := 0

	for read < limit {
		n, err := reader.Read(buf[read:])
		read += n
		if err != nil {
			if err == io.EOF {
				return buf[:read], nil
			}

			return nil, err
		}

		if n == 0 {
			emptyReads++
			if emptyReads >= maxConsecutiveEmptyReads {
				return nil, io.ErrNoProgress
			}

			continue
		}

		emptyReads = 0
	}

	return buf, nil
}

// matchPathPattern returns a registered semantic refinement for a path type.
func matchPathPattern(pathRecord registryRecord, prefix []byte) (registryRecord, bool) {
	for _, refinement := range pathPatternRefinements {
		if pathRecord.typ.ID != refinement.pathType ||
			!matchContentPatternForType(string(refinement.targetType), prefix) {
			continue
		}

		record, ok := typeByID[string(refinement.targetType)]
		if ok {
			return record, true
		}
	}

	return registryRecord{}, false
}

// resolveEvidence selects the final type and relation from path/content evidence.
func resolveEvidence(
	result *Result,
	pathRecord registryRecord,
	hasPath bool,
	contentRecord registryRecord,
	hasContent bool,
) {
	switch {
	case !hasPath && !hasContent:
		result.Type = UnknownID
		result.Relation = RelationUnknown

	case hasPath && !hasContent:
		result.Type = pathRecord.typ.ID
		result.Relation = RelationPathOnly

	case !hasPath && hasContent:
		result.Type = contentRecord.typ.ID
		result.Relation = RelationContentOnly

	case pathRecord.typ.ID == contentRecord.typ.ID:
		result.Type = pathRecord.typ.ID
		result.Relation = RelationConfirmed

	case pathRecord.typ.Family == contentRecord.typ.Family &&
		pathRecord.specificity != contentRecord.specificity:
		if pathRecord.specificity > contentRecord.specificity {
			result.Type = pathRecord.typ.ID
		} else {
			result.Type = contentRecord.typ.ID
		}
		result.Relation = RelationRefined

	case pathRecord.typ.Family == contentRecord.typ.Family:
		result.Type = contentRecord.typ.ID
		result.Relation = RelationConflict

	default:
		result.Type = contentRecord.typ.ID
		result.Relation = RelationConflict
	}
}

// hasContentVerifier reports whether a path type has a registered content probe.
func hasContentVerifier(pathRecord registryRecord) bool {
	if len(pathRecord.magic) > 0 {
		return true
	}

	for key := range contextualRefinements {
		if key.pathType == pathRecord.typ.ID {
			return true
		}
	}
	for _, record := range registryRecords {
		if record.typ.Family == pathRecord.typ.Family && len(record.magic) > 0 {
			return true
		}
	}
	for _, capability := range structuredProbeCapabilities {
		if slices.Contains(capability.typeIDs, pathRecord.typ.ID) {
			return true
		}
	}

	return false
}
