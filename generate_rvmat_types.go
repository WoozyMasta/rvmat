// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import "fmt"

// TextureAutoFillMode controls how stage textures are auto-discovered from disk.
type TextureAutoFillMode uint8

const (
	// TextureAutoFillModeDefault applies default auto-fill behavior.
	TextureAutoFillModeDefault TextureAutoFillMode = iota
	// TextureAutoFillModeDisabled disables disk auto-fill.
	TextureAutoFillModeDisabled
	// TextureAutoFillModeFromBaseTexture uses BaseTexture as stem source.
	TextureAutoFillModeFromBaseTexture
	// TextureAutoFillModeFromStageOverride uses stage override texture as stem source.
	TextureAutoFillModeFromStageOverride
)

// StageTextureSource describes how a stage texture was resolved.
type StageTextureSource string

const (
	// StageTextureSourceExplicit is a user-provided override.
	StageTextureSourceExplicit StageTextureSource = "explicit"
	// StageTextureSourceAutoFill is discovered on disk by stem/suffix.
	StageTextureSourceAutoFill StageTextureSource = "autofill"
	// StageTextureSourceDerived is generated from BaseTexture suffix rewrite.
	StageTextureSourceDerived StageTextureSource = "derived"
	// StageTextureSourceProcedural is procedural/default generated fallback.
	StageTextureSourceProcedural StageTextureSource = "procedural"
)

// GenerateSetOptions configures top-level rvmat generation orchestration.
type GenerateSetOptions struct {
	// TextureOverrides overrides stage or role textures
	// (stage1..stage7, nohq/dt/mc/as/smdi/env/fresnel).
	// Stage keys have priority over role keys.
	TextureOverrides map[string]string `json:"texture_overrides,omitempty" yaml:"texture_overrides,omitempty"`
	// OutputPath is target path for resulting main rvmat.
	OutputPath string `json:"output_path,omitempty" yaml:"output_path,omitempty"`
	// BaseTexture is source texture path used for auto-fill and derivation.
	BaseTexture string `json:"base_texture,omitempty" yaml:"base_texture,omitempty"`
	// DamageMacroTexture overrides Stage3 macro texture for damage variant.
	DamageMacroTexture string `json:"damage_macro_texture,omitempty" yaml:"damage_macro_texture,omitempty"`
	// DestructMacroTexture overrides Stage3 macro texture for destruct variant.
	DestructMacroTexture string `json:"destruct_macro_texture,omitempty" yaml:"destruct_macro_texture,omitempty"`
	// EmissiveIntensity sets emissive RGB for generated material when > 0.
	EmissiveIntensity float64 `json:"emissive_intensity,omitempty" yaml:"emissive_intensity,omitempty"`
	// BaseMaterial selects generation material profile.
	BaseMaterial BaseMaterial `json:"base_material,omitempty" yaml:"base_material,omitempty"`
	// Condition applies surface condition modifier.
	Condition Condition `json:"condition,omitempty" yaml:"condition,omitempty"`
	// Finish applies finish modifier.
	Finish Finish `json:"finish,omitempty" yaml:"finish,omitempty"`
	// TextureAutoFillMode controls disk lookup strategy.
	TextureAutoFillMode TextureAutoFillMode `json:"auto_fill_mode,omitempty" yaml:"auto_fill_mode,omitempty"`
	// ForceProceduralOnly disables disk lookup and base suffix derivation.
	ForceProceduralOnly bool `json:"force_procedural_only,omitempty" yaml:"force_procedural_only,omitempty"`
	// GenerateDamage enables damage variant generation.
	GenerateDamage bool `json:"generate_damage,omitempty" yaml:"generate_damage,omitempty"`
	// GenerateDestruct enables destruct variant generation.
	GenerateDestruct bool `json:"generate_destruct,omitempty" yaml:"generate_destruct,omitempty"`
	// DisableDamage disables default damage variant generation.
	DisableDamage bool `json:"disable_damage,omitempty" yaml:"disable_damage,omitempty"`
	// DisableDestruct disables default destruct variant generation.
	DisableDestruct bool `json:"disable_destruct,omitempty" yaml:"disable_destruct,omitempty"`
	// DisableTexGen disables compact TexGen generation.
	DisableTexGen bool `json:"disable_texgen,omitempty" yaml:"disable_texgen,omitempty"`
}

// GenerateStageResolution describes resolved texture for one stage.
type GenerateStageResolution struct {
	// Role is canonical role key (nohq/dt/mc/as/smdi/env/fresnel).
	Role string `json:"role,omitempty" yaml:"role,omitempty"`
	// Source tells whether texture is explicit/autofill/derived/procedural.
	Source StageTextureSource `json:"source,omitempty" yaml:"source,omitempty"`
	// Texture is resolved stage texture.
	Texture TextureRef `json:"texture" yaml:"texture"`
}

// GenerateSetResult is top-level generation output.
type GenerateSetResult struct {
	// StageResolutions contains stage texture resolution report by stage name.
	StageResolutions map[string]GenerateStageResolution `json:"stage_resolutions,omitempty" yaml:"stage_resolutions,omitempty"`
	// Main is generated base material.
	Main *Material `json:"main,omitempty" yaml:"main,omitempty"`
	// Damage is generated damage variant (optional).
	Damage *Material `json:"damage,omitempty" yaml:"damage,omitempty"`
	// Destruct is generated destruct variant (optional).
	Destruct *Material `json:"destruct,omitempty" yaml:"destruct,omitempty"`
	// MainOutputPath is output path for main material.
	MainOutputPath string `json:"main_output_path,omitempty" yaml:"main_output_path,omitempty"`
	// DamageOutputPath is output path for damage material.
	DamageOutputPath string `json:"damage_output_path,omitempty" yaml:"damage_output_path,omitempty"`
	// DestructOutputPath is output path for destruct material.
	DestructOutputPath string `json:"destruct_output_path,omitempty" yaml:"destruct_output_path,omitempty"`
}

// String returns human-readable auto-fill mode name.
func (m TextureAutoFillMode) String() string {
	switch m {
	case TextureAutoFillModeDefault:
		return "default"
	case TextureAutoFillModeDisabled:
		return "disabled"
	case TextureAutoFillModeFromBaseTexture:
		return "from_base_texture"
	case TextureAutoFillModeFromStageOverride:
		return "from_stage_override"
	default:
		return fmt.Sprintf("auto_fill_mode(%d)", m)
	}
}
