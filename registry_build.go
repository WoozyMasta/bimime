// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"fmt"
	"slices"
	"strings"
)

// compiledRegistry stores normalized declarations and their immutable indexes.
type compiledRegistry struct {
	// records contains normalized registry entries.
	records []registryRecord
	// byID indexes entries by normalized stable id.
	byID map[string]registryRecord
	// byExtension indexes entries by normalized extension.
	byExtension map[string]registryRecord
	// magic contains static signature matches in deterministic order.
	magic []magicMatch
	// normalProbes contains path-scoped refinements enabled by ProfileNormal.
	normalProbes map[TypeID]normalProbeCapability
	// contextualRefinements maps path and content evidence to a refined type.
	contextualRefinements map[contextualRefinementKey]TypeID
	// pathPatternRefinements contains compiled path-scoped semantic probes.
	pathPatternRefinements []pathPatternRefinement
	// structuredProbeCapabilities contains compiled non-static detectors.
	structuredProbeCapabilities []structuredProbeCapability
	// verifierRequiredPrefix stores the largest relevant content probe bound per path type.
	verifierRequiredPrefix map[TypeID]int
	// fileNames maps normalized exact basenames to type ids.
	fileNames map[string]string
	// fileRules stores normalized filename prefix and suffix rules by specificity.
	fileRules []fileRule
	// thoroughPrefix is the maximum global prefix required by content probes.
	thoroughPrefix int
}

// fileRule matches a filename prefix, suffix, or both.
type fileRule struct {
	// prefix is the required filename prefix, when set.
	prefix string
	// suffix is the required filename suffix, when set.
	suffix string
	// typeID is the stable id selected by this rule.
	typeID string
}

// mustBuildRegistry panics when the built-in registry contains invalid data.
func mustBuildRegistry(
	definitions []registryRecord,
	fileNames map[string]string,
	fileRules []fileRule,
) compiledRegistry {
	registry, err := buildRegistry(
		definitions,
		fileNames,
		fileRules,
		builtinContextualRefinements,
		builtinNormalProbeCapabilities,
		builtinPathPatternRefinements,
		builtinStructuredProbeCapabilities,
	)
	if err != nil {
		panic(fmt.Errorf("bimime: invalid built-in registry: %w", err))
	}
	return registry
}

// registryErrorf adds the registry error category to a validation message.
func registryErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errInvalidRegistry, fmt.Sprintf(format, args...))
}

