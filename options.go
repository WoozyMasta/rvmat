package rvmat

import (
	"os"
	"strings"
)

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
}

// ValidateOptions controls validation rules.
type ValidateOptions struct {
	// GameRoot is used to resolve texture paths when file checks are enabled.
	// For example, if GameRoot is "P:\\", and the texture path is "dz\vehicles\wheeled\offroad_02\data\offroad_02_roof_co.paa",
	GameRoot string
	// ExcludePaths skips file existence checks for matching texture paths.
	// Supports exact match and prefix wildcard with '*' suffix (e.g. "dz\vehicles\*").
	ExcludePaths []string
	// DisableFileCheck disables filesystem existence checks for texture paths.
	// If game is not set, this is enabled by default.
	DisableFileCheck bool
	// DisableExtensionsCheck disables extension validation for texture paths.
	DisableExtensionsCheck bool
	// DisableShaderNameCheck disables validation of PixelShaderID, VertexShaderID, and Stage names
	// against known lists from validate_lists.go.
	DisableShaderNameCheck bool
}

// TextureValidateOptions controls validation of procedural textures.
type TextureValidateOptions struct {
	// DisableProceduralFnCheck disables validation of procedural function names (color, fresnel, etc).
	DisableProceduralFnCheck bool
	// DisableProceduralArgsCheck disables argument count validation for known procedural functions.
	DisableProceduralArgsCheck bool
	// DisableTextureTagCheck disables validation of known texture tags for color(...,tag) arguments.
	DisableTextureTagCheck bool
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
		return ValidateOptions{DisableFileCheck: true}
	}

	out := *o
	if out.GameRoot == "" {
		out.DisableFileCheck = true
	}

	return out
}

// normalize normalizes the TextureValidateOptions.
func (o *TextureValidateOptions) normalize() TextureValidateOptions {
	if o == nil {
		return TextureValidateOptions{}
	}

	return *o
}
