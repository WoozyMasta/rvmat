// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Generate generates baseline material from options.
func Generate(opts GenerateOptions) (*Material, error) {
	if opts.WithDamage && opts.WithDestruct {
		return nil, errors.New("generate material: both WithDamage and WithDestruct are set")
	}

	baseMaterial, err := normalizeBaseMaterial(opts.BaseMaterial)
	if err != nil {
		return nil, fmt.Errorf("generate material: %w", err)
	}

	finish, err := normalizeFinish(opts.Finish)
	if err != nil {
		return nil, fmt.Errorf("generate material: %w", err)
	}

	condition, err := normalizeCondition(opts.Condition)
	if err != nil {
		return nil, fmt.Errorf("generate material: %w", err)
	}

	seed, ok := materialCatalog[baseMaterial]
	if !ok {
		return nil, fmt.Errorf(
			"generate material: unknown base material %s: %w",
			baseMaterial,
			ErrUnknownBaseMaterial,
		)
	}

	specular, power := applyMaterialModifiers(seed, finish, condition)
	emissive := seed.emissive
	if opts.EmissiveIntensity > 0 {
		emissive[0] = opts.EmissiveIntensity
		emissive[1] = opts.EmissiveIntensity
		emissive[2] = opts.EmissiveIntensity
	}

	m := &Material{
		Ambient:        []float64{1, 1, 1, 1},
		Diffuse:        []float64{1, 1, 1, 1},
		ForcedDiffuse:  []float64{0, 0, 0, 0},
		Emissive:       emissive[:],
		Specular:       specular[:],
		PixelShaderID:  "Super",
		VertexShaderID: "Super",
	}
	m.SpecularPower = &power

	m.Stages = generateSuperStages(seed, specular, power, opts)
	if opts.UseTexGen {
		m.TexGens = generateSuperTexGens()
	}

	if opts.WithDamage {
		return GenerateDamage(m)
	}
	if opts.WithDestruct {
		return GenerateDestruct(m)
	}

	return m, nil
}

// GenerateDamage creates damage material variant from base.
func GenerateDamage(base *Material) (*Material, error) {
	return generateVariantWithMacro(base, DefaultDamageMacroTexture)
}

// GenerateDestruct creates destruct material variant from base.
func GenerateDestruct(base *Material) (*Material, error) {
	return generateVariantWithMacro(base, DefaultDestructMacroTexture)
}

// applyMaterialModifiers applies finish and condition multipliers to material seed values.
func applyMaterialModifiers(seed materialSeed, finish Finish, condition Condition) ([4]float64, float64) {
	specular := seed.specular
	power := seed.specularPower

	finishMultS, finishMultP := finishModifier(finish)
	conditionMultS, conditionMultP := conditionModifier(condition)

	for i := range 3 {
		specular[i] *= finishMultS * conditionMultS
	}
	power *= finishMultP * conditionMultP
	if power > 1000 {
		power = 1000
	}

	return specular, power
}

// finishModifier returns multipliers for finish option.
func finishModifier(finish Finish) (specularMult, powerMult float64) {
	switch finish {
	case FinishMatte:
		return 0.7, 0.55
	case FinishGloss:
		return 1.15, 1.6
	case FinishPolished:
		return 1.25, 2.2
	case FinishSatin, FinishDefault:
		return 1.0, 1.0
	default:
		return 1.0, 1.0
	}
}

// conditionModifier returns multipliers for condition option.
func conditionModifier(condition Condition) (specularMult, powerMult float64) {
	switch condition {
	case ConditionWorn:
		return 0.9, 0.8
	case ConditionDirty:
		return 0.75, 0.65
	case ConditionOxidized:
		return 0.7, 0.7
	case ConditionClean, ConditionDefault:
		return 1.0, 1.0
	default:
		return 1.0, 1.0
	}
}

