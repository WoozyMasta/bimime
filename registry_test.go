package bimime

import (
	"testing"
	"unicode/utf8"
)

func TestDetectByMagic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data []byte
		id   string
	}{
		{name: "rap", data: []byte{0x00, 'r', 'a', 'P', 0x01}, id: "bi.rap"},
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, id: "image.png"},
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF, 0xE0}, id: "image.jpeg"},
		{name: "bmp", data: []byte{'B', 'M', 0x10, 0x00}, id: "image.bmp"},
		{name: "gif", data: []byte("GIF89a\x00"), id: "image.gif"},
		{name: "webp", data: []byte{'R', 'I', 'F', 'F', 0x24, 0, 0, 0, 'W', 'E', 'B', 'P'}, id: "image.webp"},
		{name: "dds", data: []byte("DDS xxxx"), id: "image.dds"},
		{
			name: "edds",
			data: func() []byte {
				data := make([]byte, 136)
				copy(data[:4], []byte("DDS "))
				copy(data[128:132], []byte("COPY"))
				data[132] = 4
				return data
			}(),
			id: "image.edds",
		},
		{name: "ogg", data: []byte("OggS\x00\x02"), id: "audio.ogg"},
		{name: "mp3-id3", data: []byte("ID3\x04\x00\x00"), id: "audio.mp3"},
		{name: "mp3-frame", data: []byte{0xFF, 0xFB, 0x90, 0x64}, id: "audio.mp3"},
		{name: "wav", data: []byte{'R', 'I', 'F', 'F', 0x24, 0, 0, 0, 'W', 'A', 'V', 'E'}, id: "audio.wav"},
		{name: "mp4", data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 2, 0}, id: "video.mp4"},
		{name: "odol", data: []byte("ODOL"), id: "bi.p3d.odol"},
		{name: "mlod", data: []byte("MLOD"), id: "bi.p3d.mlod"},
		{name: "oprw", data: []byte("OPRW"), id: "bi.wrp.oprw"},
		{name: "8wvr", data: []byte("8WVR"), id: "bi.wrp.8wvr"},
		{name: "9vbw", data: []byte("9VBW"), id: "bi.wrp.9vbw"},
		{name: "0wzd", data: []byte("0WZD"), id: "bi.wrp.0wzd"},
		{name: "paa", data: []byte{1, 255, 0, 0}, id: "bi.paa"},
		{name: "dxbc", data: []byte("DXBC\x00\x00"), id: "bi.shader.dxbc"},
		{name: "seq", data: []byte{0xD3, 'S', 'E', 'Q', 0x00}, id: "bi.sequence.seq"},
		{name: "texheaders", data: []byte("0DHT\x01\x00\x00\x00"), id: "bi.texheaders"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			typ, ok := DetectByMagic(testCase.data)
			if !ok {
				t.Fatalf("DetectByMagic(%s): no match", testCase.name)
			}
			if typ.ID != testCase.id {
				t.Fatalf("DetectByMagic(%s): got %q want %q", testCase.name, typ.ID, testCase.id)
			}
		})
	}
}

