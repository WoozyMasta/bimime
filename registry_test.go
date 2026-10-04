// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import (
	"errors"
	"testing"
)

func TestProbeContentDetectors(t *testing.T) {
	t.Parallel()

	edds := make([]byte, 136)
	copy(edds[:4], []byte("DDS "))
	copy(edds[128:132], []byte("COPY"))
	edds[132] = 4

	modernEDDS := make([]byte, 40)
	copy(modernEDDS[:4], []byte("DDS "))
	copy(modernEDDS[36:40], []byte("ENF1"))

	genericDDS := make([]byte, 40)
	copy(genericDDS[:4], []byte("DDS "))

	testCases := []struct {
		name string
		data []byte
		want TypeID
	}{
		{name: "rap", data: []byte{0, 'r', 'a', 'P'}, want: "bi.rap"},
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10}, want: "image.png"},
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF}, want: "image.jpeg"},
		{name: "gif", data: []byte("GIF89a"), want: "image.gif"},
		{name: "webp", data: []byte("RIFF\x24\x00\x00\x00WEBP"), want: "image.webp"},
		{name: "edds", data: edds, want: "image.edds"},
		{name: "edds-enfusion-marker", data: modernEDDS, want: "image.edds"},
		{name: "dds-without-edds-marker", data: genericDDS, want: "image.dds"},
		{name: "ogg", data: []byte("OggS\x00\x02"), want: "audio.ogg"},
		{name: "mp3-id3", data: []byte("ID3\x04\x00\x00"), want: "audio.mp3"},
		{name: "wav", data: []byte("RIFF\x24\x00\x00\x00WAVE"), want: "audio.wav"},
		{name: "mp4", data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}, want: "video.mp4"},
		{name: "pak-form", data: []byte("FORM\x01\x3c\x74\x09PAC1HEAD"), want: "bi.package.pak"},
		{name: "rdb-form", data: []byte("FORM\x00\x01\x2d\x3dRDBC\x06\x00\x00\x00"), want: "bi.db.rdb"},
		{name: "smd-form", data: []byte("FORM\x00\x00\x02\x14SMDFVERS"), want: "bi.world.smd"},
		{name: "odol", data: []byte("ODOL"), want: "bi.p3d.odol"},
		{name: "mlod", data: []byte("MLOD"), want: "bi.p3d.mlod"},
		{name: "oprw", data: []byte("OPRW"), want: "bi.wrp.oprw"},
		{name: "8wvr", data: []byte("8WVR"), want: "bi.wrp.8wvr"},
		{name: "9vbw", data: []byte("9VBW"), want: "bi.wrp.9vbw"},
		{name: "0wzd", data: []byte("0WZD"), want: "bi.wrp.0wzd"},
		{name: "sequence", data: []byte{0xD3, 'S', 'E', 'Q'}, want: "bi.sequence.seq"},
		{name: "shader", data: []byte("DXBC\x00\x00"), want: "bi.shader.dxbc"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := Probe("", testCase.data)
			if result.Type != testCase.want {
				t.Fatalf("Probe type = %q, want %q; result=%+v", result.Type, testCase.want, result)
			}
			if result.ByContent != testCase.want {
				t.Fatalf("Probe content evidence = %q, want %q", result.ByContent, testCase.want)
			}
		})
	}
}

func TestProbeRejectsInvalidMP3Frame(t *testing.T) {
	t.Parallel()

	for _, prefix := range [][]byte{
		{0xFF, 0xE0, 0x00, 0x00},
		{0xFF, 0xF9, 0x90, 0x64},
		{0xFF, 0xFB, 0xF0, 0x64},
		{0xFF, 0xFB, 0x9C, 0x64},
	} {
		result := Analyze("track.mp3", prefix, ProfileThorough)
		if result.ByContent != "" || result.Relation != RelationPathOnly {
			t.Errorf("Analyze(track.mp3, % x) = %+v, want no MP3 content evidence", prefix, result)
		}
	}
}

func TestWeakSignaturesRequirePathContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		path       string
		prefix     []byte
		want       TypeID
		wantSource ContentSource
	}{
		{name: "PAA", path: "texture.paa", prefix: []byte{1, 255}, want: "image.paa", wantSource: ContentSourceMagic},
		{name: "BMP", path: "image.bmp", prefix: []byte("BM"), want: "image.bmp", wantSource: ContentSourceMagic},
		{name: "MP3 frame", path: "track.mp3", prefix: []byte{0xFF, 0xFB, 0x90, 0x64}, want: "audio.mp3", wantSource: ContentSourceStructure},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			global := Probe("unknown.xyz", testCase.prefix)
			if global.Type != UnknownID || global.ByContent != "" {
				t.Fatalf("global Probe result = %+v, want no discovery from weak signature", global)
			}

			verified := Analyze(testCase.path, testCase.prefix, ProfileThorough)
			if verified.Type != testCase.want || verified.ByContent != testCase.want ||
				verified.ContentSource != testCase.wantSource || verified.Relation != RelationConfirmed {
				t.Fatalf("path verification = %+v, want confirmation of %q", verified, testCase.want)
			}
		})
	}

	id3 := Probe("unknown.xyz", []byte("ID3\x04\x00\x00"))
	if id3.Type != "audio.mp3" || id3.ByContent != "audio.mp3" {
		t.Fatalf("ID3 discovery = %+v, want globally discovered MP3", id3)
	}
}

