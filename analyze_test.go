package bimime

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

type countingReader struct {
	data  []byte
	index int
	reads int
}

// Read tracks read calls while serving predefined data.
func (r *countingReader) Read(p []byte) (int, error) {
	r.reads++
	if r.index >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.index:])
	r.index += n
	return n, nil
}

func TestParseDetectMode(t *testing.T) {
	t.Parallel()

	mode, err := ParseDetectMode(" strict ")
	if err != nil {
		t.Fatalf("ParseDetectMode(strict): %v", err)
	}
	if mode != DetectModeStrict {
		t.Fatalf("mode=%q want %q", mode, DetectModeStrict)
	}

	if _, err := ParseDetectMode("turbo"); err == nil {
		t.Fatal("ParseDetectMode(turbo): expected error")
	}
}

func TestNeedsContent(t *testing.T) {
	t.Parallel()

	if NeedsContent("x.rvmat", DetectModeFast) {
		t.Fatal("NeedsContent(rvmat, fast): want false")
	}
	if NeedsContent("x.rvmat", DetectModeNormal) {
		t.Fatal("NeedsContent(rvmat, normal): want false")
	}
	if !NeedsContent("x.png", DetectModeNormal) {
		t.Fatal("NeedsContent(png, normal): want true")
	}
	if !NeedsContent("x.unknown", DetectModeNormal) {
		t.Fatal("NeedsContent(unknown, normal): want true")
	}
	if !NeedsContent("x.rvmat", DetectModeStrict) {
		t.Fatal("NeedsContent(rvmat, strict): want true")
	}
}

func TestAnalyzeFastIgnoresMagic(t *testing.T) {
	t.Parallel()

	result := Analyze(
		"config.rvmat",
		[]byte{0x00, 'r', 'a', 'P', 0x00},
		AnalyzeOptions{Mode: DetectModeFast},
	)

	if result.Probe.Resolved.ID != "bi.rvmat" {
		t.Fatalf("resolved=%q want bi.rvmat", result.Probe.Resolved.ID)
	}
	if result.Probe.Source != SourceExtension {
		t.Fatalf("source=%q want %q", result.Probe.Source, SourceExtension)
	}
}

func TestAnalyzeNormalUsesMagic(t *testing.T) {
	t.Parallel()

	result := Analyze(
		"texture.png",
		[]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
		AnalyzeOptions{Mode: DetectModeNormal},
	)

	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
	if result.Probe.ByMagic.ID != "image.png" {
		t.Fatalf("byMagic=%q want image.png", result.Probe.ByMagic.ID)
	}
}

func TestAnalyzeStrictMagicMismatch(t *testing.T) {
	t.Parallel()

	result := Analyze(
		"texture.png",
		[]byte("NOTPNG"),
		AnalyzeOptions{Mode: DetectModeStrict},
	)

	if result.Valid {
		t.Fatal("strict result must be invalid on magic mismatch")
	}
	if !result.CheckedMagic {
		t.Fatal("strict result must check magic for png")
	}
	if len(result.Issues) != 1 || result.Issues[0] != AnalyzeIssueMagicMismatch {
		t.Fatalf("issues=%v want [%q]", result.Issues, AnalyzeIssueMagicMismatch)
	}
}

func TestAnalyzeStrictTextValidation(t *testing.T) {
	t.Parallel()

	result := Analyze(
		"script.sqf",
		[]byte{0x00, 0x01, 0x02},
		AnalyzeOptions{Mode: DetectModeStrict},
	)

	if result.Valid {
		t.Fatal("strict result must be invalid for binary-like text payload")
	}
	if !result.CheckedText {
		t.Fatal("strict result must check text for sqf")
	}
	if result.LooksText {
		t.Fatal("strict result must mark binary-like sample as not text")
	}
	if len(result.Issues) != 1 || result.Issues[0] != AnalyzeIssueTextExpected {
		t.Fatalf("issues=%v want [%q]", result.Issues, AnalyzeIssueTextExpected)
	}
}

func TestAnalyzeReaderSkipsReadWhenNotNeeded(t *testing.T) {
	t.Parallel()

	reader := &countingReader{data: []byte{0x00, 'r', 'a', 'P'}}
	result, err := AnalyzeReader("x.rvmat", reader, AnalyzeOptions{Mode: DetectModeNormal})
	if err != nil {
		t.Fatalf("AnalyzeReader(rvmat, normal): %v", err)
	}
	if reader.reads != 0 {
		t.Fatalf("reads=%d want 0", reader.reads)
	}
	if result.Probe.Resolved.ID != "bi.rvmat" {
		t.Fatalf("resolved=%q want bi.rvmat", result.Probe.Resolved.ID)
	}
}

func TestAnalyzeReaderReadsWhenNeeded(t *testing.T) {
	t.Parallel()

	reader := &countingReader{
		data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
	}
	result, err := AnalyzeReader("x.png", reader, AnalyzeOptions{Mode: DetectModeNormal})
	if err != nil {
		t.Fatalf("AnalyzeReader(png, normal): %v", err)
	}
	if reader.reads == 0 {
		t.Fatal("reader must be consumed for png in normal mode")
	}
	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
}

func TestAnalyzeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(filePath, []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
	}, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result, err := AnalyzeFile(filePath, AnalyzeOptions{Mode: DetectModeNormal})
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}
	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
}
