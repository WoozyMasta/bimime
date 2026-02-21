// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const defaultAnalyzePrefixSize = 4096

// ParseDetectMode parses user-facing mode string.
func ParseDetectMode(value string) (DetectMode, error) {
	mode := DetectMode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case "", DetectModeNormal:
		return DetectModeNormal, nil
	case DetectModeFast, DetectModeStrict:
		return mode, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidDetectMode, value)
	}
}

// NeedsContent reports whether selected mode needs payload bytes for detection.
func NeedsContent(path string, mode DetectMode) bool {
	mode = normalizeDetectMode(mode)
	switch mode {
	case DetectModeFast:
		return false
	case DetectModeStrict:
		return true
	case DetectModeNormal:
		byExtension, ok := detectByPathRecord(path)
		if !ok {
			return true
		}

		return hasMagicForType(byExtension.typ.ID)
	default:
		return true
	}
}

// HasMagic reports whether a registered type has at least one magic signature.
func HasMagic(typeID string) bool {
	return hasMagicForType(typeID)
}

// Analyze classifies path/prefix pair according to selected mode.
func Analyze(path string, prefix []byte, options AnalyzeOptions) AnalyzeResult {
	opts := normalizeAnalyzeOptions(options)
	probe := probeForMode(path, prefix, opts.Mode)

	result := AnalyzeResult{
		Mode:  opts.Mode,
		Probe: probe,
		Valid: true,
	}

	if opts.Mode != DetectModeStrict {
		return result
	}

	validateStrict(&result, prefix)
	return result
}

// AnalyzeReader classifies path using prefix bytes read from reader when required.
func AnalyzeReader(path string, reader io.Reader, options AnalyzeOptions) (AnalyzeResult, error) {
	opts := normalizeAnalyzeOptions(options)
	if !NeedsContent(path, opts.Mode) {
		return Analyze(path, nil, opts), nil
	}
	if reader == nil {
		return AnalyzeResult{}, ErrNilReader
	}

	prefix, err := readPrefix(reader, opts.PrefixSize)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("read content prefix: %w", err)
	}

	return Analyze(path, prefix, opts), nil
}

// AnalyzeFile classifies file path and reads only required prefix bytes.
func AnalyzeFile(path string, options AnalyzeOptions) (AnalyzeResult, error) {
	opts := normalizeAnalyzeOptions(options)
	if !NeedsContent(path, opts.Mode) {
		return Analyze(path, nil, opts), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return AnalyzeReader(path, f, opts)
}

// normalizeAnalyzeOptions fills defaults and normalizes detect mode.
func normalizeAnalyzeOptions(options AnalyzeOptions) AnalyzeOptions {
	options.Mode = normalizeDetectMode(options.Mode)
	if options.PrefixSize <= 0 {
		options.PrefixSize = defaultAnalyzePrefixSize
	}

	return options
}

// normalizeDetectMode normalizes empty/invalid mode to normal mode.
func normalizeDetectMode(mode DetectMode) DetectMode {
	switch mode {
	case DetectModeFast, DetectModeNormal, DetectModeStrict:
		return mode
	default:
		return DetectModeNormal
	}
}

// probeForMode picks extension-only or extension+magic probing by mode.
func probeForMode(path string, prefix []byte, mode DetectMode) ProbeResult {
	if mode == DetectModeFast {
		return extensionOnlyProbe(path)
	}

	return Probe(path, prefix)
}

// extensionOnlyProbe resolves type by path hint without payload magic checks.
func extensionOnlyProbe(path string) ProbeResult {
	extension := extensionKey(path)
	byExtension, ok := detectByPathRecord(path)
	if !ok {
		return ProbeResult{
			Source:      SourceUnknown,
			Extension:   extension,
			Resolved:    UnknownType,
			ByMagic:     UnknownType,
			ByExtension: UnknownType,
		}
	}

	byExtensionType := probeType(byExtension.typ)

	return ProbeResult{
		Source:      SourceExtension,
		Extension:   extension,
		Resolved:    byExtensionType,
		ByMagic:     UnknownType,
		ByExtension: byExtensionType,
	}
}

// validateStrict validates extension/magic consistency and text-like payloads.
func validateStrict(result *AnalyzeResult, prefix []byte) {
	if result == nil {
		return
	}

	if result.Probe.ByExtension.ID != UnknownType.ID && hasMagicForType(result.Probe.ByExtension.ID) {
		result.CheckedMagic = true
		if result.Probe.ByMagic.ID != result.Probe.ByExtension.ID {
			result.Valid = false
			result.Issues = append(result.Issues, AnalyzeIssueMagicMismatch)
		}
	}

	if result.Probe.Resolved.Binary {
		return
	}

	result.CheckedText = true
	result.LooksText = isLikelyText(prefix)
	if !result.LooksText {
		result.Valid = false
		result.Issues = append(result.Issues, AnalyzeIssueTextExpected)
	}
}

// hasMagicForType reports whether registry record has at least one magic signature.
func hasMagicForType(typeID string) bool {
	record, ok := typeByID[typeID]
	if ok {
		return len(record.magic) > 0
	}

	key := strings.ToLower(strings.TrimSpace(typeID))
	record, ok = typeByID[key]
	if !ok {
		return false
	}

	return len(record.magic) > 0
}

// readPrefix reads at most limit bytes from reader and tolerates short inputs.
func readPrefix(reader io.Reader, limit int) ([]byte, error) {
	if limit <= 0 {
		return nil, nil
	}

	initialCap := min(limit, 256)

	buf := make([]byte, 0, initialCap)
	var tmp [256]byte
	for len(buf) < limit {
		chunk := min(limit-len(buf), len(tmp))

		n, err := reader.Read(tmp[:chunk])
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			return buf, nil
		}

		return nil, err
	}

	return buf, nil
}

// isLikelyText reports whether data is likely plain text.
func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return false
	}

	if utf8.Valid(data) {
		for len(data) > 0 {
			r, size := utf8.DecodeRune(data)
			data = data[size:]

			if r == '\n' || r == '\r' || r == '\t' {
				continue
			}
			if r < 0x20 {
				return false
			}
		}

		return true
	}

	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 0x20 {
			return false
		}
	}

	return true
}
