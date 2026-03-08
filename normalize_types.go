// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

// defaultStageFallbackTextures defines safe procedural placeholders by stage role.
var defaultStageFallbackTextures = map[string]TextureRef{
	"stage0": NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1.0, "CO"),
	"stage1": NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 1.0, 1.0, "NOHQ"),
	"stage2": NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1.0, "DT"),
	"stage3": NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 0.0, "MC"),
	"stage4": NewProceduralColor("argb", 8, 8, 3, 0.74, 0.74, 0.74, 1.0, "AS"),
	"stage5": NewProceduralColor("argb", 8, 8, 3, 1.0, 0.0944, 0.1743, 1.0, "SMDI"),
	"stage6": NewProceduralFresnel("ai", 64, 1, 1, 1.23, 0.36),
	"stage7": ParseTextureRef(DefaultEnvironmentTextureProcedural),
}

// NormalizeResult reports what was changed by Normalize.
type NormalizeResult struct {
	// StageTexturesFilled is count of stages where missing texture was filled.
	StageTexturesFilled int `json:"stage_textures_filled,omitempty" yaml:"stage_textures_filled,omitempty"`

	// TexturePathsNormalized is count of stage textures normalized to game-style path.
	TexturePathsNormalized int `json:"texture_paths_normalized,omitempty" yaml:"texture_paths_normalized,omitempty"`

	// Changed indicates whether any normalization was applied.
	Changed bool `json:"changed" yaml:"changed"`

	// StageOrderNormalized indicates stage slice reordering.
	StageOrderNormalized bool `json:"stage_order_normalized,omitempty" yaml:"stage_order_normalized,omitempty"`

	// TexGenOrderNormalized indicates texgen slice reordering.
	TexGenOrderNormalized bool `json:"texgen_order_normalized,omitempty" yaml:"texgen_order_normalized,omitempty"`
}

// stageSortableItem carries stage with source index for stable change detection.
type stageSortableItem struct {
	stage Stage
	index int
}

// texGenSortableItem carries texgen with source index for stable change detection.
type texGenSortableItem struct {
	texGen TexGen
	index  int
}