func TestDetectByExtension(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		path string
		id   string
	}{
		{path: "x.rvmat", id: "bi.rvmat"},
		{path: "x.rap", id: "bi.rap"},
		{path: "x.jpg", id: "image.jpeg"},
		{path: "x.svg", id: "image.svg"},
		{path: "x.webp", id: "image.webp"},
		{path: "x.mp3", id: "audio.mp3"},
		{path: "x.wav", id: "audio.wav"},
		{path: "x.mp4", id: "video.mp4"},
		{path: "x.xls", id: "application.xls"},
		{path: "x.xlsx", id: "application.xlsx"},
		{path: "x.ods", id: "application.ods"},
		{path: "x.fods", id: "application.fods"},
		{path: "x.bikey", id: "bi.sign.bikey"},
		{path: "x.biprivatekey", id: "bi.sign.biprivatekey"},
		{path: "x.bisign", id: "bi.sign.bisign"},
		{path: "areaflags.map", id: "bi.world.areaflags-map"},
		{path: "economy.xml", id: "bi.ce.db.economy"},
		{path: "events.xml", id: "bi.ce.db.events"},
		{path: "globals.xml", id: "bi.ce.db.globals"},
		{path: "types.xml", id: "bi.ce.db.types"},
		{path: "init.c", id: "bi.mission.init-c"},
		{path: "cfgeconomycore.xml", id: "bi.ce.cfgeconomycore"},
		{path: "cfgenvironment.xml", id: "bi.ce.cfgenvironment"},
		{path: "cfgareaflags.xml", id: "bi.ce.cfgareaflags"},
		{path: "cfglimitsdefinition.xml", id: "bi.ce.cfglimitsdefinition"},
		{path: "cfglimitsdefinitionuser.xml", id: "bi.ce.cfglimitsdefinitionuser"},
		{path: "cfgeventspawns.xml", id: "bi.ce.cfgeventspawns"},
		{path: "cfgplayerspawnpoints.xml", id: "bi.ce.cfgplayerspawnpoints"},
		{path: "cfgspawnabletypes.xml", id: "bi.ce.cfgspawnabletypes"},
		{path: "cfgrandompresets.xml", id: "bi.ce.cfgrandompresets"},
		{path: "mapclusterproto.xml", id: "bi.ce.mapclusterproto"},
		{path: "mapgroupcluster2.xml", id: "bi.ce.mapgroupcluster"},
		{path: "mapgroupdirt.xml", id: "bi.ce.mapgroupdirt"},
		{path: "mapgrouppos.xml", id: "bi.ce.mapgrouppos"},
		{path: "mapgroupproto.xml", id: "bi.ce.mapgroupproto"},
		{path: "config.cpp", id: "bi.config.main.cpp"},
		{path: "config.bin", id: "bi.config.main.bin"},
		{path: "mod.cpp", id: "bi.mod.cpp"},
		{path: "mod.bin", id: "bi.mod.bin"},
		{path: "model.cfg", id: "bi.model.cfg"},
		{path: "mission.sqm", id: "bi.mission.sqm"},
		{path: "stringtable.csv", id: "bi.stringtable.csv"},
		{path: "stringtable.xml", id: "bi.stringtable.xml"},
		{path: "context.bin", id: "bi.crash.context-bin"},
		{path: "x.P3D", id: "bi.p3d"},
		{path: "x.wrp", id: "bi.wrp"},
		{path: "x.vbs", id: "bi.wrp.9vbw"},
		{path: "x.wzd", id: "bi.wrp.0wzd"},
		{path: "x.c", id: "bi.script.enforce"},
		{path: "x.cpp", id: "bi.config.rv.cpp"},
		{path: "x.hpp", id: "bi.config.rv.hpp"},
		{path: "x.layout", id: "bi.ui.layout"},
		{path: "x.imageset", id: "bi.ui.imageset"},
		{path: "x.styles", id: "bi.ui.styles"},
		{path: "x.qss", id: "text.qss"},
		{path: "x.html", id: "text.html"},
		{path: "x.sqf", id: "text.sqf"},
		{path: "x.sqfc", id: "application.sqfc"},
		{path: "x.sqs", id: "text.sqs"},
		{path: "x.lip", id: "text.lip"},
		{path: "x.pew", id: "text.pew"},
		{path: "x.wss", id: "audio.wss"},
		{path: "x.rpt", id: "text.rpt"},
		{path: "x.bidmp", id: "application.bidmp"},
		{path: "x.mdmp", id: "application.mdmp"},
		{path: "x.lzss", id: "application.lzss"},
		{path: "x.lzo", id: "application.lzo"},
		{path: "x.edds", id: "image.edds"},
		{path: "x.ast", id: "bi.animgraph.ast"},
		{path: "x.asi", id: "bi.animgraph.asi"},
		{path: "x.asy", id: "bi.animgraph.asy"},
		{path: "x.aw", id: "bi.animgraph.aw"},
		{path: "x.agr", id: "bi.animgraph.agr"},
		{path: "x.meta", id: "bi.meta"},
		{path: "x.gproj", id: "bi.project.gproj"},
		{path: "x.sproj", id: "bi.project.sproj"},
		{path: "x.ssln", id: "bi.project.ssln"},
		{path: "x.fsdf", id: "bi.font.fsdf"},
		{path: "x.obj", id: "model.obj"},
		{path: "x.mtl", id: "text.mtl"},
		{path: "x.fbx", id: "model.fbx"},
		{path: "x.blend", id: "model.blend"},
		{path: "x.dae", id: "model.dae"},
		{path: "x.gltf", id: "model.gltf"},
		{path: "x.glb", id: "model.glb"},
		{path: "x.psd", id: "image.psd"},
		{path: "x.psb", id: "image.psd"},
		{path: "x.xcf", id: "image.xcf"},
		{path: "x.ogg", id: "audio.ogg"},
		{path: "x.dds", id: "image.dds"},
		{path: "x.bin", id: "application.bin"},
		{path: "folder/texHeaders.bin", id: "bi.texheaders"},
	}

	for _, testCase := range testCases {
		typ, ok := DetectByExtension(testCase.path)
		if !ok {
			t.Fatalf("DetectByExtension(%q): no match", testCase.path)
		}
		if typ.ID != testCase.id {
			t.Fatalf("DetectByExtension(%q): got %q want %q", testCase.path, typ.ID, testCase.id)
		}
	}
}

