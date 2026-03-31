package bimime

import (
	"bytes"
	"strconv"
	"testing"
)

// benchSample stores one synthetic file path with optional content prefix.
type benchSample struct {
	path   string
	prefix []byte
}

// benchmarkDatasets are reused across benchmark runs.
var benchmarkDatasets = map[int][]benchSample{
	128:  buildBenchDataset(128),
	1024: buildBenchDataset(1024),
	4096: buildBenchDataset(4096),
}

// benchAnalyzeSink keeps benchmark results from being optimized away.
var benchAnalyzeSink AnalyzeResult

// benchCountSink keeps aggregate counters from being optimized away.
var benchCountSink int

// baseBenchSamples is a mixed list of text, binary, and ambiguous formats.
var baseBenchSamples = []benchSample{
	{path: "config.rvmat", prefix: []byte("class StageTI {};")},
	{path: "config.bin", prefix: []byte{0x00, 'r', 'a', 'P', 0x00}},
	{path: "texture.edds", prefix: []byte("DDS ")},
	{path: "texture.png", prefix: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}},
	{path: "texture.dds", prefix: []byte("DDS xxxx")},
	{path: "model.p3d", prefix: []byte("ODOL")},
	{path: "terrain.wrp", prefix: []byte("OPRW")},
	{path: "script.sqf", prefix: []byte("hint \"hello\";")},
	{path: "script.sqfc", prefix: []byte{0x00, 0x01, 0x7F, 0x10}},
	{path: "anim.anm", prefix: []byte{0x01, 0x02, 0x03}},
	{path: "sound.ogg", prefix: []byte("OggS\x00\x02")},
	{path: "music.mp3", prefix: []byte("ID3\x04\x00\x00")},
	{path: "video.mp4", prefix: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}},
	{path: "ui.layout", prefix: []byte("<WidgetLayout/>")},
	{path: "table.xlsx", prefix: []byte{0x50, 0x4B, 0x03, 0x04}},
	{path: "project.gproj", prefix: []byte("<Project/>")},
	{path: "unknown.xyz", prefix: []byte{0x01, 0x02, 0x03, 0x04}},
}

// BenchmarkAnalyzeBatch measures Analyze throughput on batch workloads.
func BenchmarkAnalyzeBatch(b *testing.B) {
	b.Run("Fast_128", func(b *testing.B) {
		benchmarkAnalyzeBatch(b, benchmarkDatasets[128], AnalyzeOptions{
			DefaultPlan: PlanFast(),
		})
	})
	b.Run("Normal_1024", func(b *testing.B) {
		benchmarkAnalyzeBatch(b, benchmarkDatasets[1024], AnalyzeOptions{
			DefaultPlan: PlanNormal(),
		})
	})
	b.Run("Strict_1024", func(b *testing.B) {
		benchmarkAnalyzeBatch(b, benchmarkDatasets[1024], AnalyzeOptions{
			DefaultPlan: PlanStrict(),
		})
	})
	b.Run("HybridFastMagic_1024", func(b *testing.B) {
		benchmarkAnalyzeBatch(b, benchmarkDatasets[1024], AnalyzeOptions{
			DefaultPlan:      PlanFast(),
			PlansByExtension: BIAmbiguousRAPOverrides(),
		})
	})
	b.Run("Normal_4096", func(b *testing.B) {
		benchmarkAnalyzeBatch(b, benchmarkDatasets[4096], AnalyzeOptions{
			DefaultPlan: PlanNormal(),
		})
	})
}

// BenchmarkAnalyzeReaderBatch measures AnalyzeReader throughput on batch reads.
func BenchmarkAnalyzeReaderBatch(b *testing.B) {
	b.Run("Normal_1024", func(b *testing.B) {
		benchmarkAnalyzeReaderBatch(b, benchmarkDatasets[1024], AnalyzeOptions{
			DefaultPlan: PlanNormal(),
		})
	})
	b.Run("Strict_1024", func(b *testing.B) {
		benchmarkAnalyzeReaderBatch(b, benchmarkDatasets[1024], AnalyzeOptions{
			DefaultPlan: PlanStrict(),
		})
	})
}

// BenchmarkNeedsContentBatch measures fast pre-check for read/no-read decision.
func BenchmarkNeedsContentBatch(b *testing.B) {
	dataset := benchmarkDatasets[4096]
	options := AnalyzeOptions{
		DefaultPlan: PlanNormal(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0

		for _, sample := range dataset {
			options.Path = sample.path
			if NeedsContent(options) {
				count++
			}
		}

		benchCountSink = count
	}
}

// buildBenchDataset builds a deterministic synthetic dataset.
func buildBenchDataset(count int) []benchSample {
	out := make([]benchSample, 0, count)

	for i := 0; i < count; i++ {
		base := baseBenchSamples[i%len(baseBenchSamples)]
		clone := benchSample{
			path:   "set/" + strconv.Itoa(i/100) + "/f" + strconv.Itoa(i) + "_" + base.path,
			prefix: append([]byte(nil), base.prefix...),
		}
		out = append(out, clone)
	}

	return out
}

// benchmarkAnalyzeBatch runs Analyze for all entries in dataset.
func benchmarkAnalyzeBatch(b *testing.B, dataset []benchSample, options AnalyzeOptions) {
	base := options

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid := 0

		for _, sample := range dataset {
			base.Path = sample.path
			base.Prefix = sample.prefix
			result := Analyze(base)
			if result.Valid {
				valid++
			}

			benchAnalyzeSink = result
		}

		benchCountSink = valid
	}
}

// benchmarkAnalyzeReaderBatch runs AnalyzeReader for all entries in dataset.
func benchmarkAnalyzeReaderBatch(
	b *testing.B,
	dataset []benchSample,
	options AnalyzeOptions,
) {
	base := options

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid := 0

		for _, sample := range dataset {
			base.Path = sample.path
			base.Prefix = nil
			result, err := AnalyzeReader(
				bytes.NewReader(sample.prefix),
				base,
			)
			if err != nil {
				b.Fatalf("AnalyzeReader(%s): %v", sample.path, err)
			}
			if result.Valid {
				valid++
			}

			benchAnalyzeSink = result
		}

		benchCountSink = valid
	}
}
