// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import "fmt"

const (
	// DefaultDamageMacroTexture is default macro map for damage variants.
	DefaultDamageMacroTexture = `dz\characters\data\generic_damage_mc.paa`
	// DefaultDestructMacroTexture is default macro map for destruct variants.
	DefaultDestructMacroTexture = `dz\characters\data\generic_destruct_mc.paa`
	// DefaultEnvironmentTexture is default Stage7 environment texture.
	DefaultEnvironmentTexture = `dz\data\data\env_land_co.paa`
	// DefaultEnvironmentTextureProcedural is optional procedural Stage7 fallback.
	DefaultEnvironmentTextureProcedural = `#(argb,8,8,3)color(4,4,4,1,CA)`
)

// BaseMaterial identifies baseline material family for generator.
type BaseMaterial uint8

const (
	// BaseMaterialDefault applies default baseline profile.
	BaseMaterialDefault BaseMaterial = iota
	// BaseMaterialTextile is baseline for textile and cloth-like surfaces.
	BaseMaterialTextile
	// BaseMaterialSteel is baseline for non-rust steel surfaces.
	BaseMaterialSteel
	// BaseMaterialRust is baseline for oxidized rust surfaces.
	BaseMaterialRust
	// BaseMaterialWood is baseline for wood surfaces.
	BaseMaterialWood
	// BaseMaterialGlass is baseline for glass surfaces.
	BaseMaterialGlass
	// BaseMaterialPlastic is baseline for plastic/polymer surfaces.
	BaseMaterialPlastic
	// BaseMaterialRubber is baseline for rubber surfaces.
	BaseMaterialRubber
	// BaseMaterialLeather is baseline for leather surfaces.
	BaseMaterialLeather
	// BaseMaterialEarth is baseline for ground/soil surfaces.
	BaseMaterialEarth
	// BaseMaterialPaper is baseline for paper/cardboard surfaces.
	BaseMaterialPaper
	// BaseMaterialConcrete is baseline for concrete/plaster surfaces.
	BaseMaterialConcrete
	// BaseMaterialStone is baseline for stone/rock surfaces.
	BaseMaterialStone
	// BaseMaterialSkin is baseline for skin-like organic surfaces.
	BaseMaterialSkin
)

// Finish identifies finish modifier for generator.
type Finish uint8

const (
	// FinishDefault applies default profile multipliers.
	FinishDefault Finish = iota
	// FinishMatte applies matte finish multipliers.
	FinishMatte
	// FinishSatin applies satin finish multipliers.
	FinishSatin
	// FinishGloss applies gloss finish multipliers.
	FinishGloss
	// FinishPolished applies polished finish multipliers.
	FinishPolished
)

// Condition identifies surface condition modifier for generator.
type Condition uint8

const (
	// ConditionDefault applies default profile multipliers.
	ConditionDefault Condition = iota
	// ConditionClean keeps clean material multipliers.
	ConditionClean
	// ConditionWorn applies worn multipliers.
	ConditionWorn
	// ConditionDirty applies dirty multipliers.
	ConditionDirty
	// ConditionOxidized applies oxidized multipliers.
	ConditionOxidized
)

// GenerateOptions configures baseline material generation.
type GenerateOptions struct {
	// TextureOverrides overrides generated textures by stage or role key.
	// Accepted keys are case-insensitive and include:
	// stage1..stage7, and roles nohq/dt/mc/as/smdi/env/fresnel.
	// When both a stage key and role key target the same stage,
	// stage key has priority.
	TextureOverrides map[string]string `json:"texture_overrides,omitempty" yaml:"texture_overrides,omitempty"`
	// BaseTexture is a source texture path used to derive role textures.
	// Example: "my/path/item_co.paa" -> "_nohq/_dt/_mc/_as/_smdi".
	BaseTexture string `json:"base_texture,omitempty" yaml:"base_texture,omitempty"`
	// EmissiveIntensity sets emissive RGB for generated material when > 0.
	EmissiveIntensity float64 `json:"emissive_intensity,omitempty" yaml:"emissive_intensity,omitempty"`
	// BaseMaterial selects generation profile (BaseMaterial* constants).
	BaseMaterial BaseMaterial `json:"base_material,omitempty" yaml:"base_material,omitempty"`
	// Condition applies optional surface condition modifier (Condition* constants).
	Condition Condition `json:"condition,omitempty" yaml:"condition,omitempty"`
	// Finish applies optional surface finish modifier (Finish* constants).
	Finish Finish `json:"finish,omitempty" yaml:"finish,omitempty"`
	// WithDamage switches output to damage variant.
	WithDamage bool `json:"with_damage,omitempty" yaml:"with_damage,omitempty"`
	// WithDestruct switches output to destruct variant.
	WithDestruct bool `json:"with_destruct,omitempty" yaml:"with_destruct,omitempty"`
	// UseTexGen emits Stage texGen references and shared TexGen classes.
	UseTexGen bool `json:"use_texgen,omitempty" yaml:"use_texgen,omitempty"`
}

