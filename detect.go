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

// forceTextTypeIDs marks custom MIME ids that should be treated as text.
var forceTextTypeIDs = map[string]struct{}{
	"bi.ui.layout":       {},
	"bi.ui.imageset":     {},
	"bi.ui.styles":       {},
	"bi.animgraph.ast":   {},
	"bi.animgraph.asi":   {},
	"bi.animgraph.asy":   {},
	"bi.animgraph.aw":    {},
	"bi.animgraph.agr":   {},
	"bi.emat":            {},
	"bi.effects.ptc":     {},
	"bi.meta":            {},
	"bi.font.fsdf":       {},
	"bi.project.gproj":   {},
	"bi.project.sproj":   {},
	"bi.project.ssln":    {},
	"bi.config.main.cpp": {},
	"bi.mod.cpp":         {},
	"bi.model.cfg":       {},
	"image.svg":          {},
	"model.obj":          {},
	"model.dae":          {},
	"model.gltf":         {},
}

const maxShortDescriptionLen = 60

// filePrefixRecord stores one filename-prefix mapping to a type id.
type filePrefixRecord struct {
	prefix string
	typeID string
}

// fileSuffixRecord stores one filename-suffix mapping to a type id.
type fileSuffixRecord struct {
	suffix string
	typeID string
}

// filePrefixRecords stores prefix mappings in stable matching order.
var filePrefixRecords = buildFilePrefixRecords(typeIDByFilePrefix)

// fileSuffixRecords stores suffix mappings in stable matching order.
var fileSuffixRecords = buildFileSuffixRecords(typeIDByFileSuffix)

// Registry returns all registered game types.
func Registry() []Type {
	out := make([]Type, 0, len(registryRecords))
	for _, record := range registryRecords {
		out = append(out, cloneType(record.typ))
	}

	return out
}

// Lookup finds one type by stable id.
func Lookup(id string) (Type, bool) {
	record, ok := typeByID[strings.ToLower(strings.TrimSpace(id))]
	if !ok {
		return UnknownType, false
	}

	return cloneType(record.typ), true
}

// Detect returns final resolved type for given filename and payload prefix.
func Detect(path string, prefix []byte) Type {
	return Probe(path, prefix).Resolved
}

// Probe resolves by both path hint and magic; magic match has priority.
func Probe(path string, prefix []byte) ProbeResult {
	extension := extensionKey(path)
	byExtensionRecord, okExtension := detectByPathRecordWithExtension(path, extension)
	if okExtension {
		byExtensionRecord = specializeByExtensionAndContent(byExtensionRecord, prefix)
	}

	byMagicRecord, okMagic := detectByMagicRecord(prefix)

	result := ProbeResult{
		Resolved:    UnknownType,
		ByMagic:     UnknownType,
		ByExtension: UnknownType,
		Source:      SourceUnknown,
		Extension:   extension,
	}

	if okExtension {
		result.ByExtension = probeType(byExtensionRecord.typ)
	}
	if okMagic {
		if okExtension {
			byMagicRecord = specializeRAPRecordForExtension(byMagicRecord, byExtensionRecord.typ.ID)
		}

		result.ByMagic = probeType(byMagicRecord.typ)
	}

	switch {
	case okMagic && okExtension:
		if shouldPreferExtension(byExtensionRecord.typ.ID, byMagicRecord.typ.ID, prefix) {
			result.Resolved = result.ByExtension
		} else {
			result.Resolved = result.ByMagic
		}
		result.Source = SourceMagicAndExtension
	case okMagic:
		result.Resolved = result.ByMagic
		result.Source = SourceMagic
	case okExtension:
		result.Resolved = result.ByExtension
		result.Source = SourceExtension
	}

	return result
}

// specializeRAPRecordForExtension maps generic RAP match to known extension-specific
// RAP-like bins when path hint is available.
func specializeRAPRecordForExtension(byMagicRecord registryRecord, byExtensionID string) registryRecord {
	if byMagicRecord.typ.ID != "bi.rap" {
		return byMagicRecord
	}

	var targetID string
	switch byExtensionID {
	case "bi.config.main.bin", "bi.mod.bin":
		targetID = byExtensionID
	case "bi.rvmat":
		targetID = "bi.rvmat.bin"
	case "bi.surface.bisurf":
		targetID = "bi.surface.bisurf.bin"
	default:
		return byMagicRecord
	}

	record, ok := typeByID[targetID]
	if !ok {
		return byMagicRecord
	}

	return record
}

// specializeByExtensionAndContent refines extension match by lightweight
// content hints for known ambiguous text formats.
func specializeByExtensionAndContent(
	byExtensionRecord registryRecord,
	prefix []byte,
) registryRecord {
	if byExtensionRecord.typ.ID == "text.cfg" &&
		matchContentPatternForType("bi.model.cfg", prefix) {
		record, ok := typeByID["bi.model.cfg"]
		if ok {
			return record
		}
	}

	return byExtensionRecord
}