func TestAnalyzePathMappings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want TypeID
	}{
		{name: "x.rvmat", want: "bi.rvmat"},
		{name: "x.bisurf", want: "bi.surface.bisurf"},
		{name: "x.jpg", want: "image.jpeg"},
		{name: "x.xlsx", want: "application.xlsx"},
		{name: "x.P3D", want: "bi.p3d"},
		{name: "config.cpp", want: "bi.config.main.cpp"},
		{name: "folder/texHeaders.bin", want: "bi.texheaders"},
		{name: "mapgroupcluster.xml", want: "bi.ce.mapgroupcluster"},
		{name: "mapgroupcluster2.xml", want: "bi.ce.mapgroupcluster"},
		{name: "mapgroupcluster.exe", want: UnknownID},
		{name: "mapgroupcluster", want: UnknownID},
		{name: "mapgroupcluster.xml.bak", want: UnknownID},
		{name: "env/hare_territories.xml", want: "bi.ce.env.territories"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := Analyze(testCase.name, nil, ProfileFast).Type; got != testCase.want {
				t.Fatalf("Analyze(%q) = %q, want %q", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestContextualRAPRefinements(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		want         TypeID
		wantPath     TypeID
		wantRelation Relation
	}{
		{
			name:         "config.bin",
			want:         "bi.config.main.bin",
			wantPath:     "bi.config.main.bin",
			wantRelation: RelationConfirmed,
		},
		{
			name:         "mod.bin",
			want:         "bi.mod.bin",
			wantPath:     "bi.mod.bin",
			wantRelation: RelationConfirmed,
		},
		{
			name:         "surface.rvmat",
			want:         "bi.rvmat.bin",
			wantPath:     "bi.rvmat",
			wantRelation: RelationRefined,
		},
		{
			name:         "terrain.bisurf",
			want:         "bi.surface.bisurf.bin",
			wantPath:     "bi.surface.bisurf",
			wantRelation: RelationRefined,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := Probe(testCase.name, []byte{0, 'r', 'a', 'P'})
			if result.Type != testCase.want || result.ByPath != testCase.wantPath ||
				result.ByContent != "bi.rap" || result.Relation != testCase.wantRelation {
				t.Fatalf("result = %+v, want type %q and raw RAP content evidence", result, testCase.want)
			}
		})
	}
}

func TestRegistryCopiesAndMetadata(t *testing.T) {
	t.Parallel()

	typ, ok := Lookup("bi.rap")
	if !ok || typ.ID != "bi.rap" || !typ.Has(TraitBinary) {
		t.Fatalf("Lookup(bi.rap) = %+v, %v", typ, ok)
	}
	typ.Extensions[0] = "changed"
	again, _ := Lookup("bi.rap")
	if again.Extensions[0] != "rap" {
		t.Fatal("Lookup must return an independent extension slice")
	}

	registry := Registry()
	registry[0].Extensions[0] = "changed"
	if got, _ := Lookup("bi.rap"); got.Extensions[0] != "rap" {
		t.Fatal("Registry must return independent extension slices")
	}
	for _, record := range builtinRegistry.records {
		traits := record.typ.Traits & (TraitText | TraitBinary)
		if traits != TraitText && traits != TraitBinary {
			t.Errorf("type %q must have exactly one text/binary trait, got %b", record.typ.ID, traits)
		}
	}
}

func TestStructuredProbeCapabilities(t *testing.T) {
	t.Parallel()

	if builtinRegistry.thoroughPrefix != thoroughPrefixSize {
		t.Fatalf("compiled thorough prefix = %d, want %d", builtinRegistry.thoroughPrefix, thoroughPrefixSize)
	}
	if err := validateStructuredProbeCapabilities(builtinRegistry); err != nil {
		t.Fatalf("validate structured probes: %v", err)
	}

	for _, capability := range builtinRegistry.structuredProbeCapabilities {
		if capability.requiredPrefix < 1 || capability.requiredPrefix > thoroughPrefixSize {
			t.Errorf("probe %d prefix requirement = %d, outside bounds", capability.kind, capability.requiredPrefix)
		}
		for _, target := range capability.targets {
			if _, ok := Lookup(target.typeID); !ok {
				t.Errorf("probe %d references unregistered type %q", capability.kind, target.typeID)
			}
		}
	}
}

func TestRegistryThoroughPrefixExcludesVerifyOnlyProbes(t *testing.T) {
	t.Parallel()

	pathRecord := registryRecord{typ: Type{
		ID:          "test.path",
		MIME:        "application/octet-stream",
		Description: "Path-scoped verifier target",
		Extensions:  []string{"path"},
		Traits:      TraitBinary,
	}}
	discoveryRecord := registryRecord{
		typ: Type{
			ID:          "test.discovered",
			MIME:        "application/octet-stream",
			Description: "Globally discoverable type",
			Extensions:  []string{"discovered"},
			Traits:      TraitBinary,
		},
		magic: [][]byte{[]byte("DISCOVER")},
	}
	staticVerifierRecord := registryRecord{
		typ: Type{
			ID:          "test.static-verifier",
			MIME:        "application/octet-stream",
			Description: "Path-scoped static verifier",
			Extensions:  []string{"verify"},
			Traits:      TraitBinary,
		},
		magic:      [][]byte{make([]byte, 64)},
		magicScope: probeVerify,
	}
	structuredVerifier := structuredProbeCapability{
		kind:           structuredProbeMP3,
		scope:          probeVerify,
		requiredPrefix: 32,
		source:         ContentSourceStructure,
		targets: []structuredProbeTarget{{
			typeID:       "test.path",
			verifyPrefix: 32,
		}},
	}

	registry, err := buildRegistry(
		[]registryRecord{pathRecord, discoveryRecord, staticVerifierRecord},
		nil,
		nil,
		nil,
		nil,
		nil,
		[]structuredProbeCapability{structuredVerifier},
	)
	if err != nil {
		t.Fatalf("buildRegistry() error = %v", err)
	}

	if registry.thoroughPrefix != len(discoveryRecord.magic[0]) {
		t.Fatalf("thorough prefix = %d, want discovery prefix %d", registry.thoroughPrefix, len(discoveryRecord.magic[0]))
	}
	if got := registry.verifierRequiredPrefix[pathRecord.typ.ID]; got != 32 {
		t.Fatalf("path verifier prefix = %d, want 32", got)
	}
	if got := registry.verifierRequiredPrefix[staticVerifierRecord.typ.ID]; got != 64 {
		t.Fatalf("static verifier prefix = %d, want 64", got)
	}
}

func TestCompiledVerifierRequiredPrefixes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		typeID TypeID
		want   int
	}{
		{typeID: "image.png", want: 8},
		{typeID: "image.dds", want: 4},
		{typeID: "image.edds", want: 4},
		{typeID: "bi.package.pak", want: 12},
		{typeID: "bi.animation.anm", want: 16},
		{typeID: "text.cfg", want: 0},
		{typeID: "text.xml", want: 0},
		{typeID: "bi.rvmat", want: 4},
		{typeID: "text.sqf", want: 0},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(string(testCase.typeID), func(t *testing.T) {
			t.Parallel()

			if got := builtinRegistry.verifierRequiredPrefix[testCase.typeID]; got != testCase.want {
				t.Fatalf("verifier prefix for %q = %d, want %d", testCase.typeID, got, testCase.want)
			}
		})
	}
}

