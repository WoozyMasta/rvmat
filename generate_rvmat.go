// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// autoFillRoleSuffixes defines suffix priority for role discovery.
	autoFillRoleSuffixes = map[string][]string{
		"nohq": {"nohq", "nshq", "novhq", "nopx", "nofhq", "nof", "no", "normalmap", "non", "ns", "noex", "nsex"},
		"dt":   {"dt", "detail", "cdt"},
		"mc":   {"mc"},
		"as":   {"as", "ads", "adshq"},
		"smdi": {"smdi", "dtsmdi", "sm"},
	}
)

// GenerateSet orchestrates high-level rvmat generation workflow.
func GenerateSet(opts GenerateSetOptions) (*GenerateSetResult, error) {
	mainOutputPath := normalizeOutputPath(opts.OutputPath)
	baseTexture := strings.TrimSpace(opts.BaseTexture)
	var discoveredBaseDirIndex map[string]string
	if baseTexture == "" {
		baseTexture, discoveredBaseDirIndex = discoverBaseTextureFromOutput(mainOutputPath, opts.BaseMaterial)
	}

	// normalize stage texture overrides
	overrides, explicitSource := normalizeOverrideTextures(opts.TextureOverrides)
	mode, err := normalizeTextureAutoFillMode(
		opts.TextureAutoFillMode,
		opts.ForceProceduralOnly,
		baseTexture,
		overrides,
	)
	if err != nil {
		return nil, fmt.Errorf("generate rvmat: %w", err)
	}

	// resolve auto-fill seed
	autoFilled := map[string]string{}
	if mode != TextureAutoFillModeDisabled {
		seed := resolveAutoFillSeed(mode, baseTexture, overrides)
		var autoFillDirIndex map[string]string
		if mode == TextureAutoFillModeFromBaseTexture && discoveredBaseDirIndex != nil {
			autoFillDirIndex = discoveredBaseDirIndex
		}

		for role, raw := range resolveAutoFillBySeed(seed, autoFillDirIndex) {
			if hasExplicitRoleOrStage(overrides, role) {
				continue
			}

			overrides[role] = raw
			autoFilled[role] = raw
		}
	}

	baseTextureForDerive := ""
	if mode == TextureAutoFillModeDisabled && !opts.ForceProceduralOnly {
		baseTextureForDerive = NormalizeGameTexturePath(baseTexture)
	}

	useTexGen := !opts.DisableTexGen

	normalizedOverrides := normalizeTexturePathMap(overrides)

	// generate main material
	main, err := Generate(GenerateOptions{
		TextureOverrides:  normalizedOverrides,
		BaseTexture:       baseTextureForDerive,
		EmissiveIntensity: opts.EmissiveIntensity,
		BaseMaterial:      opts.BaseMaterial,
		Condition:         opts.Condition,
		Finish:            opts.Finish,
		UseTexGen:         useTexGen,
	})
	if err != nil {
		return nil, fmt.Errorf("generate rvmat: %w", err)
	}

	applyTexturePrefixToMaterial(main, opts.TexturePrefix)

	// build result
	result := &GenerateSetResult{
		Main:               main,
		MainOutputPath:     mainOutputPath,
		StageResolutions:   resolveStageResolution(main, explicitSource, autoFilled, baseTextureForDerive),
		DamageOutputPath:   "",
		DestructOutputPath: "",
	}

	generateDamage := opts.GenerateDamage || !opts.DisableDamage
	generateDestruct := opts.GenerateDestruct || !opts.DisableDestruct

	// generate damage variant
	if generateDamage {
		damageMacro := resolveMacroTexture(opts.DamageMacroTexture, DefaultDamageMacroTexture)
		damage, derr := generateVariantWithMacro(main, damageMacro)
		if derr != nil {
			return nil, fmt.Errorf("generate rvmat damage: %w", derr)
		}

		result.Damage = damage
		result.DamageOutputPath = outputPathWithSuffix(result.MainOutputPath, "_damage")
	}

	// generate destruct variant
	if generateDestruct {
		destructMacro := resolveMacroTexture(opts.DestructMacroTexture, DefaultDestructMacroTexture)
		destruct, derr := generateVariantWithMacro(main, destructMacro)
		if derr != nil {
			return nil, fmt.Errorf("generate rvmat destruct: %w", derr)
		}

		result.Destruct = destruct
		result.DestructOutputPath = outputPathWithSuffix(result.MainOutputPath, "_destruct")
	}

	return result, nil
}

