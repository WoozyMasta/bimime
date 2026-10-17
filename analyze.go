// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	// thoroughPrefixSize covers the largest global binary probe, including DX10 EDDS.
	thoroughPrefixSize = 156
	// semanticPrefixSize bounds path-scoped text pattern probes.
	semanticPrefixSize = 4096
	// maxConsecutiveEmptyReads prevents broken readers from spinning forever.
	maxConsecutiveEmptyReads = 100
)

// analysisPlan caches path evidence and all prefix budgets for one operation.
type analysisPlan struct {
	// path is the registry entry resolved from the logical name.
	path registryRecord
	// pattern is the cached target record for a path-scoped semantic verifier.
	pattern registryRecord
	// requiredPrefix is the maximum prefix read or accepted for the profile.
	requiredPrefix int
	// verifierPrefix is the path-specific bound needed to reject expected content.
	verifierPrefix int
	// hasPath reports whether the logical name matched a registered type.
	hasPath bool
	// profile selects the probes used by analysis.
	profile Profile
	// hasPattern reports whether the path has a semantic verifier.
	hasPattern bool
}

// RequiredPrefix returns the total prefix length needed for the selected name and profile.
// Fast always returns zero.
func RequiredPrefix(name string, profile Profile) int {
	return makePlan(name, profile).requiredPrefix
}

// Analyze classifies a logical name and caller-owned prefix without performing I/O.
func Analyze(name string, prefix []byte, profile Profile) Result {
	return analyzePlanned(makePlan(name, profile), prefix)
}

// makePlan resolves path evidence and the budgets needed for one classification.
func makePlan(name string, profile Profile) analysisPlan {
	validateProfile(profile)
	path, hasPath := detectByPathRecord(name)
	plan := analysisPlan{path: path, hasPath: hasPath, profile: profile}

	switch profile {
	case ProfileFast:
		return plan

	case ProfileNormal:
		if hasPath {
			if capability, ok := builtinRegistry.normalProbes[path.typ.ID]; ok {
				plan.requiredPrefix = capability.requiredPrefix
			}
		}

	case ProfileThorough:
		plan.requiredPrefix = builtinRegistry.thoroughPrefix
		if !hasPath {
			return plan
		}

		plan.verifierPrefix = builtinRegistry.verifierRequiredPrefix[path.typ.ID]
		plan.requiredPrefix = max(plan.requiredPrefix, plan.verifierPrefix)
		for _, refinement := range builtinRegistry.pathPatternRefinements {
			if refinement.scope != probeRefine || path.typ.ID != refinement.pathType {
				continue
			}

			plan.requiredPrefix = max(plan.requiredPrefix, refinement.requiredPrefix)
			plan.pattern, plan.hasPattern = builtinRegistry.byID[string(refinement.targetType)]
			break
		}
	}

	return plan
}

// analyzePlanned classifies a prefix using path evidence and budgets resolved by makePlan.
func analyzePlanned(plan analysisPlan, prefix []byte) Result {
	required := plan.requiredPrefix
	if len(prefix) > required {
		prefix = prefix[:required]
	}

	result := Result{Type: UnknownID, RequiredPrefix: required}
	if len(prefix) < required {
		result.Flags |= FlagIncomplete
	}

	pathRecord, hasPath := plan.path, plan.hasPath
	if hasPath {
		result.ByPath = pathRecord.typ.ID
		result.Type = pathRecord.typ.ID
		result.Relation = RelationPathOnly
	}

	if plan.profile == ProfileFast || len(prefix) == 0 {
		if !hasPath {
			result.Relation = RelationUnknown
		}
		return result
	}

	if plan.profile == ProfileNormal {
		if !hasPath || required == 0 {
			return result
		}

		contentRecord, hasContent := detectNormalProbeRecord(prefix, pathRecord.typ.ID)
		contentEvidenceRecord := contentRecord
		contentSource := ContentSourceMagic
		if hasContent {
			contentRecord = specializeRAPRecordForExtension(contentRecord, string(pathRecord.typ.ID))
			if contentRecord.typ.Family != pathRecord.typ.Family {
				hasContent = false
			}
		}
		if hasContent {
			result.ByContent = contentEvidenceRecord.typ.ID
			result.ContentSource = contentSource
		}

		resolveEvidence(&result, pathRecord, hasPath, contentRecord, hasContent)
		return result
	}

	contentRecord, contentSource, hasContent := detectByMagicRecord(prefix)
	if hasPath && !hasContent {
		contentRecord, contentSource, hasContent = detectVerifyRecord(prefix, pathRecord.typ.ID)
	}
	if hasPath && !hasContent {
		if plan.hasPattern && matchRecordPatterns(plan.pattern, prefix) {
			contentRecord = plan.pattern
			hasContent = true
			contentSource = ContentSourcePattern
		}
	}
	contentEvidenceRecord := contentRecord
	if hasPath && hasContent {
		contentRecord = specializeRAPRecordForExtension(contentRecord, string(pathRecord.typ.ID))
	}

	if hasContent {
		result.ByContent = contentEvidenceRecord.typ.ID
		result.ContentSource = contentSource
	}
	resolveEvidence(&result, pathRecord, hasPath, contentRecord, hasContent)

	if plan.profile == ProfileThorough && hasPath && !hasContent {
		if plan.verifierPrefix > 0 && len(prefix) >= plan.verifierPrefix {
			result.Flags |= FlagExpectedContentMissing
		}
	}

	return result
}

