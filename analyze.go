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

// defaultMagicNeededExtensions marks ambiguous extension-only formats that
// should read content in normal mode to disambiguate source vs binary payloads.
var defaultMagicNeededExtensions = map[string]struct{}{
	"p3d":    {},
	"wrp":    {},
	"rvmat":  {},
	"bisurf": {},
}

// Analyzer stores normalized analyze configuration for repeated calls.
type Analyzer struct {
	defaultPlan      AnalyzePlan
	plansByExtension map[string]AnalyzePlan
	prefixSize       int
}

// NewAnalyzer builds analyzer from options and normalizes plan declarations.
func NewAnalyzer(options AnalyzeOptions) Analyzer {
	options = normalizeAnalyzeOptions(options)

	return Analyzer{
		defaultPlan:      options.DefaultPlan,
		plansByExtension: options.PlansByExtension,
		prefixSize:       options.PrefixSize,
	}
}

// NeedsContent reports whether selected plan requires payload prefix bytes.
func NeedsContent(options AnalyzeOptions) bool {
	analyzer := NewAnalyzer(options)
	return analyzer.NeedsContent(options.Path)
}

// HasMagic reports whether a registered type has at least one magic signature.
func HasMagic(typeID string) bool {
	return hasMagicForType(typeID)
}

// NeedsContent reports whether payload prefix is required for this path.
func (analyzer Analyzer) NeedsContent(path string) bool {
	path = trimPathSpace(path)
	plan := analyzer.effectivePlan(path)

	return needsContentForPlan(path, plan)
}

// Analyze classifies path/prefix pair according to analyzer plan.
func (analyzer Analyzer) Analyze(path string, prefix []byte) AnalyzeResult {
	path = trimPathSpace(path)
	plan := analyzer.effectivePlan(path)
	probe := probeForPlan(path, prefix, plan)

	result := AnalyzeResult{
		Plan:  plan,
		Probe: probe,
		Valid: true,
	}

	if plan.Validate != AnalyzeValidateStrict {
		return result
	}

	validateStrict(&result, prefix)
	return result
}

// Analyze classifies path/prefix pair according to extension-aware plans.
func Analyze(options AnalyzeOptions) AnalyzeResult {
	analyzer := NewAnalyzer(options)
	return analyzer.Analyze(options.Path, options.Prefix)
}

// AnalyzeReader classifies path using analyzer and optional reader.
func (analyzer Analyzer) AnalyzeReader(path string, reader io.Reader) (AnalyzeResult, error) {
	path = trimPathSpace(path)
	if !analyzer.NeedsContent(path) {
		return analyzer.Analyze(path, nil), nil
	}
	if reader == nil {
		return AnalyzeResult{}, ErrNilReader
	}

	prefix, err := readPrefix(reader, analyzer.prefixSize)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("read content prefix: %w", err)
	}

	return analyzer.Analyze(path, prefix), nil
}

// AnalyzeReader classifies path using selected plans and optional reader.
func AnalyzeReader(reader io.Reader, options AnalyzeOptions) (AnalyzeResult, error) {
	analyzer := NewAnalyzer(options)
	return analyzer.AnalyzeReader(options.Path, reader)
}

