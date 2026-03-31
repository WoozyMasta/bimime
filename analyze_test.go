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

func TestNeedsContent(t *testing.T) {
	t.Parallel()

	if NeedsContent(AnalyzeOptions{
		Path:        "x.rvmat",
		DefaultPlan: PlanFast(),
	}) {
		t.Fatal("NeedsContent(rvmat, extension-only): want false")
	}
	if NeedsContent(AnalyzeOptions{
		Path:        "x.rvmat",
		DefaultPlan: PlanNormal(),
	}) != true {
		t.Fatal("NeedsContent(rvmat, magic-needed): want true")
	}
	if NeedsContent(AnalyzeOptions{
		Path:        "x.bisurf",
		DefaultPlan: PlanNormal(),
	}) != true {
		t.Fatal("NeedsContent(bisurf, magic-needed): want true")
	}
	if NeedsContent(AnalyzeOptions{
		Path:        "x.emat",
		DefaultPlan: PlanNormal(),
	}) {
		t.Fatal("NeedsContent(emat, magic-needed): want false")
	}
	if !NeedsContent(AnalyzeOptions{
		Path:        "x.png",
		DefaultPlan: PlanNormal(),
	}) {
		t.Fatal("NeedsContent(png, magic-needed): want true")
	}
	if !NeedsContent(AnalyzeOptions{
		Path:        "x.unknown",
		DefaultPlan: PlanNormal(),
	}) {
		t.Fatal("NeedsContent(unknown, magic-needed): want true")
	}
	if !NeedsContent(AnalyzeOptions{
		Path:        "x.rvmat",
		DefaultPlan: PlanStrict(),
	}) {
		t.Fatal("NeedsContent(rvmat, strict): want true")
	}
}

func TestAnalyzeFastIgnoresMagic(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "config.rvmat",
		Prefix:      []byte{0x00, 'r', 'a', 'P', 0x00},
		DefaultPlan: PlanFast(),
	})

	if result.Probe.Resolved.ID != "bi.rvmat" {
		t.Fatalf("resolved=%q want bi.rvmat", result.Probe.Resolved.ID)
	}
	if result.Probe.Source != SourceExtension {
		t.Fatalf("source=%q want %q", result.Probe.Source, SourceExtension)
	}
}

func TestAnalyzeNormalUsesMagic(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "texture.png",
		Prefix:      []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
		DefaultPlan: PlanNormal(),
	})

	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
	if result.Probe.ByMagic.ID != "image.png" {
		t.Fatalf("byMagic=%q want image.png", result.Probe.ByMagic.ID)
	}
}

func TestAnalyzeNormalUsesMagicForRVMatByDefault(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "config.rvmat",
		Prefix:      []byte{0x00, 'r', 'a', 'P', 0x00},
		DefaultPlan: PlanNormal(),
	})

	if result.Probe.Resolved.ID != "bi.rvmat.bin" {
		t.Fatalf("resolved=%q want bi.rvmat.bin", result.Probe.Resolved.ID)
	}
	if result.Probe.ByMagic.ID != "bi.rvmat.bin" {
		t.Fatalf("byMagic=%q want bi.rvmat.bin", result.Probe.ByMagic.ID)
	}
}

func TestAnalyzeNormalUsesMagicForBisurfByDefault(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "armor.bisurf",
		Prefix:      []byte{0x00, 'r', 'a', 'P', 0x00},
		DefaultPlan: PlanNormal(),
	})

	if result.Probe.Resolved.ID != "bi.surface.bisurf.bin" {
		t.Fatalf("resolved=%q want bi.surface.bisurf.bin", result.Probe.Resolved.ID)
	}
	if result.Probe.ByMagic.ID != "bi.surface.bisurf.bin" {
		t.Fatalf("byMagic=%q want bi.surface.bisurf.bin", result.Probe.ByMagic.ID)
	}
}

func TestAnalyzeStrictMagicMismatch(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "texture.png",
		Prefix:      []byte("NOTPNG"),
		DefaultPlan: PlanStrict(),
	})

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

	result := Analyze(AnalyzeOptions{
		Path:        "script.sqf",
		Prefix:      []byte{0x00, 0x01, 0x02},
		DefaultPlan: PlanStrict(),
	})

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

func TestAnalyzeStrictWithoutPrefixReportsInsufficientContent(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "script.sqf",
		DefaultPlan: PlanStrict(),
	})

	if result.Valid {
		t.Fatal("strict result must be invalid without content prefix")
	}
	if len(result.Issues) != 1 || result.Issues[0] != AnalyzeIssueInsufficientContent {
		t.Fatalf("issues=%v want [%q]", result.Issues, AnalyzeIssueInsufficientContent)
	}
}