// Probe performs thorough in-memory classification without reading external data.
func Probe(name string, prefix []byte) Result {
	return Analyze(name, prefix, ProfileThorough)
}

// AnalyzeReader consumes at most the prefix planned for the selected profile.
// After the call, the reader remains positioned after the bytes consumed,
// so callers can continue reading the remaining data.
func AnalyzeReader(name string, reader io.Reader, profile Profile) (Result, error) {
	plan := makePlan(name, profile)
	required := plan.requiredPrefix
	if required == 0 {
		return analyzePlanned(plan, nil), nil
	}
	if reader == nil {
		return Result{}, ErrNilReader
	}

	prefix, err := readPrefix(reader, required)
	if err != nil {
		return Result{}, fmt.Errorf("read content prefix: %w", err)
	}

	return analyzePlanned(plan, prefix), nil
}

// AnalyzeFile classifies path and opens it only when the selected profile requires content.
// Path-only analysis does not verify file existence.
func AnalyzeFile(path string, profile Profile) (Result, error) {
	plan := makePlan(path, profile)
	required := plan.requiredPrefix
	if required == 0 {
		return analyzePlanned(plan, nil), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	prefix, err := readPrefix(file, required)
	if err != nil {
		return Result{}, fmt.Errorf("read content prefix: %w", err)
	}

	return analyzePlanned(plan, prefix), nil
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

// matchRecordPatterns reports whether a record's semantic verifier matches the prefix.
func matchRecordPatterns(record registryRecord, prefix []byte) bool {
	if record.contentHint != "" && !containsFoldASCII(prefix, record.contentHint) {
		return false
	}

	for _, pattern := range record.contentPatterns {
		if pattern.Match(prefix) {
			return true
		}
	}

	return false
}

// containsFoldASCII checks for an ASCII case-insensitive literal without allocation.
func containsFoldASCII(data []byte, literal string) bool {
	if literal == "" {
		return true
	}

	limit := len(data) - len(literal)
	if limit < 0 {
		return false
	}

	first := foldASCII(literal[0])
	for offset := 0; offset <= limit; {
		lower := bytes.IndexByte(data[offset:limit+1], first)
		upper := -1
		if first >= 'a' && first <= 'z' {
			upper = bytes.IndexByte(data[offset:limit+1], first-('a'-'A'))
		}
		if lower < 0 && upper < 0 {
			return false
		}

		if lower < 0 || upper >= 0 && upper < lower {
			lower = upper
		}
		offset += lower

		matched := true
		for index := 1; index < len(literal); index++ {
			if foldASCII(data[offset+index]) != foldASCII(literal[index]) {
				matched = false
				break
			}
		}

		if matched {
			return true
		}

		offset++
	}

	return false
}

// foldASCII lowercases one ASCII uppercase letter.
func foldASCII(value byte) byte {
	if value >= 'A' && value <= 'Z' {
		return value + ('a' - 'A')
	}

	return value
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