// AnalyzeFile classifies file path and reads only required prefix bytes.
func (analyzer Analyzer) AnalyzeFile(path string) (AnalyzeResult, error) {
	path = trimPathSpace(path)
	if !analyzer.NeedsContent(path) {
		return analyzer.Analyze(path, nil), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return analyzer.AnalyzeReader(path, f)
}

// AnalyzeFile classifies file path and reads only required prefix bytes.
func AnalyzeFile(options AnalyzeOptions) (AnalyzeResult, error) {
	analyzer := NewAnalyzer(options)
	return analyzer.AnalyzeFile(options.Path)
}

// normalizeAnalyzeOptions fills defaults and normalizes plan declarations.
func normalizeAnalyzeOptions(options AnalyzeOptions) AnalyzeOptions {
	if options.PrefixSize <= 0 {
		options.PrefixSize = defaultAnalyzePrefixSize
	}

	options.DefaultPlan = normalizeAnalyzePlan(options.DefaultPlan)
	options.PlansByExtension = normalizeExtensionPlans(options.PlansByExtension)
	options.Path = trimPathSpace(options.Path)

	return options
}

// normalizeAnalyzePlan normalizes zero/default values and invalid mixes.
func normalizeAnalyzePlan(plan AnalyzePlan) AnalyzePlan {
	if plan.Match == AnalyzeMatchDefault {
		plan.Match = AnalyzeMatchExtensionMagicNeeded
	}
	if plan.Validate == AnalyzeValidateDefault {
		plan.Validate = AnalyzeValidateNone
	}
	if plan.Validate == AnalyzeValidateStrict && plan.Match == AnalyzeMatchExtension {
		plan.Match = AnalyzeMatchExtensionMagicNeeded
	}

	return plan
}

// normalizeExtensionPlans normalizes extension keys and plan values.
func normalizeExtensionPlans(source map[string]AnalyzePlan) map[string]AnalyzePlan {
	if len(source) == 0 {
		return nil
	}

	out := make(map[string]AnalyzePlan, len(source))
	for ext, plan := range source {
		key := lowerKey(strings.TrimPrefix(strings.TrimSpace(ext), "."))
		if key == "" {
			continue
		}

		out[key] = normalizeAnalyzePlan(plan)
	}

	return out
}

// effectivePlanForPath returns extension-specific plan or normalized default.
func (analyzer Analyzer) effectivePlan(path string) AnalyzePlan {
	if len(analyzer.plansByExtension) == 0 {
		return analyzer.defaultPlan
	}

	ext := extensionKey(path)
	if ext != "" {
		if plan, ok := analyzer.plansByExtension[ext]; ok {
			return plan
		}
	}

	return analyzer.defaultPlan
}

// needsContentForPlan reports whether plan requires payload prefix reads.
func needsContentForPlan(path string, plan AnalyzePlan) bool {
	if plan.Validate == AnalyzeValidateStrict {
		return true
	}

	switch plan.Match {
	case AnalyzeMatchExtension:
		return false
	case AnalyzeMatchExtensionMagic:
		return true
	case AnalyzeMatchExtensionMagicNeeded:
		extension := extensionKey(path)
		if _, ok := defaultMagicNeededExtensions[extension]; ok {
			return true
		}

		byExtension, ok := detectByPathRecord(path)
		if !ok {
			return true
		}

		return len(byExtension.magic) > 0
	default:
		return true
	}
}

// probeForPlan picks extension-only or extension+magic probing by plan.
func probeForPlan(path string, prefix []byte, plan AnalyzePlan) ProbeResult {
	if plan.Match == AnalyzeMatchExtension {
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
			Resolved:    UnknownType(),
			ByMagic:     UnknownType(),
			ByExtension: UnknownType(),
		}
	}

	byExtensionType := probeType(byExtension.typ)

	return ProbeResult{
		Source:      SourceExtension,
		Extension:   extension,
		Resolved:    byExtensionType,
		ByMagic:     UnknownType(),
		ByExtension: byExtensionType,
	}
}

// validateStrict validates extension/magic consistency and text-like payloads.
func validateStrict(result *AnalyzeResult, prefix []byte) {
	if result == nil {
		return
	}
	if len(prefix) == 0 {
		result.Valid = false
		result.Issues = append(result.Issues, AnalyzeIssueInsufficientContent)
		return
	}

	if result.Probe.ByExtension.ID != UnknownType().ID && hasMagicForType(result.Probe.ByExtension.ID) {
		result.CheckedMagic = true
		if result.Probe.ByMagic.ID != result.Probe.ByExtension.ID {
			result.Valid = false
			result.Issues = append(result.Issues, AnalyzeIssueMagicMismatch)
		}
	}
	if hasContentPatternForType(result.Probe.Resolved.ID) {
		result.CheckedContentPattern = true
		if !matchContentPatternForType(result.Probe.Resolved.ID, prefix) {
			result.Valid = false
			result.Issues = append(result.Issues, AnalyzeIssueContentPatternMismatch)
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

// hasMagicForType reports whether registry record has at least one signature.
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

// hasContentPatternForType reports whether type has at least one content regex.
func hasContentPatternForType(typeID string) bool {
	record, ok := typeByID[typeID]
	if ok {
		return len(record.contentPatterns) > 0
	}

	key := strings.ToLower(strings.TrimSpace(typeID))
	record, ok = typeByID[key]
	if !ok {
		return false
	}

	return len(record.contentPatterns) > 0
}

// matchContentPatternForType reports whether payload matches any configured regex.
func matchContentPatternForType(typeID string, prefix []byte) bool {
	record, ok := typeByID[typeID]
	if !ok {
		key := strings.ToLower(strings.TrimSpace(typeID))
		record, ok = typeByID[key]
		if !ok {
			return false
		}
	}
	if len(record.contentPatterns) == 0 {
		return true
	}
	for _, pattern := range record.contentPatterns {
		if pattern.Match(prefix) {
			return true
		}
	}

	return false
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