// DetectByExtension resolves type by path hint only (well-known filename/extension).
func DetectByExtension(path string) (Type, bool) {
	record, ok := detectByPathRecordWithExtension(path, extensionKey(path))
	if !ok {
		return UnknownType, false
	}

	return cloneType(record.typ), true
}

// DetectByMagic resolves type only by payload magic bytes.
func DetectByMagic(prefix []byte) (Type, bool) {
	record, ok := detectByMagicRecord(prefix)
	if !ok {
		return UnknownType, false
	}

	return cloneType(record.typ), true
}

// IsRAP reports whether payload starts with RAP magic bytes.
func IsRAP(prefix []byte) bool {
	record, ok := detectByMagicRecord(prefix)
	return ok && record.typ.ID == "bi.rap"
}

// normalizeRegistryRecords returns a copied registry with normalized metadata fields.
func normalizeRegistryRecords(records []registryRecord) []registryRecord {
	out := make([]registryRecord, len(records))
	copy(out, records)

	for i := range out {
		out[i].typ.ShortDescription = normalizeShortDescription(
			out[i].typ.ShortDescription,
			out[i].typ.Description,
		)
		out[i].typ.Binary = inferBinaryByType(out[i].typ)
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

// inferBinaryByType infers binary/text hint from MIME and known type ids.
func inferBinaryByType(typ Type) bool {
	id := strings.ToLower(strings.TrimSpace(typ.ID))
	mime := strings.ToLower(strings.TrimSpace(typ.MIME))

	if _, ok := forceTextTypeIDs[id]; ok {
		return false
	}
	if strings.HasPrefix(mime, "text/") {
		return false
	}
	if mime == "application/json" || mime == "application/xml" {
		return false
	}

	return true
}

// detectByPathRecord resolves by known filename first, then by extension.
func detectByPathRecord(path string) (registryRecord, bool) {
	return detectByPathRecordWithExtension(path, extensionKey(path))
}

// detectByPathRecordWithExtension resolves by known filename first, then by
// precomputed extension key.
func detectByPathRecordWithExtension(path string, extension string) (registryRecord, bool) {
	fileName := fileNameKey(path)
	if typeID, ok := typeIDByFileName[fileName]; ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}
	for _, prefixRecord := range filePrefixRecords {
		if !strings.HasPrefix(fileName, prefixRecord.prefix) {
			continue
		}

		record, ok := typeByID[prefixRecord.typeID]
		if !ok {
			continue
		}

		return record, true
	}
	for _, suffixRecord := range fileSuffixRecords {
		if !strings.HasSuffix(fileName, suffixRecord.suffix) {
			continue
		}

		record, ok := typeByID[suffixRecord.typeID]
		if !ok {
			continue
		}

		return record, true
	}

	return detectByExtensionKeyRecord(extension)
}

// shouldPreferExtension resolves known ambiguous cases where magic is insufficient.
func shouldPreferExtension(byExtensionID string, byMagicID string, prefix []byte) bool {
	if byExtensionID == "image.edds" && byMagicID == "image.dds" && !canDisambiguateEDDS(prefix) {
		return true
	}
	if byMagicID == "bi.rap" {
		switch byExtensionID {
		case "bi.config.main.bin", "bi.mod.bin":
			return true
		}
	}

	return false
}

// detectByExtensionKeyRecord resolves type by already-normalized extension key.
func detectByExtensionKeyRecord(extension string) (registryRecord, bool) {
	if extension == "" {
		return registryRecord{}, false
	}

	record, ok := typeByExtension[extension]
	if !ok {
		return registryRecord{}, false
	}

	return record, true
}

// fileNameKey returns normalized file basename for path-based lookup.
func fileNameKey(path string) string {
	path = trimPathSpace(path)
	if path == "" {
		return ""
	}

	start := baseIndex(path)
	return lowerKey(path[start:])
}

// detectByMagicRecord resolves type by signature matching without cloning.
func detectByMagicRecord(prefix []byte) (registryRecord, bool) {
	if len(prefix) == 0 {
		return registryRecord{}, false
	}

	if typeID, ok := detectFORMFamilyTypeID(prefix); ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}
	if typeID, ok := detectRIFFFamilyTypeID(prefix); ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}
	if typeID, ok := detectMP4TypeID(prefix); ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}
	if typeID, ok := detectMP3TypeID(prefix); ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}
	if typeID, ok := detectDDSFamilyTypeID(prefix); ok {
		record, ok := typeByID[typeID]
		if ok {
			return record, true
		}
	}

	for _, signature := range magicIndex {
		if len(prefix) < len(signature.signature) {
			continue
		}
		if !slices.Equal(prefix[:len(signature.signature)], signature.signature) {
			continue
		}

		record, ok := typeByID[signature.typeID]
		if !ok {
			return registryRecord{}, false
		}

		return record, true
	}

	return registryRecord{}, false
}