// buildRegistry normalizes declarations, validates references, and compiles indexes.
func buildRegistry(
	definitions []registryRecord,
	fileNames map[string]string,
	fileRules []fileRule,
	contextual map[contextualRefinementKey]TypeID,
	normalProbes []normalProbeCapability,
	patterns []pathPatternRefinement,
	structuredProbes []structuredProbeCapability,
) (compiledRegistry, error) {
	for index, record := range definitions {
		if utf8Len(record.typ.ShortDescription) > maxShortDescriptionLen {
			return compiledRegistry{}, registryErrorf(
				"record %d short description exceeds %d runes",
				index,
				maxShortDescriptionLen,
			)
		}
	}

	registry := compiledRegistry{
		records:                normalizeRegistryRecords(definitions),
		byID:                   make(map[string]registryRecord, len(definitions)),
		byExtension:            make(map[string]registryRecord, len(definitions)),
		normalProbes:           make(map[TypeID]normalProbeCapability, len(normalProbes)),
		fileNames:              make(map[string]string, len(fileNames)),
		contextualRefinements:  make(map[contextualRefinementKey]TypeID, len(contextual)),
		pathPatternRefinements: append([]pathPatternRefinement(nil), patterns...),
		structuredProbeCapabilities: make(
			[]structuredProbeCapability,
			len(structuredProbes),
		),
	}

	for key, targetID := range contextual {
		registry.contextualRefinements[key] = targetID
	}
	for index, capability := range structuredProbes {
		capability.targets = append([]structuredProbeTarget(nil), capability.targets...)
		registry.structuredProbeCapabilities[index] = capability
	}

	for index := range registry.records {
		record := registry.records[index]
		id := lowerKey(strings.TrimSpace(string(record.typ.ID)))
		if id == "" {
			return compiledRegistry{}, registryErrorf("record %d has an empty type id", index)
		}
		if id == string(UnknownID) {
			return compiledRegistry{}, registryErrorf("type id %q is reserved", id)
		}
		if _, exists := registry.byID[id]; exists {
			return compiledRegistry{}, registryErrorf("duplicate type id %q", id)
		}

		record.typ.ID = TypeID(id)
		family := lowerKey(strings.TrimSpace(string(record.typ.Family)))
		if family == "" {
			family = id
		}

		record.typ.Family = FamilyID(family)
		traits := record.typ.Traits & (TraitText | TraitBinary)
		if traits != TraitText && traits != TraitBinary {
			return compiledRegistry{}, registryErrorf("type %q must declare exactly one text/binary trait", id)
		}
		if utf8Len(record.typ.ShortDescription) > maxShortDescriptionLen {
			return compiledRegistry{}, registryErrorf("type %q short description exceeds %d runes", id, maxShortDescriptionLen)
		}

		record.typ.Extensions = append([]string(nil), record.typ.Extensions...)
		registry.records[index] = record
		registry.byID[id] = record

		for extensionIndex, extension := range record.typ.Extensions {
			ext := normalizeRegistryExtension(extension)
			if ext == "" {
				return compiledRegistry{}, registryErrorf("type %q has empty extension at index %d", id, extensionIndex)
			}
			if _, exists := registry.byExtension[ext]; exists {
				return compiledRegistry{}, registryErrorf("duplicate extension %q", ext)
			}

			record.typ.Extensions[extensionIndex] = ext
			registry.byExtension[ext] = record
		}

		if len(record.magic) > 0 && record.magicScope != probeDiscover && record.magicScope != probeVerify {
			return compiledRegistry{}, registryErrorf("type %q has invalid static probe scope %d", id, record.magicScope)
		}

		for signatureIndex, signature := range record.magic {
			if len(signature) == 0 {
				return compiledRegistry{}, registryErrorf("type %q has empty magic signature at index %d", id, signatureIndex)
			}
			if len(signature) > thoroughPrefixSize {
				return compiledRegistry{}, registryErrorf("type %q magic signature exceeds probe bound", id)
			}
		}
	}

	for index := range registry.records {
		registry.records[index].typ.Extensions = append([]string(nil), registry.records[index].typ.Extensions...)
		registry.byID[string(registry.records[index].typ.ID)] = registry.records[index]
		for _, extension := range registry.records[index].typ.Extensions {
			registry.byExtension[extension] = registry.records[index]
		}
	}

	for _, capability := range normalProbes {
		if _, exists := registry.normalProbes[capability.pathType]; exists {
			return compiledRegistry{}, registryErrorf("duplicate normal probe path type %q", capability.pathType)
		}

		capability.contentTypes = append([]TypeID(nil), capability.contentTypes...)
		registry.normalProbes[capability.pathType] = capability
	}

	var err error
	registry.fileNames, err = compileFileKeys(fileNames, registry.byID, "filename")
	if err != nil {
		return compiledRegistry{}, err
	}

	registry.fileRules, err = compileFileRules(fileRules, registry.byID)
	if err != nil {
		return compiledRegistry{}, err
	}

	registry.magic = buildMagicIndex(registry.records)
	for _, match := range registry.magic {
		if match.scope == probeDiscover {
			registry.thoroughPrefix = max(registry.thoroughPrefix, len(match.signature))
		}
	}
	for _, capability := range registry.structuredProbeCapabilities {
		if capability.scope == probeDiscover {
			registry.thoroughPrefix = max(registry.thoroughPrefix, capability.requiredPrefix)
		}
	}

	if err := validateMagicIndex(registry.magic); err != nil {
		return compiledRegistry{}, err
	}
	if err := validateStructuredProbeCapabilities(registry); err != nil {
		return compiledRegistry{}, err
	}

	if err := validateRefinements(
		registry,
		registry.contextualRefinements,
		registry.normalProbes,
		registry.pathPatternRefinements,
	); err != nil {
		return compiledRegistry{}, err
	}
	registry.verifierRequiredPrefix = compileVerifierRequiredPrefixes(registry)

	return registry, nil
}

