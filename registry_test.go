package bimime

import (
	"bytes"
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
		{
			name: "pak-form",
			data: []byte{'F', 'O', 'R', 'M', 0x01, 0x3C, 0x74, 0x09, 'P', 'A', 'C', '1', 'H', 'E', 'A', 'D'},
			id:   "bi.package.pak",
		},
		{
			name: "rdb-form",
			data: []byte{'F', 'O', 'R', 'M', 0x00, 0x01, 0x2D, 0x3D, 'R', 'D', 'B', 'C', 0x06, 0x00, 0x00, 0x00},
			id:   "bi.db.rdb",
		},
		{
			name: "anm-form",
			data: []byte{'F', 'O', 'R', 'M', 0x00, 0x00, 0x3B, 0xB5, 'A', 'N', 'I', 'M', 'S', 'E', 'T', '5'},
			id:   "bi.animation.anm",
		},
		{
			name: "xob-form",
			data: []byte{'F', 'O', 'R', 'M', 0x00, 0x00, 0x06, 0x4B, 'X', 'O', 'B', '6', 'H', 'E', 'A', 'D'},
			id:   "bi.object.xob",
		},
		{
			name: "fnt-form",
			data: []byte{'F', 'O', 'R', 'M', 0x00, 0x00, 0x08, 0x14, 'F', 'N', 'T', '2', 'G', 'L', 'P', 'S'},
			id:   "bi.font.fnt",
		},
		{name: "odol", data: []byte("ODOL"), id: "bi.p3d.odol"},
		{name: "mlod", data: []byte("MLOD"), id: "bi.p3d.mlod"},
		{name: "oprw", data: []byte("OPRW"), id: "bi.wrp.oprw"},
		{name: "8wvr", data: []byte("8WVR"), id: "bi.wrp.8wvr"},
		{name: "9vbw", data: []byte("9VBW"), id: "bi.wrp.9vbw"},
		{name: "0wzd", data: []byte("0WZD"), id: "bi.wrp.0wzd"},
		{name: "paa", data: []byte{1, 255, 0, 0}, id: "image.paa"},
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
		{path: "x.bisurf", id: "bi.surface.bisurf"},
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
		{path: "x.bikey", id: "bi.sign.bikey"},
		{path: "x.biprivatekey", id: "bi.sign.biprivatekey"},
		{path: "x.bisign", id: "bi.sign.bisign"},
		{path: "areaflags.map", id: "bi.world.areaflags-map"},
		{path: "economy.xml", id: "bi.ce.db.economy"},
		{path: "events.xml", id: "bi.ce.db.events"},
		{path: "globals.xml", id: "bi.ce.db.globals"},
		{path: "messages.xml", id: "bi.ce.db.messages"},
		{path: "types.xml", id: "bi.ce.db.types"},
		{path: "init.c", id: "bi.mission.init-c"},
		{path: "cfgeconomycore.xml", id: "bi.ce.cfgeconomycore"},
		{path: "cfgenvironment.xml", id: "bi.ce.cfgenvironment"},
		{path: "cfgareaflags.xml", id: "bi.ce.cfgareaflags"},
		{path: "cfglimitsdefinition.xml", id: "bi.ce.cfglimitsdefinition"},
		{path: "cfglimitsdefinitionuser.xml", id: "bi.ce.cfglimitsdefinitionuser"},
		{path: "cfgeventspawns.xml", id: "bi.ce.cfgeventspawns"},
		{path: "cfgeventgroups.xml", id: "bi.ce.cfgeventgroups"},
		{path: "cfgplayerspawnpoints.xml", id: "bi.ce.cfgplayerspawnpoints"},
		{path: "cfgspawnabletypes.xml", id: "bi.ce.cfgspawnabletypes"},
		{path: "cfgrandompresets.xml", id: "bi.ce.cfgrandompresets"},
		{path: "cfgweather.xml", id: "bi.ce.cfgweather"},
		{path: "cfgignorelist.xml", id: "bi.ce.cfgignorelist"},
		{path: "ceproject-config.xml", id: "bi.ce.ceproject-config"},
		{path: "env/hare_territories.xml", id: "bi.ce.env.territories"},
		{path: "cfgundergroundtriggers.json", id: "bi.ce.cfgundergroundtriggers"},
		{path: "cfgeffectarea.json", id: "bi.ce.cfgeffectarea"},
		{path: "cfggameplay.json", id: "bi.ce.cfggameplay"},
		{path: "gameplay-gear-presets.json", id: "bi.ce.gameplay-gear-presets"},
		{path: "object-spawner.json", id: "bi.ce.object-spawner"},
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
		{path: "x.emat", id: "bi.emat"},
		{path: "x.fxy", id: "bi.font.fxy"},
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
		{path: "x.agf", id: "bi.animgraph.agf"},
		{path: "x.asi", id: "bi.animgraph.asi"},
		{path: "x.asy", id: "bi.animgraph.asy"},
		{path: "x.aw", id: "bi.animgraph.aw"},
		{path: "x.agr", id: "bi.animgraph.agr"},
		{path: "x.ae", id: "bi.animation.ae"},
		{path: "x.adeb", id: "bi.animation.adeb"},
		{path: "x.pap", id: "bi.animation.pap"},
		{path: "x.siga", id: "bi.animation.siga"},
		{path: "x.txa", id: "bi.animation.txa"},
		{path: "x.acp", id: "bi.audio.acp"},
		{path: "x.afm", id: "bi.audio.afm"},
		{path: "x.sig", id: "bi.audio.sig"},
		{path: "x.snd", id: "bi.audio.snd"},
		{path: "x.bt", id: "bi.ai.bt"},
		{path: "x.ct", id: "bi.component.ct"},
		{path: "x.conf", id: "bi.config.conf"},
		{path: "x.gamemat", id: "bi.material.gamemat"},
		{path: "x.physmat", id: "bi.material.physmat"},
		{path: "x.nmn", id: "bi.navmesh.nmn"},
		{path: "x.pak", id: "bi.package.pak"},
		{path: "x.rdb", id: "bi.db.rdb"},
		{path: "x.stars", id: "bi.db.stars"},
		{path: "x.st", id: "bi.db.st"},
		{path: "x.pre", id: "bi.preview.pre"},
		{path: "x.ragdoll", id: "bi.physics.ragdoll"},
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
		{path: "x.ent", id: "bi.world.ent"},
		{path: "x.et", id: "bi.world.et"},
		{path: "x.layer", id: "bi.world.layer"},
		{path: "x.smap", id: "bi.world.smap"},
		{path: "x.topo", id: "bi.world.topo"},
		{path: "x.asc", id: "bi.terrain.asc"},
		{path: "x.desc", id: "bi.terrain.desc"},
		{path: "x.terr", id: "bi.terrain.terr"},
		{path: "x.ttile", id: "bi.terrain.ttile"},
		{path: "x.bterr", id: "bi.terrain.bterr"},
		{path: "x.bttile", id: "bi.terrain.bttile"},
		{path: "x.vhcsurf", id: "bi.vehicle.vhcsurf"},
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
	if result.ByMagic.ID != "bi.rvmat.bin" {
		t.Fatalf("Probe magic id = %q, want bi.rvmat.bin", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.rvmat.bin" {
		t.Fatalf("Probe resolved id = %q, want bi.rvmat.bin", result.Resolved.ID)
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
	if result.ByMagic.ID != "bi.surface.bisurf.bin" {
		t.Fatalf("Probe magic id = %q, want bi.surface.bisurf.bin", result.ByMagic.ID)
	}
	if result.Resolved.ID != "bi.surface.bisurf.bin" {
		t.Fatalf("Probe resolved id = %q, want bi.surface.bisurf.bin", result.Resolved.ID)
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
	if result.ByMagic.ID != "bi.config.main.bin" {
		t.Fatalf("Probe magic id = %q, want bi.config.main.bin", result.ByMagic.ID)
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

func TestProbeCfgModelsPatternSpecialization(t *testing.T) {
	t.Parallel()

	result := Probe("vehicle_hatchback.cfg", []byte("class CfgModels { class Test {}; };"))
	if result.Source != SourceExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceExtension)
	}
	if result.ByExtension.ID != "bi.model.cfg" {
		t.Fatalf("Probe extension id = %q, want bi.model.cfg", result.ByExtension.ID)
	}
	if result.Resolved.ID != "bi.model.cfg" {
		t.Fatalf("Probe resolved id = %q, want bi.model.cfg", result.Resolved.ID)
	}
}

func TestProbeCfgWithoutCfgModelsStaysGeneric(t *testing.T) {
	t.Parallel()

	result := Probe("vehicle_hatchback.cfg", []byte("foo=bar"))
	if result.Source != SourceExtension {
		t.Fatalf("Probe source = %q, want %q", result.Source, SourceExtension)
	}
	if result.ByExtension.ID != "text.cfg" {
		t.Fatalf("Probe extension id = %q, want text.cfg", result.ByExtension.ID)
	}
	if result.Resolved.ID != "text.cfg" {
		t.Fatalf("Probe resolved id = %q, want text.cfg", result.Resolved.ID)
	}
}

func TestProbeXMLWithZGConfigSpecialization(t *testing.T) {
	t.Parallel()

	result := Probe("custom-name.xml", []byte("<?xml version=\"1.0\"?><zg-config/>"))
	if result.ByExtension.ID != "bi.ce.ceproject-config" {
		t.Fatalf("ByExtension=%q want %q", result.ByExtension.ID, "bi.ce.ceproject-config")
	}
	if result.Resolved.ID != "bi.ce.ceproject-config" {
		t.Fatalf("Resolved=%q want %q", result.Resolved.ID, "bi.ce.ceproject-config")
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
		{id: "bi.rvmat.bin", binary: true},
		{id: "bi.surface.bisurf", binary: false},
		{id: "bi.surface.bisurf.bin", binary: true},
		{id: "bi.emat", binary: false},
		{id: "bi.font.fxy", binary: true},
		{id: "bi.font.fnt", binary: true},
		{id: "bi.animation.anm", binary: true},
		{id: "bi.animation.ae", binary: false},
		{id: "bi.animation.adeb", binary: true},
		{id: "bi.animation.pap", binary: false},
		{id: "bi.animation.siga", binary: false},
		{id: "bi.animation.txa", binary: false},
		{id: "bi.animgraph.agf", binary: false},
		{id: "bi.audio.acp", binary: false},
		{id: "bi.audio.afm", binary: false},
		{id: "bi.audio.sig", binary: false},
		{id: "bi.audio.snd", binary: true},
		{id: "bi.ai.bt", binary: false},
		{id: "bi.component.ct", binary: false},
		{id: "bi.config.conf", binary: false},
		{id: "bi.material.gamemat", binary: false},
		{id: "bi.material.physmat", binary: false},
		{id: "bi.navmesh.nmn", binary: true},
		{id: "bi.package.pak", binary: true},
		{id: "bi.db.rdb", binary: true},
		{id: "bi.db.stars", binary: true},
		{id: "bi.db.st", binary: false},
		{id: "bi.preview.pre", binary: false},
		{id: "bi.physics.ragdoll", binary: false},
		{id: "bi.world.ent", binary: false},
		{id: "bi.world.et", binary: false},
		{id: "bi.world.layer", binary: false},
		{id: "bi.world.smap", binary: true},
		{id: "bi.world.topo", binary: true},
		{id: "bi.terrain.asc", binary: false},
		{id: "bi.terrain.desc", binary: false},
		{id: "bi.terrain.terr", binary: true},
		{id: "bi.terrain.ttile", binary: true},
		{id: "bi.terrain.bterr", binary: true},
		{id: "bi.terrain.bttile", binary: true},
		{id: "bi.vehicle.vhcsurf", binary: false},
		{id: "bi.effects.txo", binary: false},
		{id: "bi.object.xob", binary: true},
		{id: "image.dds", binary: true},
		{id: "bi.config.main.cpp", binary: false},
		{id: "bi.model.cfg", binary: false},
		{id: "bi.ce.db.economy", binary: false},
		{id: "bi.ce.db.messages", binary: false},
		{id: "bi.ce.cfgeventgroups", binary: false},
		{id: "bi.ce.cfgweather", binary: false},
		{id: "bi.ce.cfgignorelist", binary: false},
		{id: "bi.ce.ceproject-config", binary: false},
		{id: "bi.ce.env.territories", binary: false},
		{id: "bi.ce.cfgundergroundtriggers", binary: false},
		{id: "bi.ce.cfgeffectarea", binary: false},
		{id: "bi.ce.cfggameplay", binary: false},
		{id: "bi.ce.gameplay-gear-presets", binary: false},
		{id: "bi.ce.object-spawner", binary: false},
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

func TestBuildTypeByIDSkipsInvalidRecords(t *testing.T) {
	t.Parallel()

	index := buildTypeByID([]registryRecord{
		{typ: Type{ID: ""}},
		{typ: Type{ID: "A"}},
		{typ: Type{ID: "a"}},
		{typ: Type{ID: "B"}},
	})

	if len(index) != 2 {
		t.Fatalf("len(index)=%d want 2", len(index))
	}

	if _, ok := index["a"]; !ok {
		t.Fatal(`index["a"] is missing`)
	}
	if _, ok := index["b"]; !ok {
		t.Fatal(`index["b"] is missing`)
	}
}

func TestBuildFilePrefixRecordsSkipsInvalidRecords(t *testing.T) {
	t.Parallel()

	records := buildFilePrefixRecords(map[string]string{
		"":      "bi.valid",
		"  a  ": "",
		"good":  "bi.good",
	})

	if len(records) != 1 {
		t.Fatalf("len(records)=%d want 1", len(records))
	}
	if records[0].prefix != "good" {
		t.Fatalf("prefix=%q want %q", records[0].prefix, "good")
	}
	if records[0].typeID != "bi.good" {
		t.Fatalf("typeID=%q want %q", records[0].typeID, "bi.good")
	}
}

func TestBuildTypeByExtensionSkipsInvalidRecords(t *testing.T) {
	t.Parallel()

	index := buildTypeByExtension([]registryRecord{
		{typ: Type{ID: "first", Extensions: []string{"foo"}}},
		{typ: Type{ID: "invalid", Extensions: []string{""}}},
		{typ: Type{ID: "second", Extensions: []string{"foo", "bar"}}},
	})

	if len(index) != 2 {
		t.Fatalf("len(index)=%d want 2", len(index))
	}
	if got := index["foo"].typ.ID; got != "first" {
		t.Fatalf(`index["foo"].typ.ID=%q want %q`, got, "first")
	}
	if got := index["bar"].typ.ID; got != "second" {
		t.Fatalf(`index["bar"].typ.ID=%q want %q`, got, "second")
	}
}

func TestBuildMagicIndexSkipsEmptySignatures(t *testing.T) {
	t.Parallel()

	index := buildMagicIndex([]registryRecord{
		{
			typ: Type{ID: "a"},
			magic: [][]byte{
				nil,
				[]byte("AA"),
			},
		},
	})

	if len(index) != 1 {
		t.Fatalf("len(index)=%d want 1", len(index))
	}
	if got := string(index[0].signature); got != "AA" {
		t.Fatalf("signature=%q want %q", got, "AA")
	}
	if got := index[0].typeID; got != "a" {
		t.Fatalf("typeID=%q want %q", got, "a")
	}
}

func TestRegistryRecordsIntegrity(t *testing.T) {
	t.Parallel()

	seenIDs := make(map[string]struct{}, len(registryRecords))
	seenExtensions := make(map[string]struct{})
	for _, record := range registryRecords {
		id := lowerKey(record.typ.ID)
		if id == "" {
			t.Fatal("registry contains empty type id")
		}
		if _, exists := seenIDs[id]; exists {
			t.Fatalf("registry contains duplicate type id %q", id)
		}
		seenIDs[id] = struct{}{}

		for _, ext := range record.typ.Extensions {
			key := lowerKey(ext)
			if key == "" {
				t.Fatalf("registry contains empty extension for type %q", id)
			}
			if _, exists := seenExtensions[key]; exists {
				t.Fatalf("registry contains duplicate extension mapping %q", key)
			}
			seenExtensions[key] = struct{}{}
		}

		for index, signature := range record.magic {
			if len(signature) == 0 {
				t.Fatalf("registry contains empty magic for %q at index %d", id, index)
			}
		}
	}

	if len(seenIDs) != len(typeByID) {
		t.Fatalf("typeByID size mismatch: records=%d index=%d", len(seenIDs), len(typeByID))
	}
	if len(seenExtensions) != len(typeByExtension) {
		t.Fatalf(
			"typeByExtension size mismatch: records=%d index=%d",
			len(seenExtensions),
			len(typeByExtension),
		)
	}

	for _, record := range registryRecords {
		for _, signature := range record.magic {
			found := false
			for _, indexed := range magicIndex {
				if indexed.typeID == lowerKey(record.typ.ID) &&
					bytes.Equal(indexed.signature, signature) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("missing magic index entry for type %q", record.typ.ID)
			}
		}
	}
}
