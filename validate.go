// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"errors"
	"os"
	"strings"
)

// IssueLevel represents severity of validation issue.
type IssueLevel string

const (
	// IssueError indicates a validation error.
	IssueError IssueLevel = "error"
	// IssueWarning indicates a validation warning.
	IssueWarning IssueLevel = "warning"
)

// Issue represents a validation issue.
type Issue struct {
	Level   IssueLevel `json:"level" yaml:"level"`                   // Severity level
	Code    string     `json:"code,omitempty" yaml:"code,omitempty"` // Machine-readable code
	Message string     `json:"message" yaml:"message"`               // Issue message
	Path    string     `json:"path,omitempty" yaml:"path,omitempty"` // Path to the affected resource
}

// Validate validates a material and returns issues.
func Validate(m *Material, opt *ValidateOptions) []Issue {
	vopt := opt.normalize()
	var out []Issue

	if len(m.Stages) > 0 {
		if m.PixelShaderID == "" {
			out = append(out, Issue{Level: IssueWarning, Message: "PixelShaderID missing"})
		}
		if m.VertexShaderID == "" {
			out = append(out, Issue{Level: IssueWarning, Message: "VertexShaderID missing"})
		}
	}

	if !vopt.DisableShaderNameCheck {
		if m.PixelShaderID != "" {
			if !isKnownNameCI(knownPixelShaderID, m.PixelShaderID) {
				out = append(out, Issue{Level: IssueWarning, Message: "unknown PixelShaderID", Path: m.PixelShaderID})
			}
		}
		if m.VertexShaderID != "" {
			if !isKnownNameCI(knownVertexShaderID, m.VertexShaderID) {
				out = append(out, Issue{Level: IssueWarning, Message: "unknown VertexShaderID", Path: m.VertexShaderID})
			}
		}
	}

	if vopt.EnableShaderProfileCheck {
		out = append(out, validateShaderProfiles(m)...)
	}

	out = append(out, validateColor("ambient", m.Ambient)...)
	out = append(out, validateColor("diffuse", m.Diffuse)...)
	out = append(out, validateColor("forcedDiffuse", m.ForcedDiffuse)...)
	out = append(out, validateColor("emissive", m.Emissive)...)
	out = append(out, validateColor("specular", m.Specular)...)

	// Check if file validation or extension validation is enabled
	if !vopt.DisableFileCheck || !vopt.DisableExtensionsCheck {
		resolver := PathResolver{GameRoot: vopt.GameRoot}
		for _, st := range m.Stages {
			tex := st.Texture
			if tex.Raw == "" || tex.IsProcedural() {
				continue
			}

			if !vopt.DisableExtensionsCheck {
				if !hasAllowedExt(tex.Raw) {
					out = append(out, Issue{Level: IssueWarning, Message: "unexpected texture extension", Path: tex.Raw})
				}
			}

			if strings.Contains(tex.Raw, "..") {
				out = append(out, Issue{Level: IssueWarning, Message: "texture path contains '..'", Path: tex.Raw})
			}

			if !vopt.DisableFileCheck {
				if shouldExcludePath(tex.Raw, vopt.ExcludePaths) {
					continue
				}
				p := resolver.ResolvePath(tex.Raw)
				if p != "" {
					if _, err := os.Stat(p); err != nil {
						out = append(out, Issue{Level: IssueWarning, Code: "missing_resource", Message: "texture file not found", Path: p})
					}
				}
			}
		}
	}

	for _, st := range m.Stages {
		if !vopt.DisableShaderNameCheck {
			if !isKnownNameCI(knownStageNames, st.Name) {
				out = append(out, Issue{Level: IssueWarning, Message: "unknown Stage name", Path: st.Name})
			}
		}

		// Known case in game data where uvSource/uvTransform may be omitted.
		if st.Name == "StageTI" || st.Name == "Stage0" {
			continue
		}

		uvSource := st.UVSource
		uvTransform := st.UVTransform
		if st.TexGen != "" {
			resolved, err := ResolveStageTexGen(m, st)
			if err != nil {
				switch {
				case errors.Is(err, ErrTexGenNotFound):
					out = append(out, Issue{
						Level:   IssueWarning,
						Message: "stage references unknown texGen",
						Path:    st.Name,
					})

				case errors.Is(err, ErrTexGenBaseNotFound):
					out = append(out, Issue{
						Level:   IssueWarning,
						Message: "texGen inheritance base not found",
						Path:    st.Name,
					})

				case errors.Is(err, ErrTexGenCycle):
					out = append(out, Issue{
						Level:   IssueWarning,
						Message: "texGen inheritance cycle detected",
						Path:    st.Name,
					})

				default:
					out = append(out, Issue{
						Level:   IssueWarning,
						Message: "texGen resolution failed",
						Path:    st.Name,
					})
				}

				continue
			}

			if resolved != nil {
				uvSource = resolved.UVSource
				uvTransform = resolved.UVTransform
			}
		}

		// No UVs expected.
		if uvSource == "none" || uvSource == "WorldPos" {
			continue
		}

		// Check if effective uvSource/uvTransform are missing.
		if uvSource == "" && uvTransform == nil {
			out = append(out, Issue{Level: IssueWarning, Message: "stage missing effective uvSource", Path: st.Name})
			out = append(out, Issue{Level: IssueWarning, Message: "stage missing effective uvTransform", Path: st.Name})
			continue
		}

		if uvTransform == nil {
			out = append(out, Issue{Level: IssueWarning, Message: "stage missing effective uvTransform", Path: st.Name})
		}
	}

	seen := make(map[string]struct{}, len(m.Stages))
	for _, st := range m.Stages {
		if st.Name == "" {
			continue
		}
		if _, ok := seen[st.Name]; ok {
			out = append(out, Issue{Level: IssueError, Message: "duplicate Stage name", Path: st.Name})
			continue
		}
		seen[st.Name] = struct{}{}
	}

	return out
}

