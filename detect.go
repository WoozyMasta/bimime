// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"encoding/binary"
	"slices"
	"strings"
	"unicode/utf8"
)

const maxShortDescriptionLen = 60

// eddsEnfusionMarker is Reserved1[1] encoded as a little-endian FourCC.
const eddsEnfusionMarker uint32 = 0x31464E45

// probeScope controls when a detector is allowed to contribute evidence.
type probeScope uint8

const (
	// probeDiscover allows a probe to classify content without path evidence.
	probeDiscover probeScope = iota
	// probeVerify limits a probe to a matching path classification.
	probeVerify
	// probeRefine limits a probe to an explicitly selected path refinement.
	probeRefine
)

const (
	structuredProbeFORM structuredProbeKind = iota
	structuredProbeRIFF
	structuredProbeMP4
	structuredProbeMP3
	structuredProbeDDS
)

// builtinContextualRefinements declares path-aware format specializations.
var builtinContextualRefinements = map[contextualRefinementKey]TypeID{
	{pathType: "bi.config.main.bin", contentType: "bi.rap"}: "bi.config.main.bin",
	{pathType: "bi.mod.bin", contentType: "bi.rap"}:         "bi.mod.bin",
	{pathType: "bi.rvmat", contentType: "bi.rap"}:           "bi.rvmat.bin",
	{pathType: "bi.surface.bisurf", contentType: "bi.rap"}:  "bi.surface.bisurf.bin",
}

// builtinPathPatternRefinements declares path-scoped semantic specializations.
var builtinPathPatternRefinements = []pathPatternRefinement{
	{
		pathType:       "text.cfg",
		targetType:     "bi.model.cfg",
		scope:          probeRefine,
		requiredPrefix: semanticPrefixSize,
	},
	{
		pathType:       "text.xml",
		targetType:     "bi.ce.ceproject-config",
		scope:          probeRefine,
		requiredPrefix: semanticPrefixSize,
	},
}

// builtinNormalProbeCapabilities declares refinements enabled by ProfileNormal.
var builtinNormalProbeCapabilities = []normalProbeCapability{
	{
		pathType:       "bi.p3d",
		scope:          probeRefine,
		contentTypes:   []TypeID{"bi.p3d.odol", "bi.p3d.mlod"},
		requiredPrefix: 4,
	},
	{
		pathType:       "bi.wrp",
		scope:          probeRefine,
		contentTypes:   []TypeID{"bi.wrp.oprw", "bi.wrp.8wvr", "bi.wrp.9vbw", "bi.wrp.0wzd"},
		requiredPrefix: 4,
	},
	{
		pathType:       "bi.rvmat",
		scope:          probeRefine,
		contentTypes:   []TypeID{"bi.rap"},
		requiredPrefix: 4,
	},
	{
		pathType:       "bi.surface.bisurf",
		scope:          probeRefine,
		contentTypes:   []TypeID{"bi.rap"},
		requiredPrefix: 4,
	},
}

// builtinStructuredProbeCapabilities declares non-static content detectors and scopes.
var builtinStructuredProbeCapabilities = []structuredProbeCapability{
	{
		kind:           structuredProbeFORM,
		requiredPrefix: 16,
		source:         ContentSourceStructure,
		targets: []structuredProbeTarget{
			{typeID: "bi.package.pak", verifyPrefix: 12},
			{typeID: "bi.db.rdb", verifyPrefix: 12},
			{typeID: "bi.world.smd", verifyPrefix: 12},
			{typeID: "bi.font.fnt", verifyPrefix: 12},
			{typeID: "bi.object.xob", verifyPrefix: 12},
			{typeID: "bi.animation.anm", verifyPrefix: 16},
		},
	},
	{
		kind:           structuredProbeRIFF,
		requiredPrefix: 12,
		source:         ContentSourceStructure,
		targets: []structuredProbeTarget{
			{typeID: "audio.wav", verifyPrefix: 12},
			{typeID: "image.webp", verifyPrefix: 12},
		},
	},
	{
		kind:           structuredProbeMP4,
		requiredPrefix: 12,
		source:         ContentSourceStructure,
		targets:        []structuredProbeTarget{{typeID: "video.mp4", verifyPrefix: 12}},
	},
	{
		kind:           structuredProbeMP3,
		scope:          probeVerify,
		requiredPrefix: 4,
		source:         ContentSourceStructure,
		targets:        []structuredProbeTarget{{typeID: "audio.mp3", verifyPrefix: 4}},
	},
	{
		kind:           structuredProbeDDS,
		requiredPrefix: 156,
		source:         ContentSourceStructure,
		targets: []structuredProbeTarget{
			{typeID: "image.dds", verifyPrefix: 4},
			{typeID: "image.edds", verifyPrefix: 4},
		},
	},
}

