// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"bytes"
	"testing"
)

// maxFuzzPrefixSize keeps fuzz cases bounded while exceeding semantic probes.
const maxFuzzPrefixSize = 8192

// FuzzRequiredPrefix checks arbitrary logical paths for panics and unbounded plans.
func FuzzRequiredPrefix(f *testing.F) {
	f.Add("dir/model.p3d", uint8(ProfileNormal))
	f.Add(" texture.png ", uint8(ProfileThorough))
	f.Add("..\\unknown", uint8(ProfileFast))

	f.Fuzz(func(t *testing.T, name string, rawProfile uint8) {
		profile := Profile(rawProfile % 3)
		got := RequiredPrefix(name, profile)
		if got < 0 || got > semanticPrefixSize {
			t.Fatalf("RequiredPrefix(%q, %d) = %d, outside bounds", name, profile, got)
		}
	})
}

// FuzzAnalyze checks arbitrary paths and prefixes for panics and nondeterministic results.
func FuzzAnalyze(f *testing.F) {
	f.Add("script.sqf", []byte("PNG"), uint8(ProfileThorough))
	f.Add("texture.dds", bytes.Repeat([]byte{0xFF}, thoroughPrefixSize), uint8(ProfileNormal))
	f.Add("", []byte{}, uint8(ProfileFast))

	f.Fuzz(func(t *testing.T, name string, prefix []byte, rawProfile uint8) {
		if len(prefix) > maxFuzzPrefixSize {
			prefix = prefix[:maxFuzzPrefixSize]
		}
		profile := Profile(rawProfile % 3)
		first := Analyze(name, prefix, profile)
		second := Analyze(name, prefix, profile)
		if first != second {
			t.Fatalf("Analyze(%q, %d) changed between calls: %+v != %+v", name, profile, first, second)
		}
	})
}

// FuzzAnalyzeReaderEquivalence checks direct and bounded-reader classification agree.
func FuzzAnalyzeReaderEquivalence(f *testing.F) {
	f.Add("model.p3d", []byte("ODOL plus data"), uint8(ProfileNormal))
	f.Add("script.sqf", []byte("\x89PNG\r\n\x1a\n"), uint8(ProfileThorough))
	f.Add("unknown.xyz", []byte("payload"), uint8(ProfileFast))
	latePattern := bytes.Repeat([]byte{' '}, 4500+len("class CfgModels"))
	copy(latePattern[4500:], []byte("class CfgModels"))
	f.Add("vehicle.cfg", latePattern, uint8(ProfileThorough))

	f.Fuzz(func(t *testing.T, name string, input []byte, rawProfile uint8) {
		if len(input) > maxFuzzPrefixSize {
			input = input[:maxFuzzPrefixSize]
		}
		profile := Profile(rawProfile % 3)
		direct := Analyze(name, input, profile)
		fromReader, err := AnalyzeReader(name, bytes.NewReader(input), profile)
		if err != nil {
			t.Fatalf("AnalyzeReader: %v", err)
		}
		if direct != fromReader {
			t.Fatalf("Analyze=%+v, AnalyzeReader=%+v for %q / %d", direct, fromReader, name, profile)
		}
	})
}
