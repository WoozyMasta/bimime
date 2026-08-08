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

func BenchmarkAnalyzeFast100K(b *testing.B) {
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		for _, sample := range benchmarkSamples {
			benchmarkResultSink = Analyze(sample.name, sample.prefix, ProfileFast)
		}
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
	reader := bytes.NewReader(data)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		reader.Reset(data)
		result, err := AnalyzeReader("model.p3d", reader, ProfileNormal)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResultSink = result
	}
}

func BenchmarkRequiredPrefix(b *testing.B) {
	testCases := []struct {
		label   string
		name    string
		profile Profile
	}{
		{label: "Fast", name: "texture.png", profile: ProfileFast},
		{label: "NormalPathOnly", name: "script.sqf", profile: ProfileNormal},
		{label: "NormalAmbiguous", name: "model.p3d", profile: ProfileNormal},
		{label: "Thorough", name: "texture.png", profile: ProfileThorough},
	}

	for _, testCase := range testCases {
		testCase := testCase
		b.Run(testCase.label, func(b *testing.B) {
			b.ReportAllocs()
			for iteration := 0; iteration < b.N; iteration++ {
				benchmarkCountSink = RequiredPrefix(testCase.name, testCase.profile)
			}
		})
	}
}

func BenchmarkAnalyze(b *testing.B) {
	patternPrefix := make([]byte, semanticPrefixSize)
	copy(patternPrefix, []byte("class CfgModels { class Test {}; };"))
	thoroughPrefix := make([]byte, thoroughPrefixSize)
	copy(thoroughPrefix, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})

	testCases := []struct {
		label   string
		name    string
		prefix  []byte
		profile Profile
	}{
		{label: "Fast", name: "texture.png", profile: ProfileFast},
		{label: "NormalPathOnly", name: "script.sqf", profile: ProfileNormal},
		{label: "NormalRefine4B", name: "model.p3d", prefix: []byte("ODOL"), profile: ProfileNormal},
		{label: "ThoroughBinary", name: "texture.png", prefix: thoroughPrefix, profile: ProfileThorough},
		{label: "ThoroughPattern4K", name: "vehicle.cfg", prefix: patternPrefix, profile: ProfileThorough},
	}

	for _, testCase := range testCases {
		testCase := testCase
		b.Run(testCase.label, func(b *testing.B) {
			b.ReportAllocs()
			for iteration := 0; iteration < b.N; iteration++ {
				benchmarkResultSink = Analyze(testCase.name, testCase.prefix, testCase.profile)
			}
		})
	}
}

func BenchmarkReadPrefix(b *testing.B) {
	data := bytes.Repeat([]byte{'x'}, semanticPrefixSize)
	reader := bytes.NewReader(data)
	for _, size := range []int{0, 4, 16, 156, 4096} {
		b.Run(strconv.Itoa(size)+"B", func(b *testing.B) {
			b.ReportAllocs()
			for iteration := 0; iteration < b.N; iteration++ {
				reader.Reset(data)
				prefix, err := readPrefix(reader, size)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkCountSink = len(prefix)
			}
		})
	}
}

func BenchmarkAnalyzeReader(b *testing.B) {
	testCases := []struct {
		label   string
		name    string
		profile Profile
		data    []byte
	}{
		{label: "4B", name: "model.p3d", profile: ProfileNormal, data: []byte("ODOL")},
		{label: "156B", name: "texture.png", profile: ProfileThorough, data: bytes.Repeat([]byte{'x'}, thoroughPrefixSize)},
		{label: "4096B", name: "vehicle.cfg", profile: ProfileThorough, data: bytes.Repeat([]byte{'x'}, semanticPrefixSize)},
	}

	for _, testCase := range testCases {
		testCase := testCase
		reader := bytes.NewReader(testCase.data)
		b.Run(testCase.label, func(b *testing.B) {
			b.ReportAllocs()
			for iteration := 0; iteration < b.N; iteration++ {
				reader.Reset(testCase.data)
				result, err := AnalyzeReader(testCase.name, reader, testCase.profile)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkResultSink = result
			}
		})
	}
}

// buildBenchmarkSamples creates a deterministic mixed path/content workload.
func buildBenchmarkSamples(count int) []benchmarkSample {
	bases := []benchmarkSample{
		{name: "texture.png", prefix: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}},
		{name: "script.sqf", prefix: []byte("hint \"hello\";")},
		{name: "sound.ogg", prefix: []byte("OggS\x00\x02")},
		{name: "image.jpg", prefix: []byte{0xFF, 0xD8, 0xFF}},
		{name: "model.p3d", prefix: []byte("ODOL")},
		{name: "unknown.xyz", prefix: []byte{0x01, 0x02, 0x03, 0x04}},
	}
	samples := make([]benchmarkSample, count)
	for index := range samples {
		baseIndex := index % 10
		if baseIndex < 8 {
			baseIndex %= 4
		} else {
			baseIndex -= 4
		}
		base := bases[baseIndex]
		samples[index] = benchmarkSample{
			name:   "set/" + strconv.Itoa(index/100) + "/f" + strconv.Itoa(index) + "/" + base.name,
			prefix: base.prefix,
		}
	}

	return samples
}