// contextualRefinementKey identifies a path and content pair that refines to a format-specific type.
type contextualRefinementKey struct {
	// pathType is the type inferred from the logical name.
	pathType TypeID
	// contentType is the generic type inferred from the content.
	contentType TypeID
}

// normalProbeCapability defines content types used to refine one path type in Normal.
type normalProbeCapability struct {
	// pathType is the path classification that enables the probe.
	pathType TypeID
	// contentTypes lists static content types eligible for this refinement.
	contentTypes []TypeID
	// scope identifies this as an explicit path refinement.
	scope probeScope
	// requiredPrefix is the bounded prefix needed by the probe.
	requiredPrefix int
}

// pathPatternRefinement identifies a path-scoped pattern that specializes a generic path type.
type pathPatternRefinement struct {
	// pathType is the generic type inferred from the logical name.
	pathType TypeID
	// targetType is selected when its registered pattern matches.
	targetType TypeID
	// scope identifies this as an explicit path refinement.
	scope probeScope
	// requiredPrefix is the bounded sample needed by this pattern.
	requiredPrefix int
}

// structuredProbeTarget declares one type identified by a structured detector.
type structuredProbeTarget struct {
	// typeID is the stable type id produced by the detector.
	typeID TypeID
	// verifyPrefix is sufficient to reject this target when no content matches.
	verifyPrefix int
}

// structuredProbeCapability binds a structured detector to its byte and type requirements.
type structuredProbeCapability struct {
	// targets lists the types and their individual verifier requirements.
	targets []structuredProbeTarget
	// requiredPrefix is the maximum prefix needed to run this detector fully.
	requiredPrefix int
	// kind selects the detector implementation.
	kind structuredProbeKind
	// scope controls global discovery versus path verification.
	scope probeScope
	// source identifies the evidence kind produced by this detector.
	source ContentSource
}

// structuredProbeKind identifies one detector with non-static parsing rules.
type structuredProbeKind uint8

// Registry returns all registered game types.
func Registry() []Type {
	out := make([]Type, 0, len(builtinRegistry.records))
	for _, record := range builtinRegistry.records {
		out = append(out, cloneType(record.typ))
	}

	return out
}

// Lookup finds one type by stable id.
func Lookup(id TypeID) (Type, bool) {
	record, ok := builtinRegistry.byID[strings.ToLower(strings.TrimSpace(string(id)))]
	if !ok {
		return UnknownType(), false
	}

	return cloneType(record.typ), true
}

// specializeRAPRecordForExtension applies a registered contextual refinement.
func specializeRAPRecordForExtension(byMagicRecord registryRecord, byExtensionID string) registryRecord {
	targetID, ok := builtinRegistry.contextualRefinements[contextualRefinementKey{
		pathType:    TypeID(byExtensionID),
		contentType: byMagicRecord.typ.ID,
	}]
	if !ok {
		return byMagicRecord
	}

	record, ok := builtinRegistry.byID[string(targetID)]
	if !ok {
		return byMagicRecord
	}

	return record
}

// normalizeRegistryRecords returns a copied registry with normalized metadata fields.
func normalizeRegistryRecords(records []registryRecord) []registryRecord {
	out := make([]registryRecord, len(records))
	copy(out, records)

	for i := range out {
		if out[i].typ.Family == "" {
			out[i].typ.Family = FamilyID(out[i].typ.ID)
		}
		out[i].typ.ShortDescription = normalizeShortDescription(
			out[i].typ.ShortDescription,
			out[i].typ.Description,
		)
	}

	return out
}