func TestStructuredProbeEvidenceSource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data []byte
		want TypeID
	}{
		{
			name: "FORM",
			data: []byte("FORM\x01\x3c\x74\x09PAC1HEAD"),
			want: "bi.package.pak",
		},
		{
			name: "RIFF",
			data: []byte("RIFF\x24\x00\x00\x00WAVE"),
			want: "audio.wav",
		},
		{
			name: "MP4",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'},
			want: "video.mp4",
		},
		{
			name: "DDS",
			data: []byte("DDS "),
			want: "image.dds",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := Probe("", testCase.data)
			if result.ByContent != testCase.want || result.ContentSource != ContentSourceStructure {
				t.Fatalf("Probe result = %+v, want structured evidence for %q", result, testCase.want)
			}
		})
	}

	frame := Analyze("track.mp3", []byte{0xFF, 0xFB, 0x90, 0x64}, ProfileThorough)
	if frame.ByContent != "audio.mp3" || frame.ContentSource != ContentSourceStructure {
		t.Fatalf("MP3 frame verification = %+v, want structured evidence for audio.mp3", frame)
	}

	tagged := Probe("", []byte("ID3\x04\x00\x00"))
	if tagged.ByContent != "audio.mp3" || tagged.ContentSource != ContentSourceMagic {
		t.Fatalf("ID3 result = %+v, want static magic evidence", tagged)
	}
}

func TestBuildRegistryRejectsInvalidDefinitions(t *testing.T) {
	t.Parallel()

	valid := func(id TypeID, extension string) registryRecord {
		return registryRecord{typ: Type{
			ID:          id,
			MIME:        "application/octet-stream",
			Description: "Test type",
			Extensions:  []string{extension},
			Traits:      TraitBinary,
		}}
	}

	testCases := []struct {
		name        string
		definitions []registryRecord
		fileNames   map[string]string
	}{
		{
			name:        "duplicate IDs",
			definitions: []registryRecord{valid("test.one", "one"), valid("TEST.ONE", "two")},
		},
		{
			name:        "duplicate extensions",
			definitions: []registryRecord{valid("test.one", "same"), valid("test.two", "same")},
		},
		{
			name: "missing text or binary trait",
			definitions: []registryRecord{{typ: Type{
				ID:          "test.one",
				MIME:        "application/octet-stream",
				Description: "Test type",
			}}},
		},
		{
			name:        "unknown filename target",
			definitions: []registryRecord{valid("test.one", "one")},
			fileNames:   map[string]string{"config.bin": "missing.type"},
		},
		{
			name:        "duplicate signatures",
			definitions: []registryRecord{valid("test.one", "one"), valid("test.two", "two")},
		},
	}
	testCases[len(testCases)-1].definitions[0].magic = [][]byte{[]byte("MAGIC")}
	testCases[len(testCases)-1].definitions[1].magic = [][]byte{[]byte("MAGIC")}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := buildRegistry(testCase.definitions, testCase.fileNames, nil, nil, nil, nil, nil)
			if err == nil || !errors.Is(err, errInvalidRegistry) {
				t.Fatalf("buildRegistry error = %v, want invalid registry error", err)
			}
		})
	}
}

