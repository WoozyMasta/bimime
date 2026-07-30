// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// countingReader tracks bytes and calls while serving predefined input.
type countingReader struct {
	// data is the input returned by Read.
	data []byte
	// offset is the next unread byte.
	offset int
	// calls counts calls to Read.
	calls int
}

// Read serves input and records the amount consumed.
func (reader *countingReader) Read(buffer []byte) (int, error) {
	reader.calls++
	if reader.offset >= len(reader.data) {
		return 0, io.EOF
	}

	n := copy(buffer, reader.data[reader.offset:])
	reader.offset += n

	return n, nil
}

// noProgressReader simulates a broken reader that never returns bytes or an error.
type noProgressReader struct{}

// Read reports no progress on every call.
func (noProgressReader) Read([]byte) (int, error) {
	return 0, nil
}

func TestRequiredPrefixProfiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		profile Profile
		want    int
	}{
		{name: "x.p3d", profile: ProfileFast, want: 0},
		{name: "x.p3d", profile: ProfileNormal, want: 4},
		{name: "x.wrp", profile: ProfileNormal, want: 4},
		{name: "x.rvmat", profile: ProfileNormal, want: 4},
		{name: "x.bisurf", profile: ProfileNormal, want: 4},
		{name: "x.png", profile: ProfileNormal, want: 0},
		{name: "x.sqf", profile: ProfileNormal, want: 0},
		{name: "x.unknown", profile: ProfileNormal, want: 0},
		{name: "x.png", profile: ProfileThorough, want: thoroughPrefixSize},
		{name: "x.cfg", profile: ProfileThorough, want: semanticPrefixSize},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name+"/"+string(rune('0'+testCase.profile)), func(t *testing.T) {
			t.Parallel()
			if got := RequiredPrefix(testCase.name, testCase.profile); got != testCase.want {
				t.Fatalf("RequiredPrefix(%q, %d) = %d, want %d", testCase.name, testCase.profile, got, testCase.want)
			}
		})
	}
}

func TestAnalyzeHonorsProfileAndRelations(t *testing.T) {
	t.Parallel()

	png := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	fast := Analyze("script.sqf", png, ProfileFast)
	if fast.Type != "text.sqf" || fast.Relation != RelationPathOnly {
		t.Fatalf("fast result = %+v, want path-only SQF", fast)
	}

	normal := Analyze("script.sqf", png, ProfileNormal)
	if normal.Type != "text.sqf" || normal.ByContent != "" || normal.Relation != RelationPathOnly {
		t.Fatalf("normal result = %+v, want path-only SQF", normal)
	}

	thorough := Analyze("script.sqf", png, ProfileThorough)
	if thorough.Type != "image.png" || thorough.ByPath != "text.sqf" ||
		thorough.ByContent != "image.png" || thorough.Relation != RelationConflict {
		t.Fatalf("thorough result = %+v, want SQF/PNG conflict", thorough)
	}
	if !thorough.Has(FlagIncomplete) {
		t.Fatal("short thorough prefix must be marked incomplete")
	}

	refined := Analyze("model.p3d", []byte("ODOL"), ProfileNormal)
	if refined.Type != "bi.p3d.odol" || refined.ByPath != "bi.p3d" ||
		refined.ByContent != "bi.p3d.odol" || refined.Relation != RelationRefined {
		t.Fatalf("P3D result = %+v, want ODOL refinement", refined)
	}

	rvmat := Analyze("material.rvmat", []byte{0x00, 'r', 'a', 'P'}, ProfileNormal)
	if rvmat.Type != "bi.rvmat.bin" || rvmat.Relation != RelationRefined {
		t.Fatalf("RVMat result = %+v, want RAP refinement", rvmat)
	}
}

func TestAnalyzePatternRefinement(t *testing.T) {
	t.Parallel()

	result := Analyze("vehicle.cfg", []byte("class CfgModels { class Test {}; };"), ProfileThorough)
	if result.Type != "bi.model.cfg" || result.ByPath != "text.cfg" ||
		result.ByContent != "bi.model.cfg" || result.ContentSource != ContentSourcePattern ||
		result.Relation != RelationRefined {
		t.Fatalf("result = %+v, want model config refinement", result)
	}
	if !result.Has(FlagIncomplete) {
		t.Fatal("short pattern sample must be marked incomplete")
	}
}

func TestAnalyzeReaderBoundsAndNoProgress(t *testing.T) {
	t.Parallel()

	reader := &countingReader{data: []byte("ODOL extra bytes")}
	result, err := AnalyzeReader("model.p3d", reader, ProfileNormal)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if reader.offset != 4 || result.Type != "bi.p3d.odol" {
		t.Fatalf("read %d bytes, result %+v; want 4 bytes and ODOL", reader.offset, result)
	}

	unused := &countingReader{data: []byte("ignored")}
	if _, err := AnalyzeReader("script.sqf", unused, ProfileNormal); err != nil {
		t.Fatalf("AnalyzeReader without planned read: %v", err)
	}
	if unused.calls != 0 {
		t.Fatalf("reader calls = %d, want 0", unused.calls)
	}
	if _, err := AnalyzeReader("script.sqf", nil, ProfileNormal); err != nil {
		t.Fatalf("nil reader with zero requirement: %v", err)
	}

	if _, err := AnalyzeReader("x.p3d", noProgressReader{}, ProfileNormal); !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("no-progress error = %v, want io.ErrNoProgress", err)
	}
}

func TestAnalyzeReaderShortInputAndFileFastPath(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeReader("model.p3d", strings.NewReader("OD"), ProfileNormal)
	if err != nil {
		t.Fatalf("short reader: %v", err)
	}
	if result.Type != "bi.p3d" || !result.Has(FlagIncomplete) {
		t.Fatalf("short result = %+v, want incomplete generic P3D", result)
	}

	result, err = AnalyzeFile("missing.png", ProfileNormal)
	if err != nil {
		t.Fatalf("zero-read file fast path: %v", err)
	}
	if result.Type != "image.png" || result.RequiredPrefix != 0 {
		t.Fatalf("file fast-path result = %+v", result)
	}
}

func TestAnalyzePreservesPathSpacesAndSeparators(t *testing.T) {
	t.Parallel()

	if got := Analyze("dir\\Texture.PNG", nil, ProfileFast).Type; got != "image.png" {
		t.Fatalf("Windows-style path type = %q, want image.png", got)
	}
	if got := Analyze("dir/Texture.PNG", nil, ProfileFast).Type; got != "image.png" {
		t.Fatalf("slash path type = %q, want image.png", got)
	}
	if got := Analyze("Texture.PNG ", nil, ProfileFast).Type; got != UnknownID {
		t.Fatalf("space-suffixed name type = %q, want unknown", got)
	}
}

func TestAnalyzeRejectsInvalidProfile(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("invalid profile must panic")
		}
	}()
	Analyze("x.png", nil, Profile(255))
}