// normalizeShortDescription derives compact description when it is not set.
func normalizeShortDescription(shortDescription string, description string) string {
	shortDescription = strings.TrimSpace(shortDescription)
	if shortDescription == "" {
		shortDescription = strings.TrimSpace(description)
	}
	if shortDescription == "" {
		return ""
	}
	if utf8.RuneCountInString(shortDescription) <= maxShortDescriptionLen {
		return shortDescription
	}

	if idx := strings.Index(shortDescription, " ("); idx > 0 {
		candidate := strings.TrimSpace(shortDescription[:idx])
		if utf8.RuneCountInString(candidate) <= maxShortDescriptionLen {
			return candidate
		}
	}

	runes := []rune(shortDescription)
	if len(runes) <= maxShortDescriptionLen {
		return shortDescription
	}
	if maxShortDescriptionLen <= 1 {
		return string(runes[:maxShortDescriptionLen])
	}
	if maxShortDescriptionLen <= 3 {
		return string(runes[:maxShortDescriptionLen])
	}

	return string(runes[:maxShortDescriptionLen-3]) + "..."
}

// detectByPathRecord resolves by known filename first, then by extension.
func detectByPathRecord(path string) (registryRecord, bool) {
	return detectByPathRecordWithExtension(path, extensionKey(path))
}

// detectByPathRecordWithExtension resolves by known filename first,
// then by precomputed extension key.
func detectByPathRecordWithExtension(path string, extension string) (registryRecord, bool) {
	fileName := fileNameKey(path)
	if typeID, ok := builtinRegistry.fileNames[fileName]; ok {
		record, ok := builtinRegistry.byID[typeID]
		if ok {
			return record, true
		}
	}

	for _, rule := range builtinRegistry.fileRules {
		if rule.prefix != "" && !strings.HasPrefix(fileName, rule.prefix) {
			continue
		}
		if rule.suffix != "" && !strings.HasSuffix(fileName, rule.suffix) {
			continue
		}

		record, ok := builtinRegistry.byID[rule.typeID]
		if !ok {
			continue
		}

		return record, true
	}

	return detectByExtensionKeyRecord(extension)
}

// detectByExtensionKeyRecord resolves type by already-normalized extension key.
func detectByExtensionKeyRecord(extension string) (registryRecord, bool) {
	if extension == "" {
		return registryRecord{}, false
	}

	record, ok := builtinRegistry.byExtension[extension]
	if !ok {
		return registryRecord{}, false
	}

	return record, true
}

// fileNameKey returns normalized file basename for path-based lookup.
func fileNameKey(path string) string {
	if path == "" {
		return ""
	}

	start := baseIndex(path)
	return lowerKey(path[start:])
}

// detectByMagicRecord resolves static and structured signatures without cloning.
func detectByMagicRecord(prefix []byte) (registryRecord, ContentSource, bool) {
	if len(prefix) == 0 {
		return registryRecord{}, ContentSourceNone, false
	}

	for _, capability := range builtinRegistry.structuredProbeCapabilities {
		if capability.scope != probeDiscover {
			continue
		}

		typeID, ok := runStructuredProbe(capability.kind, prefix)
		if !ok || !hasStructuredProbeTarget(capability, typeID) {
			continue
		}

		record, ok := builtinRegistry.byID[string(typeID)]
		if ok {
			return record, capability.source, true
		}
	}

	for _, signature := range builtinRegistry.magic {
		if signature.scope != probeDiscover {
			continue
		}

		if len(prefix) < len(signature.signature) {
			continue
		}
		if !slices.Equal(prefix[:len(signature.signature)], signature.signature) {
			continue
		}

		record, ok := builtinRegistry.byID[signature.typeID]
		if !ok {
			return registryRecord{}, ContentSourceNone, false
		}

		return record, ContentSourceMagic, true
	}

	return registryRecord{}, ContentSourceNone, false
}