func TestBuildRegistryRejectsInvalidNormalProbes(t *testing.T) {
	t.Parallel()

	pathRecord := registryRecord{typ: Type{
		ID:          "test.base",
		Family:      "test.family",
		MIME:        "application/octet-stream",
		Description: "Generic type",
		Extensions:  []string{"base"},
		Traits:      TraitBinary,
	}}
	contentRecord := registryRecord{
		typ: Type{
			ID:          "test.child",
			Family:      "test.family",
			MIME:        "application/octet-stream",
			Description: "Specific type",
			Extensions:  []string{"child"},
			Traits:      TraitBinary,
		},
		magic:       [][]byte{[]byte("MAGC")},
		specificity: 1,
	}

	testCases := []struct {
		name         string
		pathType     TypeID
		contentTypes []TypeID
		required     int
		content      registryRecord
	}{
		{
			name:         "unknown path",
			pathType:     "missing.path",
			contentTypes: []TypeID{"test.child"},
			required:     4,
			content:      contentRecord,
		},
		{
			name:         "unbounded requirement",
			pathType:     "test.base",
			contentTypes: []TypeID{"test.child"},
			required:     thoroughPrefixSize + 1,
			content:      contentRecord,
		},
		{
			name:         "undersized requirement",
			pathType:     "test.base",
			contentTypes: []TypeID{"test.child"},
			required:     3,
			content:      contentRecord,
		},
		{
			name:         "content without signature",
			pathType:     "test.base",
			contentTypes: []TypeID{"test.child"},
			required:     4,
			content:      registryRecord{typ: contentRecord.typ},
		},
		{
			name:         "incompatible family",
			pathType:     "test.base",
			contentTypes: []TypeID{"test.child"},
			required:     4,
			content: registryRecord{
				typ: Type{
					ID:          "test.child",
					Family:      "other.family",
					MIME:        "application/octet-stream",
					Description: "Specific type",
					Extensions:  []string{"child"},
					Traits:      TraitBinary,
				},
				magic:       [][]byte{[]byte("MAGC")},
				specificity: 1,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := buildRegistry(
				[]registryRecord{pathRecord, testCase.content},
				nil,
				nil,
				nil,
				[]normalProbeCapability{{
					pathType:       testCase.pathType,
					scope:          probeRefine,
					contentTypes:   testCase.contentTypes,
					requiredPrefix: testCase.required,
				}},
				nil,
				nil,
			)
			if err == nil || !errors.Is(err, errInvalidRegistry) {
				t.Fatalf("buildRegistry error = %v, want invalid registry error", err)
			}
		})
	}
}

func TestBuildRegistryRejectsIncompatibleContextualRefinement(t *testing.T) {
	t.Parallel()

	definitions := []registryRecord{
		{typ: Type{
			ID:          "test.path",
			Family:      "test.path-family",
			MIME:        "application/octet-stream",
			Description: "Path type",
			Extensions:  []string{"path"},
			Traits:      TraitBinary,
		}},
		{typ: Type{
			ID:          "test.content",
			Family:      "test.content-family",
			MIME:        "application/octet-stream",
			Description: "Content type",
			Extensions:  []string{"content"},
			Traits:      TraitBinary,
		}},
		{typ: Type{
			ID:          "test.target",
			Family:      "test.target-family",
			MIME:        "application/octet-stream",
			Description: "Incompatible contextual target",
			Extensions:  []string{"target"},
			Traits:      TraitBinary,
		}},
	}
	contextual := map[contextualRefinementKey]TypeID{
		{pathType: "test.path", contentType: "test.content"}: "test.target",
	}

	_, err := buildRegistry(definitions, nil, nil, contextual, nil, nil, nil)
	if err == nil || !errors.Is(err, errInvalidRegistry) {
		t.Fatalf("buildRegistry error = %v, want invalid contextual refinement error", err)
	}
}