func TestAnalyzeStrictContentPatternMismatch(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "mesh.txo",
		Prefix:      []byte("not a txo payload"),
		DefaultPlan: PlanStrict(),
	})

	if result.Valid {
		t.Fatal("strict result must be invalid on content pattern mismatch")
	}
	if !result.CheckedContentPattern {
		t.Fatal("strict result must check content pattern for txo")
	}
	if len(result.Issues) != 1 || result.Issues[0] != AnalyzeIssueContentPatternMismatch {
		t.Fatalf("issues=%v want [%q]", result.Issues, AnalyzeIssueContentPatternMismatch)
	}
}

func TestAnalyzeStrictContentPatternWithLeadingComments(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path: "mesh.txo",
		Prefix: []byte(
			"// generated file\n\t# note\n  \n\t$object \"x\" {\n}",
		),
		DefaultPlan: PlanStrict(),
	})

	if !result.Valid {
		t.Fatalf("strict result must be valid, issues=%v", result.Issues)
	}
	if !result.CheckedContentPattern {
		t.Fatal("strict result must check content pattern for txo")
	}
}

func TestAnalyzeReaderSkipsReadWhenNotNeeded(t *testing.T) {
	t.Parallel()

	reader := &countingReader{data: []byte("class CfgFoo {}")}
	result, err := AnalyzeReader(reader, AnalyzeOptions{
		Path:        "x.sqf",
		DefaultPlan: PlanNormal(),
	})
	if err != nil {
		t.Fatalf("AnalyzeReader(sqf): %v", err)
	}
	if reader.reads != 0 {
		t.Fatalf("reads=%d want 0", reader.reads)
	}
	if result.Probe.Resolved.ID != "text.sqf" {
		t.Fatalf("resolved=%q want text.sqf", result.Probe.Resolved.ID)
	}
}

func TestAnalyzeReaderReadsWhenNeeded(t *testing.T) {
	t.Parallel()

	reader := &countingReader{
		data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
	}
	result, err := AnalyzeReader(reader, AnalyzeOptions{
		Path:        "x.png",
		DefaultPlan: PlanNormal(),
	})
	if err != nil {
		t.Fatalf("AnalyzeReader(png): %v", err)
	}
	if reader.reads == 0 {
		t.Fatal("reader must be consumed for png")
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

	result, err := AnalyzeFile(AnalyzeOptions{
		Path:        filePath,
		DefaultPlan: PlanNormal(),
	})
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}
	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
}

func TestNeedsContentFastWithWRPMagicOverride(t *testing.T) {
	t.Parallel()

	if NeedsContent(AnalyzeOptions{
		Path:        "terrain.wrp",
		DefaultPlan: PlanFast(),
		PlansByExtension: map[string]AnalyzePlan{
			"wrp": {
				Match: AnalyzeMatchExtensionMagic,
			},
		},
	}) != true {
		t.Fatal("NeedsContent(wrp override): want true")
	}
}

func TestAnalyzeReaderFastWithWRPMagicOverride(t *testing.T) {
	t.Parallel()

	reader := &countingReader{data: []byte("9VBW\x00\x00")}
	result, err := AnalyzeReader(reader, AnalyzeOptions{
		Path:        "terrain.wrp",
		DefaultPlan: PlanFast(),
		PlansByExtension: map[string]AnalyzePlan{
			"wrp": {
				Match: AnalyzeMatchExtensionMagic,
			},
		},
	})
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if reader.reads == 0 {
		t.Fatal("reader must be consumed for wrp magic override")
	}
	if result.Probe.Resolved.ID != "bi.wrp.9vbw" {
		t.Fatalf("resolved=%q want bi.wrp.9vbw", result.Probe.Resolved.ID)
	}
}

func TestAnalyzeNormalWithStrictOverrideByExtension(t *testing.T) {
	t.Parallel()

	result := Analyze(AnalyzeOptions{
		Path:        "script.sqf",
		Prefix:      []byte{0x00, 0x01, 0x02},
		DefaultPlan: PlanNormal(),
		PlansByExtension: map[string]AnalyzePlan{
			"sqf": {
				Match:    AnalyzeMatchExtensionMagicNeeded,
				Validate: AnalyzeValidateStrict,
			},
		},
	})

	if result.Valid {
		t.Fatal("result must be invalid for strict sqf override")
	}
	if !result.CheckedText {
		t.Fatal("result must check text in strict sqf override")
	}
	if len(result.Issues) != 1 || result.Issues[0] != AnalyzeIssueTextExpected {
		t.Fatalf("issues=%v want [%q]", result.Issues, AnalyzeIssueTextExpected)
	}
}

func TestAnalyzerReuse(t *testing.T) {
	t.Parallel()

	analyzer := NewAnalyzer(AnalyzeOptions{
		DefaultPlan: PlanNormal(),
	})
	if !analyzer.NeedsContent("x.png") {
		t.Fatal("analyzer.NeedsContent(png): want true")
	}

	result := analyzer.Analyze("texture.png", []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
	})
	if result.Probe.Resolved.ID != "image.png" {
		t.Fatalf("resolved=%q want image.png", result.Probe.Resolved.ID)
	}
}