// detectVerifyRecord runs verification probes enabled for one path type.
func detectVerifyRecord(prefix []byte, pathType TypeID) (registryRecord, ContentSource, bool) {
	if len(prefix) == 0 {
		return registryRecord{}, ContentSourceNone, false
	}

	for _, capability := range builtinRegistry.structuredProbeCapabilities {
		if capability.scope != probeVerify || !hasStructuredProbeTarget(capability, pathType) {
			continue
		}

		typeID, ok := runStructuredProbe(capability.kind, prefix)
		if !ok || !hasStructuredProbeTarget(capability, typeID) {
			continue
		}

		record, ok := builtinRegistry.byID[string(typeID)]
		if ok {
			return record, capability.source, true
		}
	}

	for _, signature := range builtinRegistry.magic {
		if signature.scope != probeVerify || signature.typeID != string(pathType) ||
			len(prefix) < len(signature.signature) ||
			!slices.Equal(prefix[:len(signature.signature)], signature.signature) {
			continue
		}

		record, ok := builtinRegistry.byID[signature.typeID]
		if ok {
			return record, ContentSourceMagic, true
		}
	}

	return registryRecord{}, ContentSourceNone, false
}

// hasStructuredProbeTarget reports whether a detector declares the requested type.
func hasStructuredProbeTarget(capability structuredProbeCapability, typeID TypeID) bool {
	for _, target := range capability.targets {
		if target.typeID == typeID {
			return true
		}
	}

	return false
}

// detectNormalProbeRecord runs only the static targets enabled for one path type.
func detectNormalProbeRecord(prefix []byte, pathType TypeID) (registryRecord, bool) {
	capability, ok := builtinRegistry.normalProbes[pathType]
	if !ok || capability.scope != probeRefine || len(prefix) == 0 {
		return registryRecord{}, false
	}

	for _, signature := range builtinRegistry.magic {
		if !slices.Contains(capability.contentTypes, TypeID(signature.typeID)) ||
			len(prefix) < len(signature.signature) ||
			!slices.Equal(prefix[:len(signature.signature)], signature.signature) {
			continue
		}

		record, ok := builtinRegistry.byID[signature.typeID]
		if ok {
			return record, true
		}
	}

	return registryRecord{}, false
}

// runStructuredProbe dispatches one declared capability to its format parser.
func runStructuredProbe(kind structuredProbeKind, prefix []byte) (TypeID, bool) {
	var (
		id string
		ok bool
	)

	switch kind {
	case structuredProbeFORM:
		id, ok = detectFORMFamilyTypeID(prefix)

	case structuredProbeRIFF:
		id, ok = detectRIFFFamilyTypeID(prefix)

	case structuredProbeMP4:
		id, ok = detectMP4TypeID(prefix)

	case structuredProbeMP3:
		id, ok = detectMP3TypeID(prefix)

	case structuredProbeDDS:
		id, ok = detectDDSFamilyTypeID(prefix)

	default:
		return "", false
	}

	return TypeID(id), ok
}

// detectFORMFamilyTypeID detects FORM-based payloads using 4-byte form type
// at offset 8 and optional subtype marker at offset 12.
func detectFORMFamilyTypeID(prefix []byte) (string, bool) {
	if len(prefix) < 12 {
		return "", false
	}
	if prefix[0] != 'F' || prefix[1] != 'O' || prefix[2] != 'R' || prefix[3] != 'M' {
		return "", false
	}

	switch {
	case prefix[8] == 'P' && prefix[9] == 'A' && prefix[10] == 'C' && prefix[11] == '1':
		return "bi.package.pak", true

	case prefix[8] == 'R' && prefix[9] == 'D' && prefix[10] == 'B' && prefix[11] == 'C':
		return "bi.db.rdb", true

	case prefix[8] == 'S' && prefix[9] == 'M' && prefix[10] == 'D' && prefix[11] == 'F':
		return "bi.world.smd", true

	case prefix[8] == 'F' && prefix[9] == 'N' && prefix[10] == 'T':
		return "bi.font.fnt", true

	case prefix[8] == 'X' && prefix[9] == 'O' && prefix[10] == 'B':
		return "bi.object.xob", true

	case prefix[8] == 'A' && prefix[9] == 'N' && prefix[10] == 'I' && prefix[11] == 'M':
		if len(prefix) < 16 {
			return "", false
		}
		if prefix[12] == 'S' && prefix[13] == 'E' && prefix[14] == 'T' {
			return "bi.animation.anm", true
		}
	}

	return "", false
}