// compileVerifierRequiredPrefixes compiles the maximum relevant probe bound for each path type.
func compileVerifierRequiredPrefixes(
	registry compiledRegistry,
) map[TypeID]int {
	familyPrefixes := make(map[FamilyID]int)
	for _, record := range registry.records {
		for _, signature := range record.magic {
			familyPrefixes[record.typ.Family] = max(familyPrefixes[record.typ.Family], len(signature))
		}
	}

	compiled := make(map[TypeID]int, len(registry.records))
	for _, record := range registry.records {
		if required := familyPrefixes[record.typ.Family]; required > 0 {
			compiled[record.typ.ID] = required
		}
	}

	for key := range registry.contextualRefinements {
		pathRecord, pathOK := registry.byID[string(key.pathType)]
		contentRecord, contentOK := registry.byID[string(key.contentType)]
		if !pathOK || !contentOK {
			continue
		}

		for _, signature := range contentRecord.magic {
			compiled[pathRecord.typ.ID] = max(compiled[pathRecord.typ.ID], len(signature))
		}
	}

	for _, capability := range registry.structuredProbeCapabilities {
		for _, target := range capability.targets {
			if _, ok := registry.byID[string(target.typeID)]; ok {
				compiled[target.typeID] = max(compiled[target.typeID], target.verifyPrefix)
			}
		}
	}

	return compiled
}

// validateStructuredProbeCapabilities verifies detector targets against the compiled registry.
func validateStructuredProbeCapabilities(registry compiledRegistry) error {
	seen := make(map[structuredProbeKind]struct{}, len(registry.structuredProbeCapabilities))
	for _, capability := range registry.structuredProbeCapabilities {
		if _, exists := seen[capability.kind]; exists {
			return registryErrorf("duplicate structured probe kind %d", capability.kind)
		}

		seen[capability.kind] = struct{}{}
		if capability.source != ContentSourceStructure || len(capability.targets) == 0 {
			return registryErrorf("structured probe %d has no source or type targets", capability.kind)
		}

		if capability.requiredPrefix < 1 || capability.requiredPrefix > thoroughPrefixSize {
			return registryErrorf(
				"structured probe %d has invalid prefix requirement %d",
				capability.kind,
				capability.requiredPrefix,
			)
		}

		if capability.scope != probeDiscover && capability.scope != probeVerify {
			return registryErrorf("structured probe %d has invalid scope %d", capability.kind, capability.scope)
		}

		seenTargets := make(map[TypeID]struct{}, len(capability.targets))
		for _, target := range capability.targets {
			if _, exists := registry.byID[string(target.typeID)]; !exists {
				return registryErrorf("structured probe %d references unknown type %q", capability.kind, target.typeID)
			}
			if _, exists := seenTargets[target.typeID]; exists {
				return registryErrorf("structured probe %d repeats type %q", capability.kind, target.typeID)
			}

			if target.verifyPrefix < 1 || target.verifyPrefix > capability.requiredPrefix {
				return registryErrorf(
					"structured probe %d target %q has invalid verifier prefix %d",
					capability.kind,
					target.typeID,
					target.verifyPrefix,
				)
			}

			seenTargets[target.typeID] = struct{}{}
		}
	}

	return nil
}

// normalizeRegistryExtension normalizes registry-owned extension declarations.
func normalizeRegistryExtension(extension string) string {
	return lowerKey(strings.TrimPrefix(strings.TrimSpace(extension), "."))
}

