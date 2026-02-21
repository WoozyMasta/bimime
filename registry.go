// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

// registryRecord stores one runtime registry record with optional magic signatures.
type registryRecord struct {
	magic [][]byte
	typ   Type
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
			ID:          "bi.paa",
			MIME:        "image/x-bohemia-paa",
			Description: "PAA texture format",
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
			ID:          "application.fods",
			MIME:        "application/vnd.oasis.opendocument.spreadsheet-flat-xml",
			Description: "Flat OpenDocument spreadsheet (.fods)",
			Extensions:  []string{"fods"},
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
	},
	{
		typ: Type{
			ID:          "bi.ui.styles",
			MIME:        "application/x-bohemia-widget-styles",
			Description: "Enfusion UI widget style definitions",
			Extensions:  []string{"styles"},
		},
	},
	{
		typ: Type{
			ID:          "text.qss",
			MIME:        "text/x-qt-stylesheet",
			Description: "Qt stylesheet text",
			Extensions:  []string{"qss"},
		},
	},
	{
		typ: Type{
			ID:          "bi.font.fnt",
			MIME:        "application/x-bohemia-fnt",
			Description: "BI engine binary font resource",
			Extensions:  []string{"fnt"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animation.anm",
			MIME:        "application/x-bohemia-anm",
			Description: "BI engine animation set binary payload",
			Extensions:  []string{"anm"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.ast",
			MIME:        "application/x-bohemia-animset-template",
			Description: "Animation graph set template text",
			Extensions:  []string{"ast"},
		},
	},
	{
		typ: Type{
			ID:          "bi.animgraph.asi",
			MIME:        "application/x-bohemia-animset-instance",
			Description: "Animation graph set instance text",
			Extensions:  []string{"asi"},
		},
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
	},
	{
		typ: Type{
			ID:          "bi.animgraph.agr",
			MIME:        "application/x-bohemia-anim-graph",
			Description: "Animation graph definition text",
			Extensions:  []string{"agr"},
		},
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
			MIME:        "application/x-bohemia-surface",
			Description: "BI engine surface config (commonly RAP-binarized)",
			Extensions:  []string{"bisurf"},
		},
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
			ID:          "bi.effects.emat",
			MIME:        "application/x-bohemia-emat",
			Description: "BI engine effect/material definition text",
			Extensions:  []string{"emat"},
		},
	},
	{
		typ: Type{
			ID:          "bi.effects.fxy",
			MIME:        "application/x-bohemia-fxy",
			Description: "BI engine FXY text definition",
			Extensions:  []string{"fxy"},
		},
	},
	{
		typ: Type{
			ID:          "bi.effects.ptc",
			MIME:        "application/x-bohemia-ptc",
			Description: "BI engine particle effect definition text",
			Extensions:  []string{"ptc"},
		},
	},
	{
		typ: Type{
			ID:          "bi.effects.txo",
			MIME:        "application/x-bohemia-txo",
			Description: "BI engine texture/effect related payload",
			Extensions:  []string{"txo"},
		},
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
			Description: "BI engine XOB preview/object binary payload",
			Extensions:  []string{"xob"},
		},
	},
	{
		typ: Type{
			ID:          "bi.meta",
			MIME:        "application/x-bohemia-meta",
			Description: "BI engine metadata sidecar text",
			Extensions:  []string{"meta"},
		},
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
			Description: "Workbench project file",
			Extensions:  []string{"gproj"},
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
	},
	{
		typ: Type{
			ID:          "bi.model.cfg",
			MIME:        "text/x-bohemia-model-config",
			Description: "Model animation config (model.cfg)",
		},
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
	"types.xml":                   "bi.ce.db.types",
	"init.c":                      "bi.mission.init-c",
	"cfgeconomycore.xml":          "bi.ce.cfgeconomycore",
	"cfgenvironment.xml":          "bi.ce.cfgenvironment",
	"cfgareaflags.xml":            "bi.ce.cfgareaflags",
	"cfglimitsdefinition.xml":     "bi.ce.cfglimitsdefinition",
	"cfglimitsdefinitionuser.xml": "bi.ce.cfglimitsdefinitionuser",
	"cfgeventspawns.xml":          "bi.ce.cfgeventspawns",
	"cfgplayerspawnpoints.xml":    "bi.ce.cfgplayerspawnpoints",
	"cfgspawnabletypes.xml":       "bi.ce.cfgspawnabletypes",
	"cfgrandompresets.xml":        "bi.ce.cfgrandompresets",
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
