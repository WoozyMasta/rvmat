// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"slices"
	"strconv"
	"strings"
)

// Normalize applies safe normalization directly to material.
func Normalize(m *Material, opt *NormalizeOptions) (NormalizeResult, []Issue) {
	normalizeOptions := opt.normalize()
	if m == nil {
		return NormalizeResult{}, []Issue{issueError(
			CodeNormalizeNilMaterial,
			"normalize failed: material is nil",
			"",
		)}
	}

	var result NormalizeResult
	var out []Issue

	if normalizeOptions.StageTextures {
		fixed := fillMissingStageTexture(m)
		result.StageTexturesFilled = fixed
		if fixed > 0 {
			result.Changed = true
		}
	}

	if normalizeOptions.TexturePaths {
		fixed := normalizeStageTexturePaths(m)
		result.TexturePathsNormalized = fixed
		if fixed > 0 {
			result.Changed = true
		}
	}

	if normalizeOptions.StageOrder {
		if normalizeStageOrder(m) {
			result.StageOrderNormalized = true
			result.Changed = true
		}
	}

	if normalizeOptions.TexGenOrder {
		if normalizeTexGenOrder(m) {
			result.TexGenOrderNormalized = true
			result.Changed = true
		}
	}

	return result, out
}

// fillMissingStageTexture assigns fallback procedural textures for known stage roles.
func fillMissingStageTexture(m *Material) int {
	var fixed int
	for i := range m.Stages {
		st := &m.Stages[i]
		if strings.TrimSpace(st.Texture.Raw) != "" {
			continue
		}

		fallback, ok := fallbackTextureForStage(st.Name)
		if !ok {
			continue
		}

		st.Texture = fallback
		fixed++
	}

	return fixed
}

// normalizeStageTexturePaths normalizes stage texture paths to game-style form.
func normalizeStageTexturePaths(m *Material) int {
	var fixed int
	for i := range m.Stages {
		st := &m.Stages[i]
		if strings.TrimSpace(st.Texture.Raw) == "" {
			continue
		}

		normalized := NormalizeGameTexturePath(st.Texture.Raw)
		if normalized == st.Texture.Raw {
			continue
		}

		st.Texture = ParseTextureRef(normalized)
		fixed++
	}

	return fixed
}

// fallbackTextureForStage returns fallback texture for known stage name.
func fallbackTextureForStage(stageName string) (TextureRef, bool) {
	tex, ok := defaultStageFallbackTextures[strings.ToLower(strings.TrimSpace(stageName))]
	return tex, ok
}

// normalizeStageOrder sorts stages by canonical order.
func normalizeStageOrder(m *Material) bool {
	if len(m.Stages) < 2 {
		return false
	}

	items := make([]stageSortableItem, len(m.Stages))
	for i := range m.Stages {
		items[i] = stageSortableItem{index: i, stage: m.Stages[i]}
	}

	slices.SortStableFunc(items, func(a, b stageSortableItem) int {
		return compareStageNames(a.stage.Name, b.stage.Name)
	})

	changed := false
	for i := range items {
		if items[i].index != i {
			changed = true
		}
		m.Stages[i] = items[i].stage
	}

	return changed
}

// normalizeTexGenOrder sorts texgens by canonical order.
func normalizeTexGenOrder(m *Material) bool {
	if len(m.TexGens) < 2 {
		return false
	}

	items := make([]texGenSortableItem, len(m.TexGens))
	for i := range m.TexGens {
		items[i] = texGenSortableItem{index: i, texGen: m.TexGens[i]}
	}

	slices.SortStableFunc(items, func(a, b texGenSortableItem) int {
		return compareTexGenNames(a.texGen.Name, b.texGen.Name)
	})

	changed := false
	for i := range items {
		if items[i].index != i {
			changed = true
		}
		m.TexGens[i] = items[i].texGen
	}

	return changed
}

// compareStageNames compares stage names in canonical order.
func compareStageNames(a, b string) int {
	ag, an, al := stageNameOrderKey(a)
	bg, bn, bl := stageNameOrderKey(b)

	if ag != bg {
		return ag - bg
	}
	if an != bn {
		return an - bn
	}

	return strings.Compare(al, bl)
}

// compareTexGenNames compares texgen names in canonical order.
func compareTexGenNames(a, b string) int {
	ag, an, al := indexedNameOrderKey(a, "texgen")
	bg, bn, bl := indexedNameOrderKey(b, "texgen")

	if ag != bg {
		return ag - bg
	}
	if an != bn {
		return an - bn
	}

	return strings.Compare(al, bl)
}

// stageNameOrderKey computes ordering key for stage names.
func stageNameOrderKey(name string) (group, num int, alpha string) {
	if n, ok := parseIndexedName(name, "stage"); ok {
		return 0, n, ""
	}
	if strings.EqualFold(strings.TrimSpace(name), "StageTI") {
		return 1, 0, ""
	}

	return 2, 0, strings.ToLower(strings.TrimSpace(name))
}

// indexedNameOrderKey computes ordering key for indexed names.
func indexedNameOrderKey(name, prefix string) (group, num int, alpha string) {
	if n, ok := parseIndexedName(name, prefix); ok {
		return 0, n, ""
	}

	return 1, 0, strings.ToLower(strings.TrimSpace(name))
}

// parseIndexedName parses names in form "<prefix><number>".
func parseIndexedName(name, prefix string) (int, bool) {
	s := strings.TrimSpace(name)
	if len(s) <= len(prefix) {
		return 0, false
	}
	if !strings.EqualFold(s[:len(prefix)], prefix) {
		return 0, false
	}

	n, err := strconv.Atoi(s[len(prefix):])
	if err != nil || n < 0 {
		return 0, false
	}

	return n, true
}