// compileFileKeys normalizes exact-filename rules and validates their targets.
func compileFileKeys(
	source map[string]string,
	byID map[string]registryRecord,
	kind string,
) (map[string]string, error) {
	compiled := make(map[string]string, len(source))
	for rawKey, rawTypeID := range source {
		key := lowerKey(strings.TrimSpace(rawKey))
		typeID := lowerKey(strings.TrimSpace(rawTypeID))

		if key == "" || typeID == "" {
			return nil, registryErrorf("empty %s rule or type id", kind)
		}
		if _, ok := byID[typeID]; !ok {
			return nil, registryErrorf("%s rule %q references unknown type %q", kind, key, typeID)
		}
		if _, exists := compiled[key]; exists {
			return nil, registryErrorf("duplicate normalized %s rule %q", kind, key)
		}

		compiled[key] = typeID
	}

	return compiled, nil
}

// compileFileRules normalizes filename rules and validates their targets.
func compileFileRules(
	source []fileRule,
	byID map[string]registryRecord,
) ([]fileRule, error) {
	compiled := make([]fileRule, 0, len(source))
	seen := make(map[[2]string]struct{}, len(source))

	for _, rule := range source {
		rule.prefix = lowerKey(strings.TrimSpace(rule.prefix))
		rule.suffix = lowerKey(strings.TrimSpace(rule.suffix))
		rule.typeID = lowerKey(strings.TrimSpace(rule.typeID))
		if (rule.prefix == "" && rule.suffix == "") || rule.typeID == "" {
			return nil, registryErrorf("empty filename rule or type id")
		}

		if _, ok := byID[rule.typeID]; !ok {
			return nil, registryErrorf(
				"filename rule %q%q references unknown type %q",
				rule.prefix,
				rule.suffix,
				rule.typeID,
			)
		}

		key := [2]string{rule.prefix, rule.suffix}
		if _, exists := seen[key]; exists {
			return nil, registryErrorf("duplicate normalized filename rule %q%q", rule.prefix, rule.suffix)
		}

		seen[key] = struct{}{}
		compiled = append(compiled, rule)
	}

	slices.SortFunc(compiled, func(a, b fileRule) int {
		if (a.prefix != "") != (b.prefix != "") {
			if a.prefix != "" {
				return -1
			}
			return 1
		}
		if len(a.prefix) > len(b.prefix) {
			return -1
		}
		if len(a.prefix) < len(b.prefix) {
			return 1
		}
		if len(a.suffix) > len(b.suffix) {
			return -1
		}
		if len(a.suffix) < len(b.suffix) {
			return 1
		}
		if order := strings.Compare(a.prefix, b.prefix); order != 0 {
			return order
		}

		return strings.Compare(a.suffix, b.suffix)
	})

	return compiled, nil
}

// validateMagicIndex rejects ambiguous duplicate static signatures.
func validateMagicIndex(index []magicMatch) error {
	seen := make(map[string]string, len(index))
	for _, match := range index {
		key := string(match.signature)
		if previous, exists := seen[key]; exists {
			return registryErrorf("duplicate magic signature %q for %q and %q", key, previous, match.typeID)
		}

		seen[key] = match.typeID
	}

	return nil
}

