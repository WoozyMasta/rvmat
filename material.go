// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

// Material represents a parsed RVMAT file.
type Material struct {
	Ambient        []float64 `json:"ambient,omitempty" yaml:"ambient,omitempty"`                   // Ambient color
	Diffuse        []float64 `json:"diffuse,omitempty" yaml:"diffuse,omitempty"`                   // Diffuse color
	ForcedDiffuse  []float64 `json:"forced_diffuse,omitempty" yaml:"forced_diffuse,omitempty"`     // Forced diffuse color
	Emmisive       []float64 `json:"emmisive,omitempty" yaml:"emmisive,omitempty"`                 // Emmisive color
	Specular       []float64 `json:"specular,omitempty" yaml:"specular,omitempty"`                 // Specular color
	SpecularPower  *float64  `json:"specular_power,omitempty" yaml:"specular_power,omitempty"`     // Specular power
	PixelShaderID  string    `json:"pixel_shader_id,omitempty" yaml:"pixel_shader_id,omitempty"`   // Pixel shader ID
	VertexShaderID string    `json:"vertex_shader_id,omitempty" yaml:"vertex_shader_id,omitempty"` // Vertex shader ID
	Stages         []Stage   `json:"stages,omitempty" yaml:"stages,omitempty"`                     // Shading stages
	TexGens        []TexGen  `json:"tex_gen,omitempty" yaml:"tex_gen,omitempty"`                   // Texture generators
	extras         []node    // Extra nodes
}

// Stage represents a StageX class.
type Stage struct {
	Name        string       `json:"name,omitempty" yaml:"name,omitempty"`                 // Name of the stage
	Texture     TextureRef   `json:"texture" yaml:"texture"`                               // Texture reference
	UVSource    string       `json:"uv_source,omitempty" yaml:"uv_source,omitempty"`       // UV source
	TexGen      string       `json:"tex_gen,omitempty" yaml:"tex_gen,omitempty"`           // Texture generator
	UVTransform *UVTransform `json:"uv_transform,omitempty" yaml:"uv_transform,omitempty"` // UV transform
	extras      []node       // Extra nodes
}

// TexGen represents a TexGenX class.
type TexGen struct {
	Name        string       `json:"name,omitempty" yaml:"name,omitempty"`                 // Name of the texture generator
	Base        string       `json:"base,omitempty" yaml:"base,omitempty"`                 // Base of the texture
	UVSource    string       `json:"uv_source,omitempty" yaml:"uv_source,omitempty"`       // UV source
	UVTransform *UVTransform `json:"uv_transform,omitempty" yaml:"uv_transform,omitempty"` // UV transform
	extras      []node       // Extra nodes
}

// UVTransform represents uvTransform or TexGen transform.
type UVTransform struct {
	Aside []float64 `json:"aside,omitempty" yaml:"aside,omitempty"` // Aside vector
	Up    []float64 `json:"up,omitempty" yaml:"up,omitempty"`       // Up vector
	Dir   []float64 `json:"dir,omitempty" yaml:"dir,omitempty"`     // Direction vector
	Pos   []float64 `json:"pos,omitempty" yaml:"pos,omitempty"`     // Position vector
}