// isKnownNameCI checks known-name maps in case-insensitive mode.
func isKnownNameCI(known map[string]struct{}, value string) bool {
	if value == "" {
		return false
	}

	if _, ok := known[value]; ok {
		return true
	}

	for k := range known {
		if strings.EqualFold(k, value) {
			return true
		}
	}

	return false
}

// validateShaderProfiles performs soft expected-stage checks for known shaders.
func validateShaderProfiles(m *Material) []Issue {
	ps := strings.ToLower(strings.TrimSpace(m.PixelShaderID))
	if ps == "" {
		return nil
	}

	profile, ok := shaderProfileHints[ps]
	if !ok {
		return nil
	}

	seen := make(map[string]struct{}, len(m.Stages))
	for _, st := range m.Stages {
		seen[strings.ToLower(strings.TrimSpace(st.Name))] = struct{}{}
	}

	out := make([]Issue, 0, len(profile.Required)+len(profile.Recommended))
	for _, stage := range profile.Required {
		if _, ok := seen[strings.ToLower(stage)]; ok {
			continue
		}
		out = append(out, Issue{
			Level:   IssueWarning,
			Message: "shader profile missing required stage",
			Path:    stage,
		})
	}

	for _, stage := range profile.Recommended {
		if _, ok := seen[strings.ToLower(stage)]; ok {
			continue
		}
		out = append(out, Issue{
			Level:   IssueWarning,
			Message: "shader profile missing common stage",
			Path:    stage,
		})
	}

	return out
}

// ValidateWithTextureOptions validates a material and its textures.
func ValidateWithTextureOptions(m *Material, opt *ValidateOptions, texOpt *TextureValidateOptions) []Issue {
	out := Validate(m, opt)
	if m == nil {
		return out
	}
	if texOpt == nil {
		return out
	}

	for _, st := range m.Stages {
		issues := st.Texture.Validate(texOpt)
		for i := range issues {
			issues[i] = withStageContext(issues[i], st.Name)
		}
		out = append(out, issues...)
	}

	return out
}

// validateColor validates a color.
func validateColor(name string, vals []float64) []Issue {
	// Colors should be 4-component RGBA.
	if len(vals) == 0 {
		return nil
	}
	if len(vals) != 4 {
		return []Issue{{Level: IssueError, Message: "color must have 4 components", Path: name}}
	}
	return nil
}

// shouldExcludePath checks if the path should be excluded.
func shouldExcludePath(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	// Normalize the path for matching
	norm := normalizePathForMatch(path)
	for _, p := range patterns {
		if p == "" {
			continue
		}

		// Check if the path matches a wildcard pattern
		pp := normalizePathForMatch(p)
		if before, ok := strings.CutSuffix(pp, "*"); ok {
			prefix := before
			if strings.HasPrefix(norm, prefix) {
				return true
			}

			continue
		}

		// Check if the path matches an exact pattern
		if norm == pp {
			return true
		}
	}

	return false
}

// normalizePathForMatch normalizes a path for matching.
func normalizePathForMatch(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ReplaceAll(p, "/", "\\")
	return strings.ToLower(p)
}

// validateTexture validates a texture.
func validateTexture(t TextureRef, opt TextureValidateOptions) []Issue {
	if t.Raw == "" {
		return nil
	}
	if t.IsPath() {
		return nil
	}

	if !t.ParsedOK || t.Procedural == nil {
		if !opt.DisableProceduralFnCheck || !opt.DisableProceduralArgsCheck || !opt.DisableTextureTagCheck {
			return []Issue{{Level: IssueWarning, Message: "procedural texture parse failed", Path: t.Raw}}
		}
		return nil
	}

	pt := t.Procedural
	fn := strings.ToLower(pt.Func)

	var out []Issue
	if !opt.DisableProceduralFnCheck {
		if _, ok := knownProceduralFns[fn]; !ok {
			out = append(out, Issue{Level: IssueWarning, Message: "unknown procedural function", Path: pt.Func})
		}
	}

	if !opt.DisableProceduralArgsCheck {
		if !proceduralArgsOK(fn, pt.Args) {
			out = append(out, Issue{Level: IssueWarning, Message: "unexpected procedural argument count", Path: pt.Func})
		}
	}

	if !opt.DisableTextureTagCheck && fn == "color" {
		tag := ""
		if pt.Color != nil {
			tag = pt.Color.Tag
		} else if len(pt.Args) == 5 {
			tag = pt.Args[4]
		}
		if tag != "" {
			if _, ok := knownTextureTags[strings.ToLower(tag)]; !ok {
				out = append(out, Issue{Level: IssueWarning, Message: "unknown texture tag", Path: tag})
			}
		}
	}

	return out
}

// proceduralArgsOK checks if the arguments of a procedural texture are valid.
func proceduralArgsOK(fn string, args []string) bool {
	switch fn {
	case "color":
		return len(args) == 4 || len(args) == 5
	case "fresnel":
		return len(args) == 2
	case "fresnelglass":
		return len(args) == 0 || len(args) == 1 || len(args) == 2
	case "irradiance":
		return len(args) == 1
	default:
		return true
	}
}

// withStageContext adds stage context to an issue.
func withStageContext(issue Issue, stage string) Issue {
	if stage == "" {
		return issue
	}

	if issue.Path == "" {
		issue.Path = stage
		return issue
	}

	issue.Path = stage + ": " + issue.Path
	return issue
}