// validateRefinements verifies all contextual and path-pattern references.
func validateRefinements(
	registry compiledRegistry,
	contextual map[contextualRefinementKey]TypeID,
	normalProbes map[TypeID]normalProbeCapability,
	patterns []pathPatternRefinement,
) error {
	for key, targetID := range contextual {
		pathRecord, ok := registry.byID[string(key.pathType)]
		if !ok {
			return registryErrorf("contextual refinement references unknown path type %q", key.pathType)
		}
		if _, ok := registry.byID[string(key.contentType)]; !ok {
			return registryErrorf("contextual refinement references unknown content type %q", key.contentType)
		}

		targetRecord, ok := registry.byID[string(targetID)]
		if !ok {
			return registryErrorf("contextual refinement references unknown target type %q", targetID)
		}

		if targetRecord.typ.ID != pathRecord.typ.ID &&
			(targetRecord.typ.Family != pathRecord.typ.Family ||
				targetRecord.specificity <= pathRecord.specificity) {
			return registryErrorf(
				"contextual target %q is not the path type %q or a more specific member of family %q",
				targetID,
				key.pathType,
				pathRecord.typ.Family,
			)
		}
	}

	for _, refinement := range patterns {
		if refinement.scope != probeRefine {
			return registryErrorf("path-pattern refinement has invalid scope %d", refinement.scope)
		}
		if refinement.requiredPrefix < 1 || refinement.requiredPrefix > semanticPrefixSize {
			return registryErrorf(
				"path-pattern refinement has invalid prefix requirement %d",
				refinement.requiredPrefix,
			)
		}

		pathRecord, ok := registry.byID[string(refinement.pathType)]
		if !ok {
			return registryErrorf("path-pattern refinement references unknown path type %q", refinement.pathType)
		}
		targetRecord, ok := registry.byID[string(refinement.targetType)]
		if !ok {
			return registryErrorf("path-pattern refinement references unknown target type %q", refinement.targetType)
		}
		if len(targetRecord.contentPatterns) == 0 {
			return registryErrorf("path-pattern target %q has no content pattern", refinement.targetType)
		}
		if pathRecord.typ.Family != targetRecord.typ.Family || pathRecord.specificity >= targetRecord.specificity {
			return registryErrorf(
				"path-pattern target %q is not a more specific member of family %q",
				refinement.targetType, pathRecord.typ.Family)
		}
	}

	seenNormalPaths := make(map[TypeID]struct{}, len(normalProbes))
	for _, capability := range normalProbes {
		if capability.scope != probeRefine {
			return registryErrorf("normal probe for %q has invalid scope %d", capability.pathType, capability.scope)
		}
		pathRecord, ok := registry.byID[string(capability.pathType)]
		if !ok {
			return registryErrorf("normal probe references unknown path type %q", capability.pathType)
		}

		if capability.requiredPrefix < 1 || capability.requiredPrefix > thoroughPrefixSize {
			return registryErrorf(
				"normal probe for %q has invalid prefix requirement %d",
				capability.pathType,
				capability.requiredPrefix,
			)
		}

		if _, exists := seenNormalPaths[capability.pathType]; exists {
			return registryErrorf("duplicate normal probe path type %q", capability.pathType)
		}

		seenNormalPaths[capability.pathType] = struct{}{}
		if len(capability.contentTypes) == 0 {
			return registryErrorf("normal probe for %q has no content targets", capability.pathType)
		}

		seenContentTypes := make(map[TypeID]struct{}, len(capability.contentTypes))
		for _, contentType := range capability.contentTypes {
			contentRecord, ok := registry.byID[string(contentType)]
			if !ok {
				return registryErrorf("normal probe for %q references unknown content type %q", capability.pathType, contentType)
			}
			if len(contentRecord.magic) == 0 {
				return registryErrorf("normal probe content type %q has no static signature", contentType)
			}

			for _, signature := range contentRecord.magic {
				if len(signature) > capability.requiredPrefix {
					return registryErrorf(
						"normal probe for %q requires %d bytes but plans %d",
						capability.pathType,
						len(signature),
						capability.requiredPrefix,
					)
				}
			}

			if _, exists := seenContentTypes[contentType]; exists {
				return registryErrorf("normal probe for %q repeats content type %q", capability.pathType, contentType)
			}

			seenContentTypes[contentType] = struct{}{}
			refinedRecord := contentRecord
			if targetID, exists := contextual[contextualRefinementKey{
				pathType: capability.pathType, contentType: contentType,
			}]; exists {
				refinedRecord = registry.byID[string(targetID)]
			}
			if refinedRecord.typ.Family != pathRecord.typ.Family ||
				refinedRecord.specificity <= pathRecord.specificity {
				return registryErrorf(
					"normal probe target %q does not refine path type %q",
					contentType,
					capability.pathType,
				)
			}
		}
	}

	return nil
}

// utf8Len counts runes in a registry description without allocating.
func utf8Len(value string) int {
	count := 0
	for range value {
		count++
	}

	return count
}