// String returns human-readable material name.
func (m BaseMaterial) String() string {
	switch m {
	case BaseMaterialDefault:
		return "default"
	case BaseMaterialTextile:
		return "textile"
	case BaseMaterialSteel:
		return "steel"
	case BaseMaterialRust:
		return "rust"
	case BaseMaterialWood:
		return "wood"
	case BaseMaterialGlass:
		return "glass"
	case BaseMaterialPlastic:
		return "plastic"
	case BaseMaterialRubber:
		return "rubber"
	case BaseMaterialLeather:
		return "leather"
	case BaseMaterialEarth:
		return "earth"
	case BaseMaterialPaper:
		return "paper"
	case BaseMaterialConcrete:
		return "concrete"
	case BaseMaterialStone:
		return "stone"
	case BaseMaterialSkin:
		return "skin"
	default:
		return fmt.Sprintf("base_material(%d)", m)
	}
}

// String returns human-readable finish name.
func (f Finish) String() string {
	switch f {
	case FinishDefault:
		return "default"
	case FinishMatte:
		return "matte"
	case FinishSatin:
		return "satin"
	case FinishGloss:
		return "gloss"
	case FinishPolished:
		return "polished"
	default:
		return fmt.Sprintf("finish(%d)", f)
	}
}

// String returns human-readable condition name.
func (c Condition) String() string {
	switch c {
	case ConditionDefault:
		return "default"
	case ConditionClean:
		return "clean"
	case ConditionWorn:
		return "worn"
	case ConditionDirty:
		return "dirty"
	case ConditionOxidized:
		return "oxidized"
	default:
		return fmt.Sprintf("condition(%d)", c)
	}
}

// normalizeBaseMaterial validates and normalizes base material value.
func normalizeBaseMaterial(material BaseMaterial) (BaseMaterial, error) {
	switch material {
	case BaseMaterialDefault:
		return BaseMaterialTextile, nil
	case BaseMaterialTextile,
		BaseMaterialSteel,
		BaseMaterialRust,
		BaseMaterialWood,
		BaseMaterialGlass,
		BaseMaterialPlastic,
		BaseMaterialRubber,
		BaseMaterialLeather,
		BaseMaterialEarth,
		BaseMaterialPaper,
		BaseMaterialConcrete,
		BaseMaterialStone,
		BaseMaterialSkin:
		return material, nil
	default:
		return 0, fmt.Errorf(
			"%w material=%s",
			ErrUnknownBaseMaterial,
			material,
		)
	}
}

// normalizeFinish validates and normalizes finish value.
func normalizeFinish(finish Finish) (Finish, error) {
	switch finish {
	case FinishDefault, FinishSatin:
		return FinishDefault, nil
	case FinishMatte:
		return FinishMatte, nil
	case FinishGloss:
		return FinishGloss, nil
	case FinishPolished:
		return FinishPolished, nil
	default:
		return 0, fmt.Errorf("%w finish=%s", ErrInvalidGenerateOption, finish)
	}
}

// normalizeCondition validates and normalizes condition value.
func normalizeCondition(condition Condition) (Condition, error) {
	switch condition {
	case ConditionDefault, ConditionClean:
		return ConditionDefault, nil
	case ConditionWorn:
		return ConditionWorn, nil
	case ConditionDirty:
		return ConditionDirty, nil
	case ConditionOxidized:
		return ConditionOxidized, nil
	default:
		return 0, fmt.Errorf("%w condition=%s", ErrInvalidGenerateOption, condition)
	}
}