// detectRIFFFamilyTypeID detects RIFF-based media payloads using form type at offset 8.
func detectRIFFFamilyTypeID(prefix []byte) (string, bool) {
	if len(prefix) < 12 {
		return "", false
	}
	if prefix[0] != 'R' || prefix[1] != 'I' || prefix[2] != 'F' || prefix[3] != 'F' {
		return "", false
	}

	switch {
	case prefix[8] == 'W' && prefix[9] == 'A' && prefix[10] == 'V' && prefix[11] == 'E':
		return "audio.wav", true

	case prefix[8] == 'W' && prefix[9] == 'E' && prefix[10] == 'B' && prefix[11] == 'P':
		return "image.webp", true
	}

	return "", false
}

// detectMP4TypeID detects ISO BMFF containers by ftyp box.
func detectMP4TypeID(prefix []byte) (string, bool) {
	if len(prefix) < 12 {
		return "", false
	}
	if prefix[4] != 'f' || prefix[5] != 't' || prefix[6] != 'y' || prefix[7] != 'p' {
		return "", false
	}

	switch {
	case prefix[8] == 'i' && prefix[9] == 's' && prefix[10] == 'o' && prefix[11] == 'm':
	case prefix[8] == 'i' && prefix[9] == 's' && prefix[10] == 'o' && prefix[11] == '2':
	case prefix[8] == 'm' && prefix[9] == 'p' && prefix[10] == '4' && prefix[11] == '1':
	case prefix[8] == 'm' && prefix[9] == 'p' && prefix[10] == '4' && prefix[11] == '2':
	case prefix[8] == 'a' && prefix[9] == 'v' && prefix[10] == 'c' && prefix[11] == '1':
	case prefix[8] == 'd' && prefix[9] == 'a' && prefix[10] == 's' && prefix[11] == 'h':
	case prefix[8] == 'm' && prefix[9] == 'm' && prefix[10] == 'p' && prefix[11] == '4':
	case prefix[8] == 'M' && prefix[9] == 'S' && prefix[10] == 'N' && prefix[11] == 'V':
	case prefix[8] == '3' && prefix[9] == 'g' && prefix[10] == 'p' && prefix[11] == '4':
	case prefix[8] == '3' && prefix[9] == 'g' && prefix[10] == 'p' && prefix[11] == '5':
	case prefix[8] == 'M' && prefix[9] == '4' && prefix[10] == 'V' && prefix[11] == ' ':
	case prefix[8] == 'M' && prefix[9] == '4' && prefix[10] == 'A' && prefix[11] == ' ':
	default:
		return "", false
	}

	return "video.mp4", true
}

// detectMP3TypeID detects MP3 stream headers without ID3 tag.
func detectMP3TypeID(prefix []byte) (string, bool) {
	if len(prefix) < 4 {
		return "", false
	}
	if prefix[0] != 0xFF || (prefix[1]&0xE0) != 0xE0 {
		return "", false
	}

	version := (prefix[1] >> 3) & 0x03
	layer := (prefix[1] >> 1) & 0x03
	bitrate := (prefix[2] >> 4) & 0x0F
	sampleRate := (prefix[2] >> 2) & 0x03

	if version == 0x01 || layer == 0 || bitrate == 0 || bitrate == 0x0F || sampleRate == 0x03 {
		return "", false
	}

	return "audio.mp3", true
}

// detectDDSFamilyTypeID disambiguates DDS vs EDDS for payloads with DDS magic.
func detectDDSFamilyTypeID(prefix []byte) (string, bool) {
	if len(prefix) < 4 {
		return "", false
	}
	if prefix[0] != 'D' || prefix[1] != 'D' || prefix[2] != 'S' || prefix[3] != ' ' {
		return "", false
	}

	if isEDDSPrefix(prefix) {
		return "image.edds", true
	}

	return "image.dds", true
}

// isEDDSPrefix reports positive EDDS markers without treating their absence as DDS proof.
func isEDDSPrefix(prefix []byte) bool {
	// DDS magic and the header fields place Reserved1[1] at bytes 36 through 39.
	const reserved1SecondOffset = 36
	const reserved1SecondEnd = reserved1SecondOffset + 4
	if len(prefix) >= reserved1SecondEnd &&
		binary.LittleEndian.Uint32(prefix[reserved1SecondOffset:reserved1SecondEnd]) == eddsEnfusionMarker {
		return true
	}

	offset, ok := eddsBlockTableOffset(prefix)
	if !ok || len(prefix) < offset+8 {
		return false
	}

	if !isEddsBlockMagic(prefix[offset : offset+4]) {
		return false
	}

	size := binary.LittleEndian.Uint32(prefix[offset+4 : offset+8])
	return size <= 0x7FFFFFFF
}

