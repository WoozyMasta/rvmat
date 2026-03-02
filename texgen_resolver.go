// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolveStageTexGen resolves effective TexGen for a stage with inheritance.
func ResolveStageTexGen(m *Material, stage Stage) (*TexGen, error) {
	if m == nil {
		return nil, nil
	}

	if strings.TrimSpace(stage.TexGen) == "" {
		return nil, nil
	}

	leaf, ok := findTexGenByRef(m.TexGens, stage.TexGen)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTexGenNotFound, stage.TexGen)
	}

	chain, err := resolveTexGenChain(m.TexGens, leaf.Name)
	if err != nil {
		return nil, err
	}

	effective := TexGen{Name: leaf.Name, Base: leaf.Base}
	for i := len(chain) - 1; i >= 0; i-- {
		tg := chain[i]
		if tg.UVSource != "" {
			effective.UVSource = tg.UVSource
		}
		if tg.UVTransform != nil {
			effective.UVTransform = mergeUVTransforms(effective.UVTransform, tg.UVTransform)
		}
	}

	return &effective, nil
}

// EffectiveUVSource returns effective uvSource for a stage.
func EffectiveUVSource(m *Material, stage Stage) (string, error) {
	if stage.TexGen == "" {
		return stage.UVSource, nil
	}

	resolved, err := ResolveStageTexGen(m, stage)
	if err != nil {
		return "", err
	}
	if resolved == nil {
		return "", nil
	}

	return resolved.UVSource, nil
}

// EffectiveUVTransform returns effective uvTransform for a stage.
func EffectiveUVTransform(m *Material, stage Stage) (*UVTransform, error) {
	if stage.TexGen == "" {
		if stage.UVTransform == nil {
			return nil, nil
		}

		c := cloneUVTransform(*stage.UVTransform)
		return &c, nil
	}

	resolved, err := ResolveStageTexGen(m, stage)
	if err != nil {
		return nil, err
	}
	if resolved == nil || resolved.UVTransform == nil {
		return nil, nil
	}

	c := cloneUVTransform(*resolved.UVTransform)
	return &c, nil
}

// findTexGenByRef finds texgen by stage ref value (e.g. "0" or "TexGen0").
func findTexGenByRef(texgens []TexGen, ref string) (*TexGen, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, false
	}

	name := ref
	if _, err := strconv.Atoi(ref); err == nil {
		name = "TexGen" + ref
	}

	for i := range texgens {
		if strings.EqualFold(texgens[i].Name, name) {
			return &texgens[i], true
		}
	}

	return nil, false
}

// findTexGenByName finds texgen by exact class-like name.
func findTexGenByName(texgens []TexGen, name string) (*TexGen, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, false
	}

	for i := range texgens {
		if strings.EqualFold(texgens[i].Name, name) {
			return &texgens[i], true
		}
	}

	return nil, false
}

// resolveTexGenChain resolves inheritance chain from leaf to base.
func resolveTexGenChain(texgens []TexGen, leafName string) ([]TexGen, error) {
	chain := make([]TexGen, 0, 4)
	seen := make(map[string]struct{}, 4)
	current := strings.TrimSpace(leafName)

	for current != "" {
		key := strings.ToLower(current)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("%w: %s", ErrTexGenCycle, current)
		}
		seen[key] = struct{}{}

		tg, ok := findTexGenByName(texgens, current)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrTexGenBaseNotFound, current)
		}
		chain = append(chain, *tg)

		current = strings.TrimSpace(tg.Base)
	}

	return chain, nil
}

// mergeUVTransforms overlays override values onto base.
func mergeUVTransforms(base, override *UVTransform) *UVTransform {
	if base == nil && override == nil {
		return nil
	}

	var out UVTransform
	if base != nil {
		out = cloneUVTransform(*base)
	}

	if override == nil {
		return &out
	}

	if len(override.Aside) > 0 {
		out.Aside = cloneFloatSlice(override.Aside)
	}
	if len(override.Up) > 0 {
		out.Up = cloneFloatSlice(override.Up)
	}
	if len(override.Dir) > 0 {
		out.Dir = cloneFloatSlice(override.Dir)
	}
	if len(override.Pos) > 0 {
		out.Pos = cloneFloatSlice(override.Pos)
	}

	return &out
}

// cloneUVTransform deep-copies UV transform.
func cloneUVTransform(in UVTransform) UVTransform {
	return UVTransform{
		Aside: cloneFloatSlice(in.Aside),
		Up:    cloneFloatSlice(in.Up),
		Dir:   cloneFloatSlice(in.Dir),
		Pos:   cloneFloatSlice(in.Pos),
	}
}

// cloneFloatSlice deep-copies float slice.
func cloneFloatSlice(in []float64) []float64 {
	if len(in) == 0 {
		return nil
	}

	out := make([]float64, len(in))
	copy(out, in)
	return out
}