// detectFORMFamilyTypeID detects FORM-based payloads using 4-byte form type at
// offset 8 and optional subtype marker at offset 12.
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

// detectRIFFFamilyTypeID detects RIFF-based media payloads using form type at
// offset 8.
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
	if len(prefix) < 2 {
		return "", false
	}
	if prefix[0] != 0xFF || (prefix[1]&0xE0) != 0xE0 {
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

// isEDDSPrefix reports whether DDS payload prefix looks like an EDDS block table.
func isEDDSPrefix(prefix []byte) bool {
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

// canDisambiguateEDDS reports whether prefix is long enough to tell DDS from EDDS.
func canDisambiguateEDDS(prefix []byte) bool {
	_, ok := eddsBlockTableOffset(prefix)
	return ok
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
	path = trimPathSpace(path)
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

// trimPathSpace trims surrounding spaces only when needed.
func trimPathSpace(path string) string {
	if path == "" {
		return ""
	}
	if path[0] > ' ' && path[len(path)-1] > ' ' {
		return path
	}

	return strings.TrimSpace(path)
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

// buildTypeByID builds type index by stable id and skips invalid duplicates.
func buildTypeByID(records []registryRecord) map[string]registryRecord {
	index := make(map[string]registryRecord, len(records))
	for _, record := range records {
		key := strings.ToLower(strings.TrimSpace(record.typ.ID))
		if key == "" {
			continue
		}
		if _, exists := index[key]; exists {
			continue
		}

		index[key] = record
	}

	return index
}

// buildFilePrefixRecords builds stable filename-prefix matcher list.
func buildFilePrefixRecords(source map[string]string) []filePrefixRecord {
	keys := buildFileKeyRecords(source)
	records := make([]filePrefixRecord, 0, len(keys))
	for _, key := range keys {
		records = append(records, filePrefixRecord{
			prefix: key.key,
			typeID: key.typeID,
		})
	}

	return records
}

// buildFileSuffixRecords builds stable filename-suffix matcher list.
func buildFileSuffixRecords(source map[string]string) []fileSuffixRecord {
	keys := buildFileKeyRecords(source)
	records := make([]fileSuffixRecord, 0, len(keys))
	for _, key := range keys {
		records = append(records, fileSuffixRecord{
			suffix: key.key,
			typeID: key.typeID,
		})
	}

	return records
}

// fileKeyRecord stores normalized key->type mapping for prefix/suffix indexes.
type fileKeyRecord struct {
	key    string
	typeID string
}

// buildFileKeyRecords builds stable normalized filename key records.
func buildFileKeyRecords(source map[string]string) []fileKeyRecord {
	records := make([]fileKeyRecord, 0, len(source))
	seen := make(map[string]struct{}, len(source))

	for rawKey, typeID := range source {
		key := lowerKey(strings.TrimSpace(rawKey))
		keyTypeID := lowerKey(strings.TrimSpace(typeID))

		if key == "" {
			continue
		}
		if keyTypeID == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}

		records = append(records, fileKeyRecord{
			key:    key,
			typeID: keyTypeID,
		})
		seen[key] = struct{}{}
	}

	slices.SortFunc(records, func(a, b fileKeyRecord) int {
		if len(a.key) > len(b.key) {
			return -1
		}
		if len(a.key) < len(b.key) {
			return 1
		}

		return strings.Compare(a.key, b.key)
	})

	return records
}

// buildTypeByExtension builds extension index and skips invalid duplicates.
func buildTypeByExtension(records []registryRecord) map[string]registryRecord {
	index := make(map[string]registryRecord, len(records))
	for _, record := range records {
		for _, ext := range record.typ.Extensions {
			key := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))

			if key == "" {
				continue
			}
			if _, exists := index[key]; exists {
				continue
			}

			index[key] = record
		}
	}

	return index
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
			flat = append(flat, magicMatch{signature: copied, typeID: strings.ToLower(record.typ.ID)})
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

// probeType returns lightweight type snapshot without extension list allocations.
func probeType(typ Type) Type {
	out := typ
	out.Extensions = nil

	return out
}

// cloneType deep-copies type metadata to protect internal registry slices.
func cloneType(typ Type) Type {
	out := typ
	if len(typ.Extensions) > 0 {
		out.Extensions = append([]string(nil), typ.Extensions...)
	}

	return out
}
