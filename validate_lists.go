// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

// Known pixel shader IDs observed in game data.
var knownPixelShaderID = map[string]struct{}{
	"TerrainX":                            {},
	"Super":                               {},
	"Multi":                               {},
	"Terrain1":                            {},
	"Terrain3":                            {},
	"TreeAdv":                             {},
	"Terrain5":                            {},
	"Terrain4":                            {},
	"Terrain9":                            {},
	"Terrain2":                            {},
	"Terrain8":                            {},
	"Terrain7":                            {},
	"Terrain15":                           {},
	"Terrain13":                           {},
	"Terrain11":                           {},
	"TreeAdvTrunk":                        {},
	"TerrainSNX":                          {},
	"Skin":                                {},
	"Terrain6":                            {},
	"Terrain12":                           {},
	"Grass":                               {},
	"SuperHair":                           {},
	"Terrain10":                           {},
	"NormalMapSpecularMap":                {},
	"Normal":                              {},
	"NormalMapSpecularDIMap":              {},
	"NormalMapDetailSpecularDIMap":        {},
	"NormalMapDiffuse":                    {},
	"Terrain14":                           {},
	"Tree":                                {},
	"CalmWater":                           {},
	"Detail":                              {},
	"NormalMapDetailMacroASSpecularDIMap": {},
	"NormalMapMacroASSpecularDIMap":       {},
	"super":                               {},
	"AlphaShadow":                         {},
	"Glass":                               {},
	"AlphaNoShadow":                       {},
	"DetailMacroAS":                       {},
	"NormalMap":                           {},
	"NormalMapDetailSpecularMap":          {},
	"SuperExt":                            {},
}

// Known vertex shader IDs observed in game data.
var knownVertexShaderID = map[string]struct{}{
	"Terrain":               {},
	"Super":                 {},
	"Multi":                 {},
	"TreeAdv":               {},
	"TreeAdvTrunk":          {},
	"Skin":                  {},
	"NormalMap":             {},
	"Grass":                 {},
	"Basic":                 {},
	"TreeNoFade":            {},
	"NormalMapDiffuseAlpha": {},
	"VSTerrain":             {},
	"NormalMapAS":           {},
	"CalmWater":             {},
	"TreeAdvModNormals":     {},
	"Tree":                  {},
	"TreeADV":               {},
	"super":                 {},
	"Glass":                 {},
	"BasicAS":               {},
	"NormalMapDiffuse":      {},
	"Treenofade":            {},
	"BasicAlpha":            {},
}

// Known stage names observed in game data.
var knownStageNames = map[string]struct{}{
	"Stage0":    {},
	"Stage1":    {},
	"Stage2":    {},
	"Stage3":    {},
	"Stage4":    {},
	"Stage5":    {},
	"Stage6":    {},
	"Stage7":    {},
	"Stage8":    {},
	"Stage9":    {},
	"Stage10":   {},
	"Stage11":   {},
	"Stage12":   {},
	"Stage13":   {},
	"Stage14":   {},
	"StageTI":   {},
	"StageLast": {},
}

// Known procedural texture functions observed in game data.
var knownProceduralFns = map[string]struct{}{
	"color":        {},
	"fresnel":      {},
	"fresnelglass": {},
	"irradiance":   {},
}

// Known procedural texture header formats.
var knownProceduralFormats = map[string]struct{}{
	"argb": {},
	"ai":   {},
}

// Known texture tags observed in procedural color() references.
var knownTextureTags = map[string]struct{}{
	"as":        {},
	"ca":        {},
	"cdt":       {},
	"co":        {},
	"dt":        {},
	"dtsmdi":    {},
	"mask":      {},
	"mc":        {},
	"mca":       {},
	"nohq":      {},
	"smdi":      {},
	"ads":       {},
	"adshq":     {},
	"cat":       {},
	"draftlco":  {},
	"gs":        {},
	"lca":       {},
	"lco":       {},
	"mco":       {},
	"no":        {},
	"noex":      {},
	"nof":       {},
	"nofex":     {},
	"nofhq":     {},
	"non":       {},
	"nopx":      {},
	"nsex":      {},
	"nshq":      {},
	"novhq":     {},
	"pr":        {},
	"raw":       {},
	"sky":       {},
	"sm":        {},
	"normalmap": {},
}

// shaderProfileHint keeps soft stage hints for known pixel shader profiles.
type shaderProfileHint struct {
	Required    []string // Required stages in normal valid files.
	Recommended []string // Common stages that are usually present.
}

// shaderProfileHints defines data-driven profile hints used by Validate.
//
// Notes from local corpus (2026-03-03, P:\DZ):
//   - super: Stage1/2/3/5/6 found in all files, Stage4/7 almost always.
//   - multi: Stage1..14 found in all files, Stage0 almost always.
//   - glass: Stage1/2 found in all files.
var shaderProfileHints = map[string]shaderProfileHint{
	"super": {
		Required: []string{
			"Stage1",
			"Stage2",
			"Stage3",
			"Stage5",
			"Stage6",
		},
		Recommended: []string{
			"Stage4",
			"Stage7",
		},
	},
	"multi": {
		Required: []string{
			"Stage0",
			"Stage1",
			"Stage2",
			"Stage3",
			"Stage4",
			"Stage5",
			"Stage6",
			"Stage7",
			"Stage8",
			"Stage9",
			"Stage10",
			"Stage11",
			"Stage12",
			"Stage13",
			"Stage14",
		},
	},
	"glass": {
		Required: []string{
			"Stage1",
			"Stage2",
		},
	},
}
