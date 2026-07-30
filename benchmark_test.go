// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"bytes"
	"strconv"
	"testing"
)

// benchmarkSample stores a logical name and a reusable content prefix.
type benchmarkSample struct {
	// name is the path supplied to the classifier.
	name string
	// prefix is caller-owned content available to the classifier.
	prefix []byte
}

// benchmarkSamples provides a mixed 100k-entry workload.
var benchmarkSamples = buildBenchmarkSamples(100_000)

// benchmarkResultSink keeps benchmark results observable.
var benchmarkResultSink Result

// benchmarkCountSink keeps aggregate benchmark values observable.
var benchmarkCountSink int

func BenchmarkRequiredPrefixNormal100K(b *testing.B) {
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		count := 0
		for _, sample := range benchmarkSamples {
			if RequiredPrefix(sample.name, ProfileNormal) > 0 {
				count++
			}
		}
		benchmarkCountSink = count
	}
}

func BenchmarkAnalyzeNormal100K(b *testing.B) {
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		for _, sample := range benchmarkSamples {
			benchmarkResultSink = Analyze(sample.name, sample.prefix, ProfileNormal)
		}
	}
}

func BenchmarkAnalyzeThorough100K(b *testing.B) {
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		for _, sample := range benchmarkSamples {
			benchmarkResultSink = Analyze(sample.name, sample.prefix, ProfileThorough)
		}
	}
}

func BenchmarkAnalyzeReaderP3D(b *testing.B) {
	data := []byte("ODOL plus trailing data")
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		result, err := AnalyzeReader("model.p3d", bytes.NewReader(data), ProfileNormal)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResultSink = result
	}
}

// buildBenchmarkSamples creates a deterministic mixed path/content workload.
func buildBenchmarkSamples(count int) []benchmarkSample {
	bases := []benchmarkSample{
		{name: "texture.png", prefix: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}},
		{name: "model.p3d", prefix: []byte("ODOL")},
		{name: "terrain.wrp", prefix: []byte("OPRW")},
		{name: "script.sqf", prefix: []byte("hint \"hello\";")},
		{name: "unknown.xyz", prefix: []byte{0x01, 0x02, 0x03, 0x04}},
	}
	samples := make([]benchmarkSample, count)
	for index := range samples {
		base := bases[index%len(bases)]
		samples[index] = benchmarkSample{
			name:   "set/" + strconv.Itoa(index/100) + "/f" + strconv.Itoa(index) + "/" + base.name,
			prefix: base.prefix,
		}
	}

	return samples
}