// normalizeTextureAutoFillMode validates mode and applies force-procedural override.
func normalizeTextureAutoFillMode(mode TextureAutoFillMode, forceProcedural bool, baseTexture string, overrides map[string]string) (TextureAutoFillMode, error) {
	if forceProcedural {
		return TextureAutoFillModeDisabled, nil
	}

	switch mode {
	case TextureAutoFillModeDefault:
		if strings.TrimSpace(baseTexture) != "" {
			return TextureAutoFillModeFromBaseTexture, nil
		}
		if strings.TrimSpace(overrides["stage1"]) != "" {
			return TextureAutoFillModeFromStageOverride, nil
		}
		return TextureAutoFillModeFromBaseTexture, nil

	case TextureAutoFillModeDisabled,
		TextureAutoFillModeFromBaseTexture,
		TextureAutoFillModeFromStageOverride:
		return mode, nil

	default:
		return 0, fmt.Errorf("%w auto_fill_mode=%s", ErrInvalidGenerateOption, mode)
	}
}

// normalizeOverrideTextures normalizes override keys and captures explicit source map.
func normalizeOverrideTextures(in map[string]string) (map[string]string, map[string]StageTextureSource) {
	out := normalizeOverrideKeysToStage(in)
	source := map[string]StageTextureSource{}
	if len(out) == 0 {
		return out, source
	}

	for key := range out {
		source[key] = StageTextureSourceExplicit
	}

	return out, source
}

// resolveAutoFillSeed picks source texture path used for stem-based disk lookup.
func resolveAutoFillSeed(mode TextureAutoFillMode, baseTexture string, overrides map[string]string) string {
	switch mode {
	case TextureAutoFillModeFromBaseTexture:
		return strings.TrimSpace(baseTexture)

	case TextureAutoFillModeFromStageOverride:
		if raw := strings.TrimSpace(overrides["stage1"]); raw != "" {
			return raw
		}

		for _, key := range []string{"stage2", "stage3", "stage4", "stage5"} {
			if raw := strings.TrimSpace(overrides[key]); raw != "" {
				return raw
			}
		}
	}

	return ""
}

// resolveMacroTexture returns override macro texture or fallback default.
func resolveMacroTexture(override, fallback string) string {
	if raw := strings.TrimSpace(override); raw != "" {
		return NormalizeGameTexturePath(raw)
	}

	return NormalizeGameTexturePath(fallback)
}

// normalizeOutputPath adds ".rvmat" extension when output path has no extension.
func normalizeOutputPath(path string) string {
	out := strings.TrimSpace(path)
	if out == "" {
		return ""
	}

	if filepath.Ext(out) != "" {
		return out
	}

	return out + ".rvmat"
}

// discoverBaseTextureFromOutput searches sibling base texture from output stem.
func discoverBaseTextureFromOutput(outputPath string, baseMaterial BaseMaterial) (string, map[string]string) {
	if strings.TrimSpace(outputPath) == "" {
		return "", nil
	}

	ext := filepath.Ext(outputPath)
	stem := strings.TrimSuffix(outputPath, ext)
	if stem == "" {
		return "", nil
	}

	baseStem := stem
	if normalizedStem, _, ok := splitBaseTextureStem(stem); ok && strings.TrimSpace(normalizedStem) != "" {
		baseStem = normalizedStem
	}

	dirLowerNameIndex := buildDirLowerNameIndex(filepath.Dir(baseStem))
	for _, suffix := range colorSuffixPriorityForMaterial(baseMaterial) {
		if suffix == "" {
			continue
		}

		if found := pickExistingTextureByPriorityFromIndex(baseStem+suffix, dirLowerNameIndex); found != "" {
			return found, dirLowerNameIndex
		}

		return baseStem + suffix + ".paa", dirLowerNameIndex
	}

	return "", dirLowerNameIndex
}

// resolveAutoFillBySeed resolves existing sibling textures by suffix priorities.
func resolveAutoFillBySeed(seedRaw string, dirLowerNameIndex map[string]string) map[string]string {
	out := map[string]string{}
	stem, _, ok := splitBaseTextureStem(seedRaw)
	if !ok {
		return out
	}

	dir := filepath.Dir(stem)
	base := filepath.Base(stem)
	if strings.TrimSpace(base) == "" || base == "." {
		return out
	}

	if len(dirLowerNameIndex) == 0 {
		dirLowerNameIndex = buildDirLowerNameIndex(dir)
	}

	extensions := textureExtensionsByPriority()
	for role, suffixes := range autoFillRoleSuffixes {
		for _, suffix := range suffixes {
			for _, fileExt := range extensions {
				candidateBaseName := strings.ToLower(base + "_" + suffix + fileExt)
				if matchedName, ok := dirLowerNameIndex[candidateBaseName]; ok {
					out[role] = filepath.Join(dir, matchedName)
					break
				}
			}

			if strings.TrimSpace(out[role]) != "" {
				break
			}
		}
	}

	return out
}

