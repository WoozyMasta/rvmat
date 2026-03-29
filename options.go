// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"os"
	"slices"
	"strings"
)

// TexturePathMode controls texture path existence behavior in validation.
type TexturePathMode string

const (
	// TexturePathModeStrict validates texture file existence via filesystem.
	TexturePathModeStrict TexturePathMode = "strict"

	// TexturePathModeTrust trusts known game-root prefixes as existing.
	TexturePathModeTrust TexturePathMode = "trust"

	// TexturePathModeIgnore skips texture file existence validation.
	TexturePathModeIgnore TexturePathMode = "ignore"
)

var defaultTrustedPrefixes = []string{`dz\`, `ca\`, `a3\`}

// ParseOptions controls parsing behavior.
type ParseOptions struct {
	// DisableCaseInsensitive disables case-insensitive matching for known keys and class names.
	DisableCaseInsensitive bool
	// DisableComments disables // and /* */ comments.
	DisableComments bool
	// DisableRelaxedNumbers disables non-numeric tokens in numeric arrays (parsed as 0).
	// Useful for stats/analysis on files with expression strings in arrays.
	DisableRelaxedNumbers bool
}

// FormatOptions controls writer formatting.
type FormatOptions struct {
	// Indent is the indentation string for nested blocks (default is four spaces).
	Indent string
	// CompactStages writes StageN classes in one line when they only contain
	// texture and texGen assignments.
	CompactStages bool
}

// ValidateOptions controls validation rules.
type ValidateOptions struct {
	// GameRoot is used to resolve texture paths when file checks are enabled.
	// For example, if GameRoot is "P:\\", and the texture path is "dz\vehicles\wheeled\offroad_02\data\offroad_02_roof_co.paa",
	GameRoot string
	// TexturePathMode controls file existence behavior (strict/trust/ignore).
	TexturePathMode TexturePathMode
	// TrustedPrefixes lists trusted game-root prefixes for trust mode.
	// Empty value uses built-in defaults: dz\, ca\, a3\.
	TrustedPrefixes []string
	// ExcludePaths skips file existence checks for matching texture paths.
	// Supports exact match and prefix wildcard with '*' suffix (e.g. "dz\vehicles\*").
	ExcludePaths []string
	// AllowedTextureExtensions overrides allowed texture extension list for
	// CodeValidateUnexpectedTextureExtension checks.
	// Empty value uses built-in defaults: .paa, .pax, .tga, .png.
	AllowedTextureExtensions []string
	// DisableExtensionsCheck disables extension validation for texture paths.
	DisableExtensionsCheck bool
	// DisableShaderNameCheck disables validation of PixelShaderID, VertexShaderID, and Stage names
	// against known lists from validate_lists.go.
	DisableShaderNameCheck bool
	// EnableShaderProfileCheck enables soft stage profile checks for known shaders.
	EnableShaderProfileCheck bool
}

// TextureValidateOptions controls validation of procedural textures.
type TextureValidateOptions struct {
	// AllowedTextureTags overrides allowed texture tags for color(...,tag)
	// validation (RVMAT2028).
	// Empty value uses built-in known tags from validate_lists.go.
	AllowedTextureTags []string
	// DisableProceduralFnCheck disables validation of procedural function names (color, fresnel, etc).
	DisableProceduralFnCheck bool
	// DisableProceduralArgsCheck disables argument count validation for known procedural functions.
	DisableProceduralArgsCheck bool
	// DisableTextureTagCheck disables validation of known texture tags for color(...,tag) arguments.
	DisableTextureTagCheck bool
}

// NormalizeOptions controls material normalization behavior.
type NormalizeOptions struct {
	// StageTextures fills fallback textures for empty known stage slots.
	StageTextures bool
	// StageOrder sorts stages in canonical order.
	StageOrder bool
	// TexGenOrder sorts texgens in canonical order.
	TexGenOrder bool
	// TexturePaths normalizes stage texture paths to game-style form.
	TexturePaths bool
}

// IsGameRootExist reports whether the game root exists and is a directory.
func (o *ValidateOptions) IsGameRootExist() bool {
	if o == nil {
		return false
	}
	if strings.TrimSpace(o.GameRoot) == "" {
		return false
	}
	info, err := os.Stat(o.GameRoot)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// normalize normalizes the ParseOptions.
func (o *ParseOptions) normalize() ParseOptions {
	if o == nil {
		return ParseOptions{}
	}

	return *o
}

// normalize normalizes the FormatOptions.
func (o *FormatOptions) normalize() FormatOptions {
	if o == nil {
		return FormatOptions{Indent: "    "}
	}

	out := *o
	if out.Indent == "" {
		out.Indent = "    "
	}

	return out
}

// normalize normalizes the ValidateOptions.
func (o *ValidateOptions) normalize() ValidateOptions {
	if o == nil {
		return ValidateOptions{
			TexturePathMode: TexturePathModeIgnore,
			TrustedPrefixes: slices.Clone(defaultTrustedPrefixes),
		}
	}

	out := *o
	if !isTexturePathModeValid(out.TexturePathMode) {
		if out.GameRoot == "" {
			out.TexturePathMode = TexturePathModeIgnore
		} else {
			out.TexturePathMode = TexturePathModeStrict
		}
	}
	if len(out.TrustedPrefixes) == 0 {
		out.TrustedPrefixes = slices.Clone(defaultTrustedPrefixes)
	}
	if len(out.AllowedTextureExtensions) == 0 {
		out.AllowedTextureExtensions = slices.Clone(defaultTextureExtensions)
	}

	return out
}

// isTexturePathModeValid reports whether value is a supported path mode.
func isTexturePathModeValid(value TexturePathMode) bool {
	switch value {
	case TexturePathModeStrict, TexturePathModeTrust, TexturePathModeIgnore:
		return true
	default:
		return false
	}
}

// normalize normalizes the TextureValidateOptions.
func (o *TextureValidateOptions) normalize() TextureValidateOptions {
	if o == nil {
		return TextureValidateOptions{}
	}

	return *o
}

// normalize normalizes the NormalizeOptions.
func (o *NormalizeOptions) normalize() NormalizeOptions {
	if o == nil {
		return NormalizeOptions{
			StageTextures: true,
			StageOrder:    true,
			TexGenOrder:   true,
			TexturePaths:  true,
		}
	}

	return *o
}