func TestProbeMagicOverridesExtension(t *testing.T) {
	t.Parallel()

	result := Probe("config.rvmat", []byte{0x00, 'r', 'a', 'P', 0x00})
	if result.Source != SourceMagicAndExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceMagicAndExtension)
	}
	if result.ByExtension.ID != "bi.rvmat" {
		t.Fatalf("Probe extension id = %q, want bi.rvmat", result.ByExtension.ID)
	}
	if result.ByMagic.ID != "bi.rap" {
		t.Fatalf("Probe magic id = %q, want bi.rap", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.rap" {
		t.Fatalf("Probe resolved id = %q, want bi.rap", result.Resolved.ID)
	}
}

func TestProbeRAPOverridesBisurfExtension(t *testing.T) {
	t.Parallel()

	result := Probe("armor.bisurf", []byte{0x00, 'r', 'a', 'P', 0x00})
	if result.Source != SourceMagicAndExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceMagicAndExtension)
	}
	if result.ByExtension.ID != "bi.surface.bisurf" {
		t.Fatalf("Probe extension id = %q, want bi.surface.bisurf", result.ByExtension.ID)
	}
	if result.ByMagic.ID != "bi.rap" {
		t.Fatalf("Probe magic id = %q, want bi.rap", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.rap" {
		t.Fatalf("Probe resolved id = %q, want bi.rap", result.Resolved.ID)
	}
}

func TestProbeConfigBinPrefersPathHintOverRAP(t *testing.T) {
	t.Parallel()

	result := Probe("config.bin", []byte{0x00, 'r', 'a', 'P', 0x00})
	if result.Source != SourceMagicAndExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceMagicAndExtension)
	}
	if result.ByExtension.ID != "bi.config.main.bin" {
		t.Fatalf("Probe extension id = %q, want bi.config.main.bin", result.ByExtension.ID)
	}
	if result.ByMagic.ID != "bi.rap" {
		t.Fatalf("Probe magic id = %q, want bi.rap", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.config.main.bin" {
		t.Fatalf("Probe resolved id = %q, want bi.config.main.bin", result.Resolved.ID)
	}
}

func TestProbeMagicOverridesWRPExtension(t *testing.T) {
	t.Parallel()

	result := Probe("terrain.wrp", []byte("9VBW\x00\x00"))
	if result.Source != SourceMagicAndExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceMagicAndExtension)
	}
	if result.ByExtension.ID != "bi.wrp" {
		t.Fatalf("Probe extension id = %q, want bi.wrp", result.ByExtension.ID)
	}
	if result.ByMagic.ID != "bi.wrp.9vbw" {
		t.Fatalf("Probe magic id = %q, want bi.wrp.9vbw", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.wrp.9vbw" {
		t.Fatalf("Probe resolved id = %q, want bi.wrp.9vbw", result.Resolved.ID)
	}
}