// pickExistingTextureByPriorityFromIndex resolves existing texture by extension
// priority from pre-indexed lower-case directory names.
func pickExistingTextureByPriorityFromIndex(baseWithoutExt string, dirLowerNameIndex map[string]string) string {
	if len(dirLowerNameIndex) == 0 {
		return ""
	}

	dir := filepath.Dir(baseWithoutExt)
	base := filepath.Base(baseWithoutExt)
	if strings.TrimSpace(base) == "" || base == "." {
		return ""
	}

	for _, fileExt := range textureExtensionsByPriority() {
		candidate := strings.ToLower(base + fileExt)
		if matchedName, ok := dirLowerNameIndex[candidate]; ok {
			return filepath.Join(dir, matchedName)
		}
	}

	return ""
}

// buildDirLowerNameIndex returns lower-case filename to original name map.
func buildDirLowerNameIndex(dir string) map[string]string {
	cleanDir := filepath.Clean(strings.TrimSpace(dir))
	if cleanDir == "" {
		cleanDir = "."
	}

	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		return nil
	}

	index := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}

		index[strings.ToLower(name)] = name
	}

	return index
}

// hasExplicitRoleOrStage reports whether role or its stage was explicitly overridden.
func hasExplicitRoleOrStage(overrides map[string]string, role string) bool {
	stageName, ok := stageNameForTextureRole(role)
	if !ok {
		return false
	}

	return strings.TrimSpace(overrides[stageName]) != ""
}

// resolveStageResolution builds stage texture source report for generated material.
func resolveStageResolution(m *Material, explicit map[string]StageTextureSource, autoFilled map[string]string, baseTexture string) map[string]GenerateStageResolution {
	out := map[string]GenerateStageResolution{}
	if m == nil {
		return out
	}

	for _, stage := range m.Stages {
		stageKey := strings.ToLower(strings.TrimSpace(stage.Name))
		role, _ := textureRoleForStageName(stageKey)
		var source StageTextureSource

		switch {
		case explicit[stageKey] == StageTextureSourceExplicit:
			source = StageTextureSourceExplicit

		case role != "" && strings.TrimSpace(autoFilled[role]) != "":
			source = StageTextureSourceAutoFill

		case role != "" && role != "env" && strings.TrimSpace(baseTexture) != "" && stage.Texture.IsPath():
			source = StageTextureSourceDerived

		case stage.Texture.IsProcedural():
			source = StageTextureSourceProcedural

		default:
			source = StageTextureSourceProcedural
		}

		out[stage.Name] = GenerateStageResolution{
			Role:    role,
			Source:  source,
			Texture: stage.Texture,
		}
	}

	return out
}

// outputPathWithSuffix inserts suffix before file extension.
func outputPathWithSuffix(path, suffix string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}

	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	if ext == "" {
		ext = ".rvmat"
	}

	return base + suffix + ext
}

// normalizeTexturePathMap normalizes all texture path values in map.
func normalizeTexturePathMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return in
	}

	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = NormalizeGameTexturePath(value)
	}

	return out
}

// applyTexturePrefixToMaterial prepends prefix to path textures in material.
func applyTexturePrefixToMaterial(mat *Material, prefix string) {
	if mat == nil {
		return
	}

	for i := range mat.Stages {
		stage := &mat.Stages[i]
		if !stage.Texture.IsPath() {
			continue
		}

		stage.Texture.Raw = withTexturePrefix(stage.Texture.Raw, prefix)
	}
}

// withTexturePrefix returns normalized texture path with prefix applied.
func withTexturePrefix(rawPath, prefix string) string {
	normalizedPath := NormalizeGameTexturePath(rawPath)
	if strings.TrimSpace(normalizedPath) == "" {
		return normalizedPath
	}
	if hasKnownGameRootPrefix(normalizedPath) {
		return normalizedPath
	}

	normalizedPrefix := strings.TrimSpace(NormalizeGameTexturePath(prefix))
	if normalizedPrefix == "" {
		return normalizedPath
	}
	if normalizedPath == normalizedPrefix || strings.HasPrefix(normalizedPath, normalizedPrefix+`\`) {
		return normalizedPath
	}

	return NormalizeGameTexturePath(normalizedPrefix + `\` + normalizedPath)
}

// hasKnownGameRootPrefix reports whether path already uses a known game root.
func hasKnownGameRootPrefix(path string) bool {
	for _, root := range []string{`dz\`, `ca\`, `a3\`} {
		if strings.HasPrefix(path, root) {
			return true
		}
	}

	return false
}

// colorSuffixPriorityForMaterial returns base color suffix preference by material.
func colorSuffixPriorityForMaterial(baseMaterial BaseMaterial) []string {
	switch baseMaterial {
	case BaseMaterialGlass:
		return []string{"_ca", "_co"}
	default:
		return []string{"_co", "_ca"}
	}
}