// generateSuperStages builds baseline Super stage set.
func generateSuperStages(seed materialSeed, specular [4]float64, power float64, opts GenerateOptions) []Stage {
	stages := make([]Stage, 0, 7)
	for i := 1; i <= 5; i++ {
		stageName := fmt.Sprintf("Stage%d", i)
		role, _ := textureRoleForStageName(stageName)
		tex := textureForRole(opts, stageName, role, seed, specular, power)
		stage := Stage{Name: stageName, Texture: tex}
		applyStageUV(&stage, opts.UseTexGen)
		stages = append(stages, stage)
	}

	fresnel := NewProceduralFresnel("ai", 64, 1, 1, seed.fresnelA, seed.fresnelB)
	if seed.fresnelGlass {
		fresnel = NewProceduralFresnelGlass(
			"ai",
			64,
			1,
			1,
			seed.fresnelA,
			0,
			false,
		)
	}

	stage6 := Stage{Name: "Stage6", Texture: fresnel}
	applyStageUV(&stage6, opts.UseTexGen)
	stages = append(stages, stage6)

	stage7 := Stage{
		Name:    "Stage7",
		Texture: textureForRole(opts, "Stage7", "env", seed, specular, power),
	}
	applyStageUV(&stage7, opts.UseTexGen)
	stages = append(stages, stage7)

	return stages
}

// textureForRole resolves texture for stage role from overrides, derived base, and fallback.
func textureForRole(opts GenerateOptions, stageName, role string, seed materialSeed, specular [4]float64, power float64) TextureRef {
	if raw := textureOverride(opts.TextureOverrides, stageName, role); raw != "" {
		return ParseTextureRef(raw)
	}
	if raw := derivedTextureForRole(opts.BaseTexture, role); raw != "" {
		return ParseTextureRef(raw)
	}

	return fallbackTextureForRole(role, seed, specular, power)
}

// textureOverride gets role or stage override texture.
func textureOverride(overrides map[string]string, stageName, role string) string {
	if len(overrides) == 0 {
		return ""
	}

	stageKey := strings.ToLower(strings.TrimSpace(stageName))
	roleKey := strings.ToLower(strings.TrimSpace(role))
	if raw := strings.TrimSpace(overrides[stageKey]); raw != "" {
		return raw
	}
	if raw := strings.TrimSpace(overrides[roleKey]); raw != "" {
		return raw
	}

	return ""
}

// derivedTextureForRole derives role texture path from base texture stem.
func derivedTextureForRole(baseTexture, role string) string {
	stem, ext, ok := splitBaseTextureStem(baseTexture)
	if !ok {
		return ""
	}
	if role == "env" {
		return ""
	}

	return stem + "_" + role + ext
}