// eddsBlockTableOffset returns first block-table offset and whether it is readable.
func eddsBlockTableOffset(prefix []byte) (int, bool) {
	if len(prefix) < 4 {
		return 0, false
	}
	if prefix[0] != 'D' || prefix[1] != 'D' || prefix[2] != 'S' || prefix[3] != ' ' {
		return 0, false
	}
	if len(prefix) < 88 {
		return 0, false
	}

	if prefix[84] == 'D' && prefix[85] == 'X' && prefix[86] == '1' && prefix[87] == '0' {
		const offsetDX10 = 4 + 124 + 20 // DDS magic + header + DX10 header.
		if len(prefix) < offsetDX10+8 {
			return 0, false
		}

		return offsetDX10, true
	}

	const offset = 4 + 124 // DDS magic + header.
	if len(prefix) < offset+8 {
		return 0, false
	}

	return offset, true
}

// isEddsBlockMagic reports whether 4-byte value is known EDDS block marker.
func isEddsBlockMagic(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	if data[0] == 'C' && data[1] == 'O' && data[2] == 'P' && data[3] == 'Y' {
		return true
	}
	if data[0] == 'L' && data[1] == 'Z' && data[2] == '4' && data[3] == ' ' {
		return true
	}

	return false
}

// extensionKey returns normalized extension without leading dot.
func extensionKey(path string) string {
	if path == "" {
		return ""
	}

	start := baseIndex(path)
	dot := -1
	for i := len(path) - 1; i >= start; i-- {
		if path[i] == '.' {
			dot = i
			break
		}
	}

	// No dot, leading-dot basename, or trailing dot: no extension.
	if dot <= start || dot == len(path)-1 {
		return ""
	}

	return lowerKey(path[dot+1:])
}

// baseIndex returns basename start index for both slash and backslash paths.
func baseIndex(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return i + 1
		}
	}

	return 0
}

// lowerKey lowers ASCII keys with minimal allocations and falls back for UTF-8.
func lowerKey(value string) string {
	if value == "" {
		return ""
	}

	hasUpper := false
	hasNonASCII := false
	for i := 0; i < len(value); i++ {
		b := value[i]
		if b >= 'A' && b <= 'Z' {
			hasUpper = true
		}
		if b >= utf8.RuneSelf {
			hasNonASCII = true
			break
		}
	}

	if !hasUpper && !hasNonASCII {
		return value
	}
	if hasNonASCII {
		return strings.ToLower(value)
	}

	buf := []byte(value)
	for i := range buf {
		if buf[i] >= 'A' && buf[i] <= 'Z' {
			buf[i] += 'a' - 'A'
		}
	}

	return string(buf)
}

// buildMagicIndex flattens and sorts signatures by descending length.
func buildMagicIndex(records []registryRecord) []magicMatch {
	flat := make([]magicMatch, 0, 32)
	for _, record := range records {
		if len(record.magic) == 0 {
			continue
		}

		for _, signature := range record.magic {
			if len(signature) == 0 {
				continue
			}

			copied := make([]byte, len(signature))
			copy(copied, signature)
			flat = append(flat, magicMatch{
				signature: copied,
				typeID:    strings.ToLower(string(record.typ.ID)),
				scope:     record.magicScope,
			})
		}
	}

	slices.SortStableFunc(flat, func(a, b magicMatch) int {
		if len(a.signature) > len(b.signature) {
			return -1
		}
		if len(a.signature) < len(b.signature) {
			return 1
		}

		return 0
	})

	return flat
}

// cloneType deep-copies type metadata to protect internal registry slices.
func cloneType(typ Type) Type {
	out := typ
	if len(typ.Extensions) > 0 {
		out.Extensions = append([]string(nil), typ.Extensions...)
	}

	return out
}