func TestProbeEddsPrefersExtensionWhenMagicAmbiguous(t *testing.T) {
	t.Parallel()

	result := Probe("texture.edds", []byte("DDS "))
	if result.Source != SourceMagicAndExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceMagicAndExtension)
	}
	if result.ByExtension.ID != "image.edds" {
		t.Fatalf("Probe extension id = %q, want image.edds", result.ByExtension.ID)
	}
	if result.ByMagic.ID != "image.dds" {
		t.Fatalf("Probe magic id = %q, want image.dds", result.ByMagic.ID)
	}
	if result.Resolved.ID != "image.edds" {
		t.Fatalf("Probe resolved id = %q, want image.edds", result.Resolved.ID)
	}
}

func TestLookupAndRegistry(t *testing.T) {
	t.Parallel()

	if _, ok := Lookup("bi.rap"); !ok {
		t.Fatal("Lookup(bi.rap) must succeed")
	}

	registry := Registry()
	if len(registry) < 20 {
		t.Fatalf("Registry too small: got %d", len(registry))
	}
}

func TestIsRAP(t *testing.T) {
	t.Parallel()

	if !IsRAP([]byte{0x00, 'r', 'a', 'P'}) {
		t.Fatal("IsRAP must return true for RAP magic")
	}
	if IsRAP([]byte("DDS ")) {
		t.Fatal("IsRAP must return false for non-RAP payload")
	}
}

func TestBinaryHints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		id     string
		binary bool
	}{
		{id: "bi.rap", binary: true},
		{id: "image.webp", binary: true},
		{id: "audio.mp3", binary: true},
		{id: "audio.wav", binary: true},
		{id: "video.mp4", binary: true},
		{id: "bi.sign.bikey", binary: true},
		{id: "bi.sign.biprivatekey", binary: true},
		{id: "bi.sign.bisign", binary: true},
		{id: "bi.config.main.bin", binary: true},
		{id: "image.dds", binary: true},
		{id: "bi.config.main.cpp", binary: false},
		{id: "bi.model.cfg", binary: false},
		{id: "bi.ce.db.economy", binary: false},
		{id: "text.html", binary: false},
		{id: "image.svg", binary: false},
		{id: "model.obj", binary: false},
		{id: "text.mtl", binary: false},
		{id: "model.dae", binary: false},
		{id: "model.gltf", binary: false},
		{id: "model.fbx", binary: true},
		{id: "model.blend", binary: true},
		{id: "model.glb", binary: true},
		{id: "image.psd", binary: true},
		{id: "image.xcf", binary: true},
		{id: "application.xls", binary: true},
		{id: "application.xlsx", binary: true},
		{id: "application.ods", binary: true},
		{id: "application.fods", binary: true},
		{id: "bi.project.gproj", binary: false},
		{id: "text.sqf", binary: false},
		{id: "application.sqfc", binary: true},
		{id: "application.lzss", binary: true},
		{id: "application.lzo", binary: true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.id, func(t *testing.T) {
			t.Parallel()

			typ, ok := Lookup(testCase.id)
			if !ok {
				t.Fatalf("Lookup(%q): no match", testCase.id)
			}
			if typ.Binary != testCase.binary {
				t.Fatalf("Lookup(%q): Binary=%v want %v", testCase.id, typ.Binary, testCase.binary)
			}
		})
	}
}

func TestShortDescriptionLengthLimit(t *testing.T) {
	t.Parallel()

	for _, record := range registryRecords {
		short := record.typ.ShortDescription
		if short == "" {
			t.Fatalf("type %q has empty short description", record.typ.ID)
		}

		if utf8.RuneCountInString(short) > maxShortDescriptionLen {
			t.Fatalf(
				"type %q short description is too long: len=%d limit=%d desc=%q",
				record.typ.ID,
				utf8.RuneCountInString(short),
				maxShortDescriptionLen,
				short,
			)
		}
	}
}
