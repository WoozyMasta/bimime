// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// noProgressReader simulates a broken reader that never returns bytes or an error.
type noProgressReader struct{}

// intermittentReader returns empty reads before serving its input.
type intermittentReader struct {
	// data is the input returned after the empty reads.
	data []byte
	// zeroReads counts empty reads already returned.
	zeroReads int
}

// countingReader tracks bytes and calls while serving predefined input.
type countingReader struct {
	// data is the input returned by Read.
	data []byte
	// offset is the next unread byte.
	offset int
	// calls counts calls to Read.
	calls int
}

// oneByteReader returns no more than one byte per call.
type oneByteReader struct {
	// data is the remaining input.
	data []byte
	// calls counts reads performed by the analyzer.
	calls int
}

// dataEOFReader returns its remaining bytes and io.EOF in one call.
type dataEOFReader struct {
	// data is the remaining input.
	data []byte
}

// partialErrorReader returns its remaining bytes together with a read error.
type partialErrorReader struct {
	// data is the remaining input.
	data []byte
	// err is returned with the final bytes.
	err error
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

// Read serves at most one byte from the remaining input.
func (reader *oneByteReader) Read(buffer []byte) (int, error) {
	reader.calls++
	if len(reader.data) == 0 {
		return 0, io.EOF
	}

	n := copy(buffer[:min(len(buffer), 1)], reader.data)
	reader.data = reader.data[n:]

	return n, nil
}

// Read returns the remaining bytes and EOF, following the io.Reader contract.
func (reader *dataEOFReader) Read(buffer []byte) (int, error) {
	n := copy(buffer, reader.data)
	reader.data = reader.data[n:]

	return n, io.EOF
}

// Read returns the remaining bytes and the configured error.
func (reader *partialErrorReader) Read(buffer []byte) (int, error) {
	n := copy(buffer, reader.data)
	reader.data = reader.data[n:]

	return n, reader.err
}

// Read reports no progress on every call.
func (noProgressReader) Read([]byte) (int, error) {
	return 0, nil
}

// Read returns two temporary empty reads before serving the configured input.
func (reader *intermittentReader) Read(buffer []byte) (int, error) {
	if reader.zeroReads < 2 {
		reader.zeroReads++
		return 0, nil
	}

	n := copy(buffer, reader.data)
	reader.data = reader.data[n:]
	if len(reader.data) == 0 {
		return n, io.EOF
	}

	return n, nil
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

	normalOverread := Analyze("model.p3d", png, ProfileNormal)
	if normalOverread.Type != "bi.p3d" || normalOverread.ByContent != "" ||
		normalOverread.Relation != RelationPathOnly {
		t.Fatalf("normal over-read result = %+v, want path-only P3D", normalOverread)
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

func TestAnalyzeReaderOneByteReads(t *testing.T) {
	t.Parallel()

	reader := &oneByteReader{data: []byte("ODOL trailing data")}
	result, err := AnalyzeReader("model.p3d", reader, ProfileNormal)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if result.Type != "bi.p3d.odol" || reader.calls != 4 || string(reader.data) != " trailing data" {
		t.Fatalf("result = %+v, calls = %d, remaining = %q; want ODOL after four one-byte reads", result, reader.calls, reader.data)
	}
}

func TestAnalyzeReaderEOFBehavior(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{name: "immediate EOF", reader: strings.NewReader("")},
		{name: "data with EOF", reader: &dataEOFReader{data: []byte("OD")}},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := AnalyzeReader("model.p3d", testCase.reader, ProfileNormal)
			if err != nil {
				t.Fatalf("AnalyzeReader: %v", err)
			}
			if result.Type != "bi.p3d" || !result.Has(FlagIncomplete) {
				t.Fatalf("result = %+v, want incomplete generic P3D", result)
			}
		})
	}
}

func TestAnalyzeReaderReturnsErrorAfterPartialData(t *testing.T) {
	t.Parallel()

	readErr := errors.New("partial read failure")
	reader := &partialErrorReader{data: []byte("OD"), err: readErr}
	_, err := AnalyzeReader("model.p3d", reader, ProfileNormal)
	if !errors.Is(err, readErr) {
		t.Fatalf("AnalyzeReader error = %v, want partial read failure", err)
	}
}

func TestAnalyzeReaderFastSkipsRead(t *testing.T) {
	t.Parallel()

	reader := &countingReader{data: []byte("ODOL")}
	result, err := AnalyzeReader("model.p3d", reader, ProfileFast)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if reader.calls != 0 || result.Type != "bi.p3d" || result.RequiredPrefix != 0 {
		t.Fatalf("reader calls = %d, result = %+v; want no reads and path-only P3D", reader.calls, result)
	}
}

func TestAnalyzeReaderIntermittentNoProgress(t *testing.T) {
	t.Parallel()

	reader := &intermittentReader{data: []byte("ODOL")}
	result, err := AnalyzeReader("model.p3d", reader, ProfileNormal)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if result.Type != "bi.p3d.odol" || reader.zeroReads != 2 {
		t.Fatalf("result = %+v, zero reads = %d; want ODOL and two zero reads", result, reader.zeroReads)
	}
}

func TestAnalyzeReaderExpectedContentMissing(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeReader("texture.png", strings.NewReader(strings.Repeat("x", thoroughPrefixSize)), ProfileThorough)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}
	if result.Type != "image.png" || result.Relation != RelationPathOnly ||
		!result.Has(FlagExpectedContentMissing) || result.Has(FlagIncomplete) {
		t.Fatalf("result = %+v, want complete PNG path evidence with missing content", result)
	}
}

func TestAnalyzeReaderAndFileMatchAnalyze(t *testing.T) {
	t.Parallel()

	data := []byte("ODOL")
	readerResult, err := AnalyzeReader("model.p3d", strings.NewReader(string(data)), ProfileNormal)
	if err != nil {
		t.Fatalf("AnalyzeReader: %v", err)
	}

	path := filepath.Join(t.TempDir(), "model.p3d")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	fileResult, err := AnalyzeFile(path, ProfileNormal)
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}

	want := Analyze("model.p3d", data, ProfileNormal)
	if readerResult != want || fileResult != want {
		t.Fatalf("Analyze=%+v, Reader=%+v, File=%+v; want equal results", want, readerResult, fileResult)
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
	if got := Analyze(" config.cpp", nil, ProfileFast).Type; got != "bi.config.rv.cpp" {
		t.Fatalf("space-prefixed config type = %q, want extension mapping", got)
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
