// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"errors"
	"os"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

// Validate validates a material and returns issues.
func Validate(m *Material, opt *ValidateOptions) []lint.Diagnostic {
	vopt := opt.normalize()
	var out []lint.Diagnostic

	if len(m.Stages) > 0 {
		if m.PixelShaderID == "" {
			out = append(out, warningDiagnostic(
				CodeValidatePixelShaderMissing,
				"PixelShaderID missing",
				"",
			))
		}
		if m.VertexShaderID == "" {
			out = append(out, warningDiagnostic(
				CodeValidateVertexShaderMissing,
				"VertexShaderID missing",
				"",
			))
		}
	}

	if !vopt.DisableShaderNameCheck {
		if m.PixelShaderID != "" {
			if !isKnownNameCI(knownPixelShaderID, m.PixelShaderID) {
				out = append(out, warningDiagnostic(
					CodeValidateUnknownPixelShaderID,
					"unknown PixelShaderID",
					m.PixelShaderID,
				))
			}
		}
		if m.VertexShaderID != "" {
			if !isKnownNameCI(knownVertexShaderID, m.VertexShaderID) {
				out = append(out, warningDiagnostic(
					CodeValidateUnknownVertexShaderID,
					"unknown VertexShaderID",
					m.VertexShaderID,
				))
			}
		}
	}

	if vopt.EnableShaderProfileCheck {
		out = append(out, validateShaderProfiles(m)...)
	}

	for _, texGen := range m.TexGens {
		if texGen.UVTransform == nil {
			continue
		}

		out = append(out, validateUVTransform(texGen.Name, texGen.UVTransform)...)
	}

	out = append(out, validateColor("ambient", m.Ambient)...)
	out = append(out, validateColor("diffuse", m.Diffuse)...)
	out = append(out, validateColor("forcedDiffuse", m.ForcedDiffuse)...)
	out = append(out, validateColor("emissive", m.Emissive)...)
	out = append(out, validateColor("specular", m.Specular)...)

	// Check if path-mode validation or extension validation is enabled.
	if vopt.TexturePathMode != TexturePathModeIgnore || !vopt.DisableExtensionsCheck {
		resolver := PathResolver{GameRoot: vopt.GameRoot}
		for _, st := range m.Stages {
			tex := st.Texture
			if tex.Raw == "" || tex.IsProcedural() {
				continue
			}

			if !vopt.DisableExtensionsCheck {
				if !hasAllowedExt(tex.Raw) {
					out = append(out, warningDiagnostic(
						CodeValidateUnexpectedTextureExtension,
						"unexpected texture extension",
						tex.Raw,
					))
				}
			}

			if strings.Contains(tex.Raw, "..") {
				out = append(out, warningDiagnostic(
					CodeValidateTexturePathParentTraversal,
					"texture path contains '..'",
					tex.Raw,
				))
			}

			if vopt.TexturePathMode == TexturePathModeIgnore {
				continue
			}

			if shouldExcludePath(tex.Raw, vopt.ExcludePaths) {
				continue
			}

			if vopt.TexturePathMode == TexturePathModeTrust &&
				hasTrustedGameRootPrefix(tex.Raw, vopt.TrustedPrefixes) {
				continue
			}

			p := resolver.ResolvePath(tex.Raw)
			if p != "" {
				if _, err := os.Stat(p); err != nil {
					out = append(out, warningDiagnostic(
						CodeValidateTextureFileNotFound,
						"texture file not found",
						p,
					))
				}
			}
		}
	}

	for _, st := range m.Stages {
		if !vopt.DisableShaderNameCheck {
			if !isKnownNameCI(knownStageNames, st.Name) {
				out = append(out, warningDiagnostic(
					CodeValidateUnknownStageName,
					"unknown Stage name",
					st.Name,
				))
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
					out = append(out, warningDiagnostic(
						CodeValidateStageUnknownTexGen,
						"stage references unknown texGen",
						st.Name,
					))

				case errors.Is(err, ErrTexGenBaseNotFound):
					out = append(out, warningDiagnostic(
						CodeValidateTexGenBaseNotFound,
						"texGen inheritance base not found",
						st.Name,
					))

				case errors.Is(err, ErrTexGenCycle):
					out = append(out, warningDiagnostic(
						CodeValidateTexGenCycle,
						"texGen inheritance cycle detected",
						st.Name,
					))

				default:
					out = append(out, warningDiagnostic(
						CodeValidateTexGenResolutionFailed,
						"texGen resolution failed",
						st.Name,
					))
				}

				continue
			}

			if resolved != nil {
				uvSource = resolved.UVSource
				uvTransform = resolved.UVTransform
			}
		}

		// No UVs expected.
		if uvTransform != nil {
			out = append(out, validateUVTransform(st.Name, uvTransform)...)
		}

		if uvSource == "none" || uvSource == "WorldPos" {
			continue
		}

		// Check if effective uvSource/uvTransform are missing.
		if uvSource == "" && uvTransform == nil {
			out = append(out, warningDiagnostic(
				CodeValidateStageMissingEffectiveUVSource,
				"stage missing effective uvSource",
				st.Name,
			))
			out = append(out, warningDiagnostic(
				CodeValidateStageMissingEffectiveUVTransform,
				"stage missing effective uvTransform",
				st.Name,
			))
			continue
		}

		if uvTransform == nil {
			out = append(out, warningDiagnostic(
				CodeValidateStageMissingEffectiveUVTransform,
				"stage missing effective uvTransform",
				st.Name,
			))
		}
	}

	seen := make(map[string]struct{}, len(m.Stages))
	for _, st := range m.Stages {
		if st.Name == "" {
			continue
		}
		if _, ok := seen[st.Name]; ok {
			out = append(out, errorDiagnostic(
				CodeValidateDuplicateStageName,
				"duplicate Stage name",
				st.Name,
			))
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
func validateShaderProfiles(m *Material) []lint.Diagnostic {
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

	out := make([]lint.Diagnostic, 0, len(profile.Required)+len(profile.Recommended))
	hasMissingProfileStages := false
	for _, stage := range profile.Required {
		if _, ok := seen[strings.ToLower(stage)]; ok {
			continue
		}
		hasMissingProfileStages = true
		out = append(out, warningDiagnostic(
			CodeValidateShaderProfileMissingRequiredStage,
			"shader profile missing required stage",
			stage,
		))
	}

	for _, stage := range profile.Recommended {
		if _, ok := seen[strings.ToLower(stage)]; ok {
			continue
		}
		hasMissingProfileStages = true
		out = append(out, warningDiagnostic(
			CodeValidateShaderProfileMissingCommonStage,
			"shader profile missing common stage",
			stage,
		))
	}

	if !hasMissingProfileStages {
		out = append(out, validateStrictShaderStageSet(ps, seen)...)
	}

	return out
}

// validateStrictShaderStageSet checks strict stage sets for known shader profiles.
func validateStrictShaderStageSet(profile string, seen map[string]struct{}) []lint.Diagnostic {
	switch profile {
	case "super":
		return validateExpectedStageSet(
			seen,
			[]string{
				"stage1", "stage2", "stage3", "stage4", "stage5", "stage6", "stage7",
			},
			[]string{"stage0"},
			"shader profile stage set mismatch (expected Stage1..Stage7, Stage0 optional)",
		)
	case "multi":
		return validateExpectedStageSet(
			seen,
			[]string{
				"stage0", "stage1", "stage2", "stage3", "stage4", "stage5", "stage6", "stage7",
				"stage8", "stage9", "stage10", "stage11", "stage12", "stage13", "stage14",
			},
			nil,
			"shader profile stage set mismatch (expected Stage0..Stage14)",
		)
	default:
		return nil
	}
}

// validateExpectedStageSet validates required/optional stage sets.
func validateExpectedStageSet(seen map[string]struct{}, required, optional []string, message string) []lint.Diagnostic {
	requiredSet := make(map[string]struct{}, len(required))
	for _, stage := range required {
		requiredSet[strings.ToLower(strings.TrimSpace(stage))] = struct{}{}
	}

	optionalSet := make(map[string]struct{}, len(optional))
	for _, stage := range optional {
		optionalSet[strings.ToLower(strings.TrimSpace(stage))] = struct{}{}
	}

	for stage := range requiredSet {
		if _, ok := seen[stage]; !ok {
			return []lint.Diagnostic{warningDiagnostic(
				CodeValidateShaderProfileStageSetMismatch,
				message,
				"",
			)}
		}
	}

	for stage := range seen {
		if _, ok := requiredSet[stage]; ok {
			continue
		}
		if _, ok := optionalSet[stage]; ok {
			continue
		}
		if strings.HasPrefix(stage, "stage") {
			return []lint.Diagnostic{warningDiagnostic(
				CodeValidateShaderProfileStageSetMismatch,
				message,
				"",
			)}
		}
	}

	return nil
}

// ValidateWithTextureOptions validates a material and its textures.
func ValidateWithTextureOptions(m *Material, opt *ValidateOptions, texOpt *TextureValidateOptions) []lint.Diagnostic {
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
func validateColor(name string, vals []float64) []lint.Diagnostic {
	// Colors should be 4-component RGBA.
	if len(vals) == 0 {
		return nil
	}
	if len(vals) != 4 {
		return []lint.Diagnostic{errorDiagnostic(
			CodeValidateColorComponentCount,
			"color must have 4 components",
			name,
		)}
	}
	return nil
}

// validateUVTransform validates uvTransform vectors layout.
func validateUVTransform(path string, transform *UVTransform) []lint.Diagnostic {
	if transform == nil {
		return nil
	}

	out := make([]lint.Diagnostic, 0, 4)
	out = append(out, validateUVTransformVector(path, "aside", transform.Aside)...)
	out = append(out, validateUVTransformVector(path, "up", transform.Up)...)
	out = append(out, validateUVTransformVector(path, "dir", transform.Dir)...)
	out = append(out, validateUVTransformVector(path, "pos", transform.Pos)...)

	return out
}

// validateUVTransformVector validates one uvTransform vector component.
func validateUVTransformVector(path, field string, values []float64) []lint.Diagnostic {
	vectorPath := path
	if strings.TrimSpace(field) != "" {
		vectorPath = path + ".uvTransform." + field
	}

	if len(values) == 0 {
		return []lint.Diagnostic{errorDiagnostic(
			CodeValidateUVTransformVectorRequired,
			"uvTransform vector is required",
			vectorPath,
		)}
	}

	if len(values) != 3 {
		return []lint.Diagnostic{errorDiagnostic(
			CodeValidateUVTransformVectorComponentCount,
			"uvTransform vector must have 3 components",
			vectorPath,
		)}
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

// hasTrustedGameRootPrefix reports whether path starts with trusted root prefix.
func hasTrustedGameRootPrefix(path string, trustedPrefixes []string) bool {
	normalizedPath := NormalizeGameTexturePath(path)
	if normalizedPath == "" {
		return false
	}

	for index := range trustedPrefixes {
		prefix := strings.Trim(NormalizeGameTexturePath(trustedPrefixes[index]), `\`)
		if prefix == "" {
			continue
		}

		if normalizedPath == prefix ||
			strings.HasPrefix(normalizedPath, prefix+`\`) {
			return true
		}
	}

	return false
}

// buildAllowedTextureExtensionSet builds normalized extension allow-list set.
func buildAllowedTextureExtensionSet(extensions []string) map[string]struct{} {
	out := make(map[string]struct{}, len(extensions))

	for index := range extensions {
		extension := strings.ToLower(strings.TrimSpace(extensions[index]))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}

		out[extension] = struct{}{}
	}

	return out
}

// validateTexture validates a texture.
func validateTexture(t TextureRef, opt TextureValidateOptions) []lint.Diagnostic {
	if t.Raw == "" {
		return nil
	}
	if t.IsPath() {
		return nil
	}

	if !t.ParsedOK || t.Procedural == nil {
		if !opt.DisableProceduralFnCheck || !opt.DisableProceduralArgsCheck || !opt.DisableTextureTagCheck {
			return []lint.Diagnostic{warningDiagnostic(
				CodeValidateProceduralTextureParseFailed,
				"procedural texture parse failed",
				t.Raw,
			)}
		}
		return nil
	}

	pt := t.Procedural
	fn := strings.ToLower(pt.Func)

	var out []lint.Diagnostic
	if !opt.DisableProceduralFnCheck {
		if _, ok := knownProceduralFns[fn]; !ok {
			out = append(out, warningDiagnostic(
				CodeValidateUnknownProceduralFunction,
				"unknown procedural function",
				pt.Func,
			))
		}
	}

	if !opt.DisableProceduralFnCheck {
		if _, ok := knownProceduralFormats[strings.ToLower(strings.TrimSpace(pt.Format))]; !ok {
			out = append(out, warningDiagnostic(
				CodeValidateUnknownProceduralTextureFormat,
				"unknown procedural texture format",
				pt.Format,
			))
		}
	}

	if pt.Width <= 0 || pt.Height <= 0 || pt.Mip < 0 {
		out = append(out, warningDiagnostic(
			CodeValidateInvalidProceduralTextureHeaderDimensions,
			"invalid procedural texture header dimensions",
			t.Raw,
		))
	}

	if !opt.DisableProceduralArgsCheck {
		if !proceduralArgsOK(fn, pt.Args) {
			out = append(out, warningDiagnostic(
				CodeValidateUnexpectedProceduralArgumentCount,
				"unexpected procedural argument count",
				pt.Func,
			))
		}
		if !proceduralNumericArgsOK(fn, pt, pt.Args) {
			out = append(out, warningDiagnostic(
				CodeValidateInvalidProceduralNumericArguments,
				"invalid procedural numeric arguments",
				pt.Func,
			))
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
			if _, ok := knownTextureTags[strings.ToLower(strings.TrimSpace(tag))]; !ok {
				out = append(out, warningDiagnostic(
					CodeValidateUnknownTextureTag,
					"unknown texture tag",
					tag,
				))
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

// proceduralNumericArgsOK checks parsed numeric arguments for known functions.
func proceduralNumericArgsOK(fn string, pt *ProceduralTexture, args []string) bool {
	if pt == nil {
		return false
	}

	switch fn {
	case "color":
		return pt.Color != nil
	case "fresnel":
		return pt.Fresnel != nil
	case "fresnelglass":
		if len(args) == 0 {
			return true
		}
		return pt.Fresnel != nil
	case "irradiance":
		return pt.Irradiance != nil
	default:
		return true
	}
}

// withStageContext adds stage context to one diagnostic path.
func withStageContext(diagnostic lint.Diagnostic, stage string) lint.Diagnostic {
	if stage == "" {
		return diagnostic
	}

	if diagnostic.Path == "" {
		diagnostic.Path = stage
		return diagnostic
	}

	diagnostic.Path = stage + ": " + diagnostic.Path
	return diagnostic
}