// splitBaseTextureStem extracts stem and extension from a base texture path.
func splitBaseTextureStem(baseTexture string) (string, string, bool) {
	raw := strings.TrimSpace(baseTexture)
	if raw == "" {
		return "", "", false
	}

	ext := ".paa"
	lastDot := strings.LastIndex(raw, ".")
	lastSlash := strings.LastIndexAny(raw, `/\`)
	if lastDot > lastSlash {
		ext = raw[lastDot:]
		raw = raw[:lastDot]
	}

	lower := strings.ToLower(raw)
	for _, suffix := range []string{
		"_co", "_ca", "_cat",
		"_raw",
		"_lco", "_mco", "_lca",
		"_mask",
		"_pr",
		"_dtsmdi", "_smdi", "_sm", "_detail", "_cdt", "_dt",
		"_adshq", "_ads", "_as",
		"_mc",
		"_normalmap", "_nshq", "_novhq", "_nofhq", "_nofex", "_nof", "_nopx", "_nsex", "_noex", "_nohq", "_non", "_ns", "_no",
		"_sky",
		"_ti",
	} {
		if strings.HasSuffix(lower, suffix) {
			return raw[:len(raw)-len(suffix)], ext, true
		}
	}

	return raw, ext, true
}

// fallbackTextureForRole returns fallback texture for role.
func fallbackTextureForRole(role string, seed materialSeed, specular [4]float64, power float64) TextureRef {
	switch role {
	case "nohq":
		return NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 1, 1, "nohq")
	case "dt":
		return NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1, "dt")
	case "mc":
		return NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1, "mc")
	case "as":
		ao := defaultASFromMaterialClass(seed.materialClass)
		return NewProceduralColor("argb", 8, 8, 3, ao, ao, ao, 1, "as")
	case "smdi":
		specMean := (specular[0] + specular[1] + specular[2]) / 3
		specRatio := clamp01(specMean / 0.35)
		glossRatio := glossRatioFromSpecularPower(power)
		gMin, gMax, bMin, bMax := defaultSMDIRangeFromMaterialClass(seed.materialClass)
		g := gMin + (gMax-gMin)*specRatio
		b := bMin + (bMax-bMin)*glossRatio
		return NewProceduralColor("argb", 8, 8, 3, 1, clamp01(g), clamp01(b), 1, "smdi")
	case "env":
		return ParseTextureRef(DefaultEnvironmentTexture)
	default:
		return TextureRef{}
	}
}

// applyStageUV sets UV fields for generated stages.
func applyStageUV(st *Stage, useTexGen bool) {
	if useTexGen {
		st.TexGen = "0"
		return
	}

	st.UVSource = "tex"
	st.UVTransform = &UVTransform{
		Aside: []float64{1, 0, 0},
		Up:    []float64{0, 1, 0},
		Dir:   []float64{0, 0, 0},
		Pos:   []float64{0, 0, 0},
	}
}

// generateSuperTexGens creates canonical TexGen0..7 set for Super shader.
func generateSuperTexGens() []TexGen {
	base := TexGen{
		Name:     "TexGen0",
		UVSource: "tex",
		UVTransform: &UVTransform{
			Aside: []float64{1, 0, 0},
			Up:    []float64{0, 1, 0},
			Dir:   []float64{0, 0, 0},
			Pos:   []float64{0, 0, 0},
		},
	}

	return []TexGen{base}
}

// generateVariantWithMacro clones base and sets Stage3 macro map texture.
func generateVariantWithMacro(base *Material, macroRaw string) (*Material, error) {
	if base == nil {
		return nil, fmt.Errorf("generate variant: nil material: %w", ErrMaterialNotFound)
	}

	out := cloneMaterial(base)
	stage := findMaterialStageByName(out, "Stage3")
	if stage == nil {
		return nil, fmt.Errorf("generate variant: Stage3 not found: %w", ErrStageNotFound)
	}

	stage.Texture = ParseTextureRef(macroRaw)
	return out, nil
}

// cloneMaterial deep-copies material for safe variant generation.
func cloneMaterial(in *Material) *Material {
	if in == nil {
		return nil
	}

	out := *in
	out.Ambient = cloneFloatSlice(in.Ambient)
	out.Diffuse = cloneFloatSlice(in.Diffuse)
	out.ForcedDiffuse = cloneFloatSlice(in.ForcedDiffuse)
	out.Emissive = cloneFloatSlice(in.Emissive)
	out.Specular = cloneFloatSlice(in.Specular)
	if in.SpecularPower != nil {
		p := *in.SpecularPower
		out.SpecularPower = &p
	}

	out.Stages = make([]Stage, len(in.Stages))
	for i := range in.Stages {
		out.Stages[i] = cloneStage(in.Stages[i])
	}

	out.TexGens = make([]TexGen, len(in.TexGens))
	for i := range in.TexGens {
		out.TexGens[i] = cloneTexGen(in.TexGens[i])
	}

	out.extras = append([]node(nil), in.extras...)
	return &out
}

// cloneStage deep-copies stage.
func cloneStage(in Stage) Stage {
	out := in
	out.Texture = cloneTextureRef(in.Texture)
	if in.UVTransform != nil {
		uv := cloneUVTransform(*in.UVTransform)
		out.UVTransform = &uv
	}
	out.extras = append([]node(nil), in.extras...)
	return out
}

// cloneTexGen deep-copies texgen.
func cloneTexGen(in TexGen) TexGen {
	out := in
	if in.UVTransform != nil {
		uv := cloneUVTransform(*in.UVTransform)
		out.UVTransform = &uv
	}
	out.extras = append([]node(nil), in.extras...)
	return out
}

// cloneTextureRef deep-copies texture ref.
func cloneTextureRef(in TextureRef) TextureRef {
	out := in
	if in.Procedural != nil {
		pt := *in.Procedural
		pt.Args = append([]string(nil), in.Procedural.Args...)
		if in.Procedural.Color != nil {
			color := *in.Procedural.Color
			pt.Color = &color
		}
		if in.Procedural.Fresnel != nil {
			fresnel := *in.Procedural.Fresnel
			pt.Fresnel = &fresnel
		}
		if in.Procedural.Irradiance != nil {
			irr := *in.Procedural.Irradiance
			pt.Irradiance = &irr
		}
		out.Procedural = &pt
	}

	return out
}

// findMaterialStageByName finds stage by name (case-insensitive).
func findMaterialStageByName(m *Material, stageName string) *Stage {
	if m == nil {
		return nil
	}

	for i := range m.Stages {
		if strings.EqualFold(strings.TrimSpace(m.Stages[i].Name), strings.TrimSpace(stageName)) {
			return &m.Stages[i]
		}
	}

	return nil
}

// defaultASFromMaterialClass returns fallback AO value for _AS map.
//
// _AS uses ambient information (mostly G channel in engine practice):
// 1.0 means fully lit ambient, lower values mean stronger occlusion.
func defaultASFromMaterialClass(class materialClass) float64 {
	switch class {
	case materialClassGlass:
		return 0.96
	case materialClassMetal:
		return 0.90
	case materialClassPolymer:
		return 0.82
	case materialClassWood:
		return 0.76
	case materialClassLeather:
		return 0.72
	case materialClassTerrain:
		return 0.66
	case materialClassMineral:
		return 0.70
	case materialClassTextile:
		return 0.74
	case materialClassOrganic:
		return 0.78
	default:
		return 0.82
	}
}

// defaultSMDIRangeFromMaterialClass returns SMDI channel ranges by material family.
//
// SMDI semantics:
//   - R: diffuse inverse (kept at 1 in fallback).
//   - G: specular intensity.
//   - B: glossiness multiplier.
func defaultSMDIRangeFromMaterialClass(class materialClass) (gMin, gMax, bMin, bMax float64) {
	switch class {
	case materialClassGlass:
		return 0.35, 0.85, 0.45, 0.95
	case materialClassMetal:
		return 0.10, 0.55, 0.20, 0.85
	case materialClassPolymer:
		return 0.05, 0.32, 0.15, 0.60
	case materialClassWood:
		return 0.04, 0.24, 0.10, 0.45
	case materialClassLeather:
		return 0.06, 0.28, 0.12, 0.50
	case materialClassTerrain:
		return 0.02, 0.18, 0.05, 0.30
	case materialClassMineral:
		return 0.03, 0.22, 0.08, 0.38
	case materialClassTextile:
		return 0.03, 0.20, 0.08, 0.35
	case materialClassOrganic:
		return 0.08, 0.30, 0.16, 0.50
	default:
		return 0.05, 0.35, 0.12, 0.55
	}
}

// glossRatioFromSpecularPower maps rvmat specularPower to normalized gloss ratio.
func glossRatioFromSpecularPower(power float64) float64 {
	normalized := clamp01(power / 1000)
	return clamp01(0.15 + 0.85*math.Sqrt(normalized))
}

// clamp01 clamps value to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}

	return v
}
