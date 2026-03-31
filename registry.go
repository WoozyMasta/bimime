// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import "regexp"

// registryRecord stores one runtime registry record with optional magic signatures.
type registryRecord struct {
	contentPatterns []*regexp.Regexp
	magic           [][]byte
	typ             Type
}

// magicMatch stores one signature mapped to a type id for probing.
type magicMatch struct {
	typeID    string
	signature []byte
}

// registryRecords defines known BI game file types.
var registryRecords = normalizeRegistryRecords([]registryRecord{
	{
		typ: Type{
			ID:          "bi.rap",
			MIME:        "application/x-bohemia-rap",
			Description: "raP binarized config payload",
			Extensions:  []string{"rap"},
		},
		magic: [][]byte{{0x00, 'r', 'a', 'P'}},
	},
	{
		typ: Type{
			ID:          "bi.p3d.odol",
			MIME:        "model/x-bohemia-p3d-odol",
			Description: "P3D ODOL binarized model",
		},
		magic: [][]byte{[]byte("ODOL")},
	},
	{
		typ: Type{
			ID:          "bi.p3d.mlod",
			MIME:        "model/x-bohemia-p3d-mlod",
			Description: "P3D MLOD model source format",
		},
		magic: [][]byte{[]byte("MLOD")},
	},
	{
		typ: Type{
			ID:          "bi.wrp.oprw",
			MIME:        "application/x-bohemia-wrp-oprw",
			Description: "WRP OPRW binarized terrain format",
		},
		magic: [][]byte{[]byte("OPRW")},
	},
	{
		typ: Type{
			ID:          "bi.wrp.8wvr",
			MIME:        "application/x-bohemia-wrp-8wvr",
			Description: "WRP 8WVR terrain source format",
		},
		magic: [][]byte{[]byte("8WVR")},
	},
	{
		typ: Type{
			ID:          "bi.wrp.9vbw",
			MIME:        "application/x-bohemia-wrp-9vbw",
			Description: "WRP 9VBW terrain variant from Visitor 4 with multi-rvmat",
			Extensions:  []string{"vbs"},
		},
		magic: [][]byte{[]byte("9VBW")},
	},
	{
		typ: Type{
			ID:          "bi.wrp.0wzd",
			MIME:        "application/x-bohemia-wrp-0wzd",
			Description: "WRP 0WZD terrain variant with persistent object IDs",
			Extensions:  []string{"wzd"},
		},
		magic: [][]byte{[]byte("0WZD")},
	},
	{
		typ: Type{
			ID:          "bi.navmesh.tesm",
			MIME:        "application/x-bohemia-tesm",
			Description: "TESM AI navigation mesh binary format",
			Extensions:  []string{"nm"},
		},
		magic: [][]byte{[]byte("TESM")},
	},
	{
		typ: Type{
			ID:          "bi.pbo",
			MIME:        "application/x-bohemia-pbo",
			Description: "PBO archive format",
			Extensions:  []string{"pbo"},
		},
		magic: [][]byte{[]byte("sreV")},
	},
	{
		typ: Type{
			ID:          "image.paa",
			MIME:        "image/x-bohemia-paa",
			Description: "PAA/LEVF texture image format",
			Extensions:  []string{"paa", "pac"},
		},
		magic: [][]byte{
			{1, 255}, {2, 255}, {3, 255}, {4, 255}, {5, 255},
			{68, 68}, {85, 21}, {136, 136}, {128, 128},
		},
	},
	{
		typ: Type{
			ID:          "image.png",
			MIME:        "image/png",
			Description: "PNG image",
			Extensions:  []string{"png"},
		},
		magic: [][]byte{{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}},
	},
	{
		typ: Type{
			ID:          "image.jpeg",
			MIME:        "image/jpeg",
			Description: "JPEG image",
			Extensions:  []string{"jpg", "jpeg"},
		},
		magic: [][]byte{{0xFF, 0xD8, 0xFF}},
	},
	{
		typ: Type{
			ID:          "image.bmp",
			MIME:        "image/bmp",
			Description: "BMP image",
			Extensions:  []string{"bmp"},
		},
		magic: [][]byte{[]byte("BM")},
	},
	{
		typ: Type{
			ID:          "image.gif",
			MIME:        "image/gif",
			Description: "GIF image",
			Extensions:  []string{"gif"},
		},
		magic: [][]byte{[]byte("GIF87a"), []byte("GIF89a")},
	},
	{
		typ: Type{
			ID:          "image.svg",
			MIME:        "image/svg+xml",
			Description: "SVG vector image",
			Extensions:  []string{"svg"},
		},
	},
	{
		typ: Type{
			ID:          "image.webp",
			MIME:        "image/webp",
			Description: "WebP image",
			Extensions:  []string{"webp"},
		},
	},
	{
		typ: Type{
			ID:          "image.dds",
			MIME:        "image/vnd.ms-dds",
			Description: "DDS texture image",
			Extensions:  []string{"dds"},
		},
		magic: [][]byte{[]byte("DDS ")},
	},
	{
		typ: Type{
			ID:          "image.edds",
			MIME:        "image/x-edds",
			Description: "EDDS texture container with DDS header and block table",
			Extensions:  []string{"edds"},
		},
	},
	{
		typ: Type{
			ID:          "audio.ogg",
			MIME:        "audio/ogg",
			Description: "Ogg audio stream",
			Extensions:  []string{"ogg"},
		},
		magic: [][]byte{[]byte("OggS")},
	},
	{
		typ: Type{
			ID:          "audio.mp3",
			MIME:        "audio/mpeg",
			Description: "MP3 audio stream",
			Extensions:  []string{"mp3"},
		},
		magic: [][]byte{[]byte("ID3")},
	},
	{
		typ: Type{
			ID:          "audio.wav",
			MIME:        "audio/wav",
			Description: "WAV audio stream",
			Extensions:  []string{"wav"},
		},
	},
	{
		typ: Type{
			ID:          "video.mp4",
			MIME:        "video/mp4",
			Description: "MP4 video container",
			Extensions:  []string{"mp4"},
		},
	},
	{
		typ: Type{
			ID:          "bi.rvmat",
			MIME:        "text/x-bohemia-rvmat",
			Description: "RV material text config, can also be RAP",
			Extensions:  []string{"rvmat"},
		},
	},
	{
		typ: Type{
			ID:          "bi.rvmat.bin",
			MIME:        "application/x-bohemia-rvmat-rap",
			Description: "RV material RAP-binarized payload",
		},
		magic: [][]byte{{0x00, 'r', 'a', 'P'}},
	},
	{
		typ: Type{
			ID:          "bi.p3d",
			MIME:        "model/x-bohemia-p3d",
			Description: "P3D model (generic extension match without magic disambiguation)",
			Extensions:  []string{"p3d"},
		},
	},
	{
		typ: Type{
			ID:          "bi.wrp",
			MIME:        "application/x-bohemia-wrp",
			Description: "WRP terrain (generic extension match without magic disambiguation)",
			Extensions:  []string{"wrp"},
		},
	},
	{
		typ: Type{
			ID:          "bi.script.enforce",
			MIME:        "text/x-bohemia-enforce-script",
			Description: "Enforce Script source text (Enfusion engine)",
			Extensions:  []string{"c"},
		},
	},
	{
		typ: Type{
			ID:          "bi.config.rv.cpp",
			MIME:        "text/x-bohemia-rv-config",
			Description: "RV config source text",
			Extensions:  []string{"cpp"},
		},
	},
	{
		typ: Type{
			ID:          "bi.config.rv.hpp",
			MIME:        "text/x-bohemia-rv-config-header",
			Description: "RV config include/macros definitions text",
			Extensions:  []string{"hpp"},
		},
	},
	{
		typ: Type{
			ID:          "text.cfg",
			MIME:        "text/x-ini",
			Description: "Config text",
			Extensions:  []string{"cfg"},
		},
	},
	{
		typ: Type{
			ID:          "text.xml",
			MIME:        "application/xml",
			Description: "XML text",
			Extensions:  []string{"xml"},
		},
	},
	{
		typ: Type{
			ID:          "text.json",
			MIME:        "application/json",
			Description: "JSON text",
			Extensions:  []string{"json"},
		},
	},
	{
		typ: Type{
			ID:          "text.csv",
			MIME:        "text/csv",
			Description: "CSV text",
			Extensions:  []string{"csv"},
		},
	},
	{
		typ: Type{
			ID:          "application.xls",
			MIME:        "application/vnd.ms-excel",
			Description: "Excel spreadsheet (.xls)",
			Extensions:  []string{"xls"},
		},
	},
	{
		typ: Type{
			ID:          "application.xlsx",
			MIME:        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Description: "Excel spreadsheet (.xlsx)",
			Extensions:  []string{"xlsx"},
		},
	},
	{
		typ: Type{
			ID:          "application.ods",
			MIME:        "application/vnd.oasis.opendocument.spreadsheet",
			Description: "OpenDocument spreadsheet (.ods)",
			Extensions:  []string{"ods"},
		},
	},
	{
		typ: Type{
			ID:          "text.html",
			MIME:        "text/html",
			Description: "HTML text (including lightweight widget markup)",
			Extensions:  []string{"html", "htm"},
		},
	},
	{
		typ: Type{
			ID:          "text.plain",
			MIME:        "text/plain",
			Description: "Plain text",
			Extensions:  []string{"txt"},
		},
	},
	{
		typ: Type{
			ID:          "text.shader.vert",
			MIME:        "text/x-shader-vert",
			Description: "Vertex shader source text",
			Extensions:  []string{"vert"},
		},
	},
	{
		typ: Type{
			ID:          "text.shader.frag",
			MIME:        "text/x-shader-frag",
			Description: "Fragment shader source text",
			Extensions:  []string{"frag"},
		},
	},
	{
		typ: Type{
			ID:          "image.tga",
			MIME:        "image/x-tga",
			Description: "TGA image",
			Extensions:  []string{"tga"},
		},
	},
	{
		typ: Type{
			ID:          "model.obj",
			MIME:        "model/obj",
			Description: "Wavefront OBJ model source text",
			Extensions:  []string{"obj"},
		},
	},
	{
		typ: Type{
			ID:          "text.mtl",
			MIME:        "text/plain",
			Description: "Wavefront MTL material library text",
			Extensions:  []string{"mtl"},
		},
	},
	{
		typ: Type{
			ID:          "model.fbx",
			MIME:        "model/vnd.autodesk.fbx",
			Description: "FBX model/animation interchange file",
			Extensions:  []string{"fbx"},
		},
	},
	{
		typ: Type{
			ID:          "model.blend",
			MIME:        "application/x-blender",
			Description: "Blender scene/source file",
			Extensions:  []string{"blend"},
		},
	},
	{
		typ: Type{
			ID:          "model.dae",
			MIME:        "model/vnd.collada+xml",
			Description: "COLLADA model/animation scene",
			Extensions:  []string{"dae"},
		},
	},
	{
		typ: Type{
			ID:          "model.gltf",
			MIME:        "model/gltf+json",
			Description: "glTF scene description (JSON)",
			Extensions:  []string{"gltf"},
		},
	},
	{
		typ: Type{
			ID:          "model.glb",
			MIME:        "model/gltf-binary",
			Description: "glTF binary scene bundle",
			Extensions:  []string{"glb"},
		},
	},
	{
		typ: Type{
			ID:          "image.psd",
			MIME:        "image/vnd.adobe.photoshop",
			Description: "Adobe Photoshop document",
			Extensions:  []string{"psd", "psb"},
		},
	},
	{
		typ: Type{
			ID:          "image.xcf",
			MIME:        "image/x-xcf",
			Description: "GIMP project image",
			Extensions:  []string{"xcf"},
		},
	},
	{
		typ: Type{
			ID:          "font.ttf",
			MIME:        "font/ttf",
			Description: "TrueType font",
			Extensions:  []string{"ttf"},
		},
	},
	{
		typ: Type{
			ID:          "bi.ui.layout",
			MIME:        "application/x-bohemia-layout",
			Description: "Enfusion UI widget layout definition text",
			Extensions:  []string{"layout"},
		},
	},
	{
		typ: Type{
			ID:          "bi.ui.imageset",
			MIME:        "application/x-bohemia-imageset",
			Description: "Enfusion UI image set definition text",
			Extensions:  []string{"imageset"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\bImageSetClass\b`)},
	},
	{
		typ: Type{
			ID:          "bi.ui.styles",
			MIME:        "application/x-bohemia-widget-styles",
			Description: "Enfusion UI widget style definitions",
			Extensions:  []string{"styles"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)<\s*WidgetStyles\b`)},
	},
	{
		typ: Type{
			ID:          "text.qss",
			MIME:        "text/x-qt-stylesheet",
			Description: "Qt Style Sheet text (Workbench themes)",
			Extensions:  []string{"qss"},
		},
	},
	{
		typ: Type{
			ID:          "bi.font.fnt",
			MIME:        "application/x-bohemia-fnt",
			Description: "Enfusion UI font resource (binary)",
			Extensions:  []string{"fnt"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.anm",
			MIME:        "application/x-bohemia-anm",
			Description: "Enfusion binary animation file",
			Extensions:  []string{"anm"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.ae",
			MIME:        "text/x-bohemia-animation-events",
			Description: "Animation events table text",
			Extensions:  []string{"ae"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.adeb",
			MIME:        "application/x-bohemia-adeb",
			Description: "Animation debug stream, likely binary; magic unknown",
			Extensions:  []string{"adeb"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.pap",
			MIME:        "text/x-bohemia-pap",
			Description: "Procedural animation project text",
			Extensions:  []string{"pap"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.siga",
			MIME:        "text/x-bohemia-siga",
			Description: "Procedural animation signal text",
			Extensions:  []string{"siga"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.txa",
			MIME:        "text/x-bohemia-txa",
			Description: "Text animation source (txa)",
			Extensions:  []string{"txa"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.agf",
			MIME:        "text/x-bohemia-anim-graph-file",
			Description: "Animation graph sheet file text",
			Extensions:  []string{"agf"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.ast",
			MIME:        "application/x-bohemia-animset-template",
			Description: "Animation graph set template text",
			Extensions:  []string{"ast"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\$animsettemplate\b`)},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.asi",
			MIME:        "application/x-bohemia-animset-instance",
			Description: "Animation graph set instance text",
			Extensions:  []string{"asi"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\$animsetinstance\b`)},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.asy",
			MIME:        "application/x-bohemia-anim-sync-table",
			Description: "Animation graph sync table text",
			Extensions:  []string{"asy"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.aw",
			MIME:        "application/x-bohemia-anim-workspace",
			Description: "Animation graph workspace text",
			Extensions:  []string{"aw"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\$animworkspace\b`)},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.agr",
			MIME:        "application/x-bohemia-anim-graph",
			Description: "Animation graph definition text",
			Extensions:  []string{"agr"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\$animgraph\b`)},
	},
	{
		typ: Type{
			ID:          "bi.sequence.seq",
			MIME:        "application/x-bohemia-seq",
			Description: "BI engine sequence binary payload",
			Extensions:  []string{"seq"},
		},
		magic: [][]byte{{0xD3, 'S', 'E', 'Q'}},
	},
	{
		typ: Type{
			ID:          "bi.surface.bisurf",
			MIME:        "text/x-bohemia-surface",
			Description: "BI engine surface config text, can also be RAP",
			Extensions:  []string{"bisurf"},
		},
	},
	{
		typ: Type{
			ID:          "bi.surface.bisurf.bin",
			MIME:        "application/x-bohemia-surface-rap",
			Description: "BI engine surface config RAP-binarized payload",
		},
		magic: [][]byte{{0x00, 'r', 'a', 'P'}},
	},
	{
		typ: Type{
			ID:          "bi.world.map",
			MIME:        "application/x-bohemia-map",
			Description: "BI engine map payload (.map), format varies by tool/title",
			Extensions:  []string{"map"},
		},
	},
	{
		typ: Type{
			ID:          "bi.world.areaflags-map",
			MIME:        "application/x-bohemia-areaflags-map",
			Description: "Central Economy territory flags for loot and AI zoning",
		},
	},
	{
		typ: Type{
			ID:          "bi.emat",
			MIME:        "text/x-bohemia-emat",
			Description: "Enfusion material definition text",
			Extensions:  []string{"emat"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\bmaterial\b`)},
	},
	{
		typ: Type{
			ID:          "bi.font.fxy",
			MIME:        "application/x-bohemia-fxy",
			Description: "BI bitmap font glyph index/mapping payload",
			Extensions:  []string{"fxy"},
		},
	},
	{
		typ: Type{
			ID:          "bi.effects.ptc",
			MIME:        "application/x-bohemia-ptc",
			Description: "Enfusion particle system definition text",
			Extensions:  []string{"ptc"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\bEffectDef\b`)},
	},
	{
		typ: Type{
			ID:          "bi.effects.txo",
			MIME:        "text/x-bohemia-txo",
			Description: "Enfusion text model source (source for xob)",
			Extensions:  []string{"txo"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\$object\b`)},
	},
	{
		typ: Type{
			ID:          "bi.shader.dxbc",
			MIME:        "application/x-directx-shader",
			Description: "DirectX bytecode shader payload",
			Extensions:  []string{"pso", "vso", "cso"},
		},
		magic: [][]byte{[]byte("DXBC")},
	},
	{
		typ: Type{
			ID:          "bi.object.xob",
			MIME:        "application/x-bohemia-xob",
			Description: "Enfusion binary model file (built from txo)",
			Extensions:  []string{"xob"},
		},
	},
	{
		typ: Type{
			ID:          "bi.audio.acp",
			MIME:        "text/x-bohemia-acp",
			Description: "Audio component definition text",
			Extensions:  []string{"acp"},
		},
	},
	{
		typ: Type{
			ID:          "bi.audio.afm",
			MIME:        "text/x-bohemia-afm",
			Description: "Audio final mixer definition text",
			Extensions:  []string{"afm"},
		},
	},
	{
		typ: Type{
			ID:          "bi.audio.sig",
			MIME:        "text/x-bohemia-sig",
			Description: "Audio signal logic definition text",
			Extensions:  []string{"sig"},
		},
	},
	{
		typ: Type{
			ID:          "bi.audio.snd",
			MIME:        "application/x-bohemia-snd",
			Description: "Sound container payload",
			Extensions:  []string{"snd"},
		},
	},
	{
		typ: Type{
			ID:          "bi.ai.bt",
			MIME:        "text/x-bohemia-behavior-tree",
			Description: "AI behavior tree definition text",
			Extensions:  []string{"bt"},
		},
	},
	{
		typ: Type{
			ID:          "bi.component.ct",
			MIME:        "text/x-bohemia-component-template",
			Description: "Entity component template definition text",
			Extensions:  []string{"ct"},
		},
	},
	{
		typ: Type{
			ID:          "bi.config.conf",
			MIME:        "text/x-bohemia-config",
			Description: "Generic config text",
			Extensions:  []string{"conf"},
		},
	},
	{
		typ: Type{
			ID:          "bi.material.gamemat",
			MIME:        "text/x-bohemia-gamemat",
			Description: "Game material definition text",
			Extensions:  []string{"gamemat"},
		},
	},
	{
		typ: Type{
			ID:          "bi.material.physmat",
			MIME:        "text/x-bohemia-physmat",
			Description: "Physics material definition text",
			Extensions:  []string{"physmat"},
		},
	},
	{
		typ: Type{
			ID:          "bi.navmesh.nmn",
			MIME:        "application/x-bohemia-nmn",
			Description: "Navmesh instance, likely binary; magic unknown",
			Extensions:  []string{"nmn"},
		},
	},
	{
		typ: Type{
			ID:          "bi.package.pak",
			MIME:        "application/x-bohemia-pak",
			Description: "Game/mod data archive package",
			Extensions:  []string{"pak"},
		},
	},
	{
		typ: Type{
			ID:          "bi.db.rdb",
			MIME:        "application/x-bohemia-rdb",
			Description: "Resource database index payload",
			Extensions:  []string{"rdb"},
		},
	},
	{
		typ: Type{
			ID:          "bi.db.stars",
			MIME:        "application/x-bohemia-stars",
			Description: "Runtime stars database, likely binary; magic unknown",
			Extensions:  []string{"stars"},
		},
	},
	{
		typ: Type{
			ID:          "bi.db.st",
			MIME:        "text/x-bohemia-st",
			Description: "Localization string table text",
			Extensions:  []string{"st"},
		},
	},
	{
		typ: Type{
			ID:          "bi.meta",
			MIME:        "application/x-bohemia-meta",
			Description: "BI engine metadata sidecar text",
			Extensions:  []string{"meta"},
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\bMetaFileClass\b`)},
	},
	{
		typ: Type{
			ID:          "bi.font.fsdf",
			MIME:        "application/x-bohemia-fsdf",
			Description: "BI engine SDF font resource definition text",
			Extensions:  []string{"fsdf"},
		},
	},
	{
		typ: Type{
			ID:          "bi.project.gproj",
			MIME:        "application/x-bohemia-project",
			Description: "Project-specific Workbench settings",
			Extensions:  []string{"gproj"},
		},
	},
	{
		typ: Type{
			ID:          "bi.preview.pre",
			MIME:        "text/x-bohemia-pre",
			Description: "Resource Viewer preset text",
			Extensions:  []string{"pre"},
		},
	},
	{
		typ: Type{
			ID:          "bi.physics.ragdoll",
			MIME:        "text/x-bohemia-ragdoll",
			Description: "Ragdoll definition text",
			Extensions:  []string{"ragdoll"},
		},
	},
	{
		typ: Type{
			ID:          "bi.project.sproj",
			MIME:        "application/x-bohemia-project",
			Description: "Workbench project XML",
			Extensions:  []string{"sproj"},
		},
	},
	{
		typ: Type{
			ID:          "bi.project.ssln",
			MIME:        "application/x-bohemia-solution",
			Description: "Workbench solution XML",
			Extensions:  []string{"ssln"},
		},
	},
	{
		typ: Type{
			ID:          "bi.texheaders",
			MIME:        "application/x-bohemia-texheaders",
			Description: "texHeaders.bin texture metadata index (0DHT, version 1)",
		},
		magic: [][]byte{[]byte("0DHT")},
	},
	{
		typ: Type{
			ID:          "bi.sign.bikey",
			MIME:        "application/x-bohemia-bikey",
			Description: "BI engine public key (.bikey)",
			Extensions:  []string{"bikey"},
		},
	},
	{
		typ: Type{
			ID:          "bi.sign.biprivatekey",
			MIME:        "application/x-bohemia-biprivatekey",
			Description: "BI engine private key (.biprivatekey)",
			Extensions:  []string{"biprivatekey"},
		},
	},
	{
		typ: Type{
			ID:          "bi.sign.bisign",
			MIME:        "application/x-bohemia-bisign",
			Description: "BI engine signature file (.bisign)",
			Extensions:  []string{"bisign"},
		},
	},
	{
		typ: Type{
			ID:          "bi.config.main.cpp",
			MIME:        "text/x-bohemia-addon-config",
			Description: "Addon root config.cpp (RaP-capable text config)",
		},
	},
	{
		typ: Type{
			ID:          "bi.config.main.bin",
			MIME:        "application/x-bohemia-addon-config-rap",
			Description: "Addon root config.bin (RAP-binarized config)",
		},
		magic: [][]byte{{0x00, 'r', 'a', 'P'}},
	},
	{
		typ: Type{
			ID:          "bi.mod.cpp",
			MIME:        "text/x-bohemia-mod-config",
			Description: "Mod root mod.cpp metadata config",
		},
	},
	{
		typ: Type{
			ID:          "bi.mod.bin",
			MIME:        "application/x-bohemia-mod-config-rap",
			Description: "Mod root mod.bin metadata config",
		},
		magic: [][]byte{{0x00, 'r', 'a', 'P'}},
	},
	{
		typ: Type{
			ID:          "bi.model.cfg",
			MIME:        "text/x-bohemia-model-config",
			Description: "Model animation config (model.cfg)",
		},
		contentPatterns: []*regexp.Regexp{regexp.MustCompile(`(?i)\bclass\s+cfgmodels\b`)},
	},
	{
		typ: Type{
			ID:          "bi.ce.db.economy",
			MIME:        "application/xml",
			Description: "Central Economy runtime toggles for init, persistence, and respawn groups",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.db.events",
			MIME:        "application/xml",
			Description: "Central Economy dynamic event definitions and spawn controllers",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.db.globals",
			MIME:        "application/xml",
			Description: "Central Economy global spawn limits and behavior variables",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.db.types",
			MIME:        "application/xml",
			Description: "Central Economy catalog of spawnable entities with tags and limits",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.db.messages",
			MIME:        "application/xml",
			Description: "Central Economy server message schedule and timing definitions",
		},
	},
	{
		typ: Type{
			ID:          "bi.mission.init-c",
			MIME:        "text/x-bohemia-enforce-script",
			Description: "Mission startup script executed during scenario initialization",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgeconomycore",
			MIME:        "application/xml",
			Description: "Core economy behavior switches and bootstrap parameters",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgenvironment",
			MIME:        "application/xml",
			Description: "Animal and infected habitat and territory configuration",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgareaflags",
			MIME:        "application/xml",
			Description: "Definitions of area limiter flags and semantic names",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfglimitsdefinition",
			MIME:        "application/xml",
			Description: "Definitions of tag and category limiters for economy rules",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfglimitsdefinitionuser",
			MIME:        "application/xml",
			Description: "User-friendly limiter aliases built on base limiter definitions",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgeventspawns",
			MIME:        "application/xml",
			Description: "Placement and rotation sets for dynamic event spawn points",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgeventgroups",
			MIME:        "application/xml",
			Description: "Grouping and weighting rules for dynamic event pools",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgplayerspawnpoints",
			MIME:        "application/xml",
			Description: "Base rules and positions used to generate player spawns",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgspawnabletypes",
			MIME:        "application/xml",
			Description: "Attachment and cargo randomization rules per spawnable entity",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgrandompresets",
			MIME:        "application/xml",
			Description: "Reusable random presets for cargo and attachment generation",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgweather",
			MIME:        "application/xml",
			Description: "Weather behavior configuration with limits and transitions",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgignorelist",
			MIME:        "application/xml",
			Description: "Central Economy ignored entities and exclusions list",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.mapclusterproto",
			MIME:        "application/xml",
			Description: "Cluster map-group prototypes with loot container layouts",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.mapgroupcluster",
			MIME:        "application/xml",
			Description: "Exported positions of cluster map groups across the world",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.mapgroupdirt",
			MIME:        "application/xml",
			Description: "Positions of map groups not bound to explicit world objects",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.mapgrouppos",
			MIME:        "application/xml",
			Description: "Exported positions of standard map groups and building instances",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.mapgroupproto",
			MIME:        "application/xml",
			Description: "Standard map-group prototypes with loot container layouts",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.ceproject-config",
			MIME:        "application/xml",
			Description: "CEProject zg-config map descriptor (ceproject-config.xml)",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.env.territories",
			MIME:        "application/xml",
			Description: "CE territory layer payload (*_territories.xml)",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgundergroundtriggers",
			MIME:        "application/json",
			Description: "Underground trigger volumes and logic definitions",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfgeffectarea",
			MIME:        "application/json",
			Description: "Effect area configuration for CE zones and modifiers",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.cfggameplay",
			MIME:        "application/json",
			Description: "Gameplay tuning profile used by CE and mission systems",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.gameplay-gear-presets",
			MIME:        "application/json",
			Description: "Gameplay gear preset catalog for random loadout assembly",
		},
	},
	{
		typ: Type{
			ID:          "bi.ce.object-spawner",
			MIME:        "application/json",
			Description: "Object spawner definitions and placement rules",
		},
	},
	{
		typ: Type{
			ID:          "bi.mission.sqm",
			MIME:        "text/x-bohemia-mission-sqm",
			Description: "Mission editor data (mission.sqm)",
			Extensions:  []string{"sqm"},
		},
	},
	{
		typ: Type{
			ID:          "bi.stringtable.csv",
			MIME:        "text/csv",
			Description: "Stringtable.csv localization table",
		},
	},
	{
		typ: Type{
			ID:          "bi.stringtable.xml",
			MIME:        "application/xml",
			Description: "Stringtable.xml localization table",
		},
	},
	{
		typ: Type{
			ID:          "bi.world.ent",
			MIME:        "text/x-bohemia-ent",
			Description: "World scene definition text",
			Extensions:  []string{"ent"},
		},
	},
	{
		typ: Type{
			ID:          "bi.world.et",
			MIME:        "text/x-bohemia-et",
			Description: "Entity template definition text",
			Extensions:  []string{"et"},
		},
	},
	{
		typ: Type{
			ID:          "bi.world.layer",
			MIME:        "text/x-bohemia-layer",
			Description: "World layer definition text",
			Extensions:  []string{"layer"},
		},
	},
	{
		typ: Type{
			ID:          "bi.world.smap",
			MIME:        "application/x-bohemia-smap",
			Description: "Sound map data payload",
			Extensions:  []string{"smap"},
		},
	},
	{
		typ: Type{
			ID:          "bi.world.topo",
			MIME:        "application/x-bohemia-topo",
			Description: "Topography data payload",
			Extensions:  []string{"topo"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.asc",
			MIME:        "text/plain",
			Description: "ESRI ASCII height map import/export",
			Extensions:  []string{"asc"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.desc",
			MIME:        "text/x-bohemia-terrain-desc",
			Description: "Terrain dialog configuration text",
			Extensions:  []string{"desc"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.terr",
			MIME:        "application/x-bohemia-terr",
			Description: "Terrain data payload",
			Extensions:  []string{"terr"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.ttile",
			MIME:        "application/x-bohemia-ttile",
			Description: "Runtime terrain tile, likely binary; magic unknown",
			Extensions:  []string{"ttile"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.bterr",
			MIME:        "application/x-bohemia-bterr",
			Description: "Editor terrain data, likely binary; magic unknown",
			Extensions:  []string{"bterr"},
		},
	},
	{
		typ: Type{
			ID:          "bi.terrain.bttile",
			MIME:        "application/x-bohemia-bttile",
			Description: "Editor terrain tile, likely binary; magic unknown",
			Extensions:  []string{"bttile"},
		},
	},
	{
		typ: Type{
			ID:          "bi.vehicle.vhcsurf",
			MIME:        "text/x-bohemia-vhcsurf",
			Description: "Vehicle surface properties definition text",
			Extensions:  []string{"vhcsurf"},
		},
	},
	{
		typ: Type{
			ID:          "text.sqf",
			MIME:        "text/x-sqf",
			Description: "SQF script source text (Real Virtuality)",
			Extensions:  []string{"sqf"},
		},
	},
	{
		typ: Type{
			ID:          "application.sqfc",
			MIME:        "application/x-sqfc",
			Description: "SQF bytecode binary (Real Virtuality)",
			Extensions:  []string{"sqfc"},
		},
	},
	{
		typ: Type{
			ID:          "text.sqs",
			MIME:        "text/x-sqs",
			Description: "SQS script source text (Real Virtuality)",
			Extensions:  []string{"sqs"},
		},
	},
	{
		typ: Type{
			ID:          "text.lip",
			MIME:        "text/x-lip",
			Description: "Lip-sync text data",
			Extensions:  []string{"lip"},
		},
	},
	{
		typ: Type{
			ID:          "text.pew",
			MIME:        "text/x-pew",
			Description: "Visitor project source text",
			Extensions:  []string{"pew"},
		},
	},
	{
		typ: Type{
			ID:          "audio.wss",
			MIME:        "audio/x-bohemia-wss",
			Description: "WSS audio stream",
			Extensions:  []string{"wss"},
		},
	},
	{
		typ: Type{
			ID:          "text.rpt",
			MIME:        "text/plain",
			Description: "Engine report log text",
			Extensions:  []string{"rpt"},
		},
	},
	{
		typ: Type{
			ID:          "application.bidmp",
			MIME:        "application/octet-stream",
			Description: "Engine crash dump (.bidmp)",
			Extensions:  []string{"bidmp"},
		},
	},
	{
		typ: Type{
			ID:          "application.mdmp",
			MIME:        "application/octet-stream",
			Description: "Engine crash dump (.mdmp)",
			Extensions:  []string{"mdmp"},
		},
	},
	{
		typ: Type{
			ID:          "bi.crash.context-bin",
			MIME:        "application/octet-stream",
			Description: "Crash context.bin",
		},
	},
	{
		typ: Type{
			ID:          "application.bin",
			MIME:        "application/octet-stream",
			Description: "Generic binary payload",
			Extensions:  []string{"bin"},
		},
	},
	{
		typ: Type{
			ID:          "application.lzss",
			MIME:        "application/x-lzss",
			Description: "Raw LZSS stream (no stable magic header)",
			Extensions:  []string{"lzss"},
		},
	},
	{
		typ: Type{
			ID:          "application.lzo",
			MIME:        "application/x-lzo",
			Description: "Raw LZO1X stream (no stable magic header)",
			Extensions:  []string{"lzo"},
		},
	},
})

// typeByID indexes registry records by id.
var typeByID = buildTypeByID(registryRecords)

// typeByExtension indexes registry records by extension.
var typeByExtension = buildTypeByExtension(registryRecords)

// magicIndex stores signatures sorted by length desc for stable probing.
var magicIndex = buildMagicIndex(registryRecords)

// typeIDByFileName maps well-known file names to stable type ids.
var typeIDByFileName = map[string]string{
	"texheaders.bin":              "bi.texheaders",
	"config.cpp":                  "bi.config.main.cpp",
	"config.bin":                  "bi.config.main.bin",
	"mod.cpp":                     "bi.mod.cpp",
	"mod.bin":                     "bi.mod.bin",
	"model.cfg":                   "bi.model.cfg",
	"mission.sqm":                 "bi.mission.sqm",
	"stringtable.csv":             "bi.stringtable.csv",
	"stringtable.xml":             "bi.stringtable.xml",
	"context.bin":                 "bi.crash.context-bin",
	"areaflags.map":               "bi.world.areaflags-map",
	"economy.xml":                 "bi.ce.db.economy",
	"events.xml":                  "bi.ce.db.events",
	"globals.xml":                 "bi.ce.db.globals",
	"messages.xml":                "bi.ce.db.messages",
	"types.xml":                   "bi.ce.db.types",
	"init.c":                      "bi.mission.init-c",
	"cfgeconomycore.xml":          "bi.ce.cfgeconomycore",
	"cfgenvironment.xml":          "bi.ce.cfgenvironment",
	"cfgareaflags.xml":            "bi.ce.cfgareaflags",
	"cfglimitsdefinition.xml":     "bi.ce.cfglimitsdefinition",
	"cfglimitsdefinitionuser.xml": "bi.ce.cfglimitsdefinitionuser",
	"cfgeventspawns.xml":          "bi.ce.cfgeventspawns",
	"cfgeventgroups.xml":          "bi.ce.cfgeventgroups",
	"cfgplayerspawnpoints.xml":    "bi.ce.cfgplayerspawnpoints",
	"cfgspawnabletypes.xml":       "bi.ce.cfgspawnabletypes",
	"cfgrandompresets.xml":        "bi.ce.cfgrandompresets",
	"cfgweather.xml":              "bi.ce.cfgweather",
	"cfgignorelist.xml":           "bi.ce.cfgignorelist",
	"ceproject-config.xml":        "bi.ce.ceproject-config",
	"cfgundergroundtriggers.json": "bi.ce.cfgundergroundtriggers",
	"cfgeffectarea.json":          "bi.ce.cfgeffectarea",
	"cfggameplay.json":            "bi.ce.cfggameplay",
	"gameplay-gear-presets.json":  "bi.ce.gameplay-gear-presets",
	"object-spawner.json":         "bi.ce.object-spawner",
	"mapclusterproto.xml":         "bi.ce.mapclusterproto",
	"mapgroupdirt.xml":            "bi.ce.mapgroupdirt",
	"mapgrouppos.xml":             "bi.ce.mapgrouppos",
	"mapgroupproto.xml":           "bi.ce.mapgroupproto",
}

// typeIDByFilePrefix maps filename prefixes to type ids.
// Used for rolling exports like mapgroupcluster1.xml, mapgroupcluster2.xml.
var typeIDByFilePrefix = map[string]string{
	"mapgroupcluster": "bi.ce.mapgroupcluster",
}

// typeIDByFileSuffix maps filename suffixes to type ids.
// Used for dynamic names like hare_territories.xml.
var typeIDByFileSuffix = map[string]string{
	"_territories.xml": "bi.ce.env.territories",
}
