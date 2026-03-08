package rvmat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSetDefaultsToProceduralWithDefaultVariants(t *testing.T) {
	result, err := GenerateSet(GenerateSetOptions{
		OutputPath: filepath.Join(t.TempDir(), "default.rvmat"),
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}
	if result.Main == nil {
		t.Fatalf("expected main material")
	}
	if result.Damage == nil || result.Destruct == nil {
		t.Fatalf("expected damage/destruct materials by default")
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil || !stage1.Texture.IsProcedural() {
		t.Fatalf("expected procedural Stage1 texture by default")
	}
	if stage1.TexGen != "0" {
		t.Fatalf("expected compact texgen mode for Stage1, got %q", stage1.TexGen)
	}
	if stage1.UVTransform != nil || stage1.UVSource != "" {
		t.Fatalf("expected no inline UV in compact texgen mode")
	}
	if len(result.Main.TexGens) != 1 {
		t.Fatalf("expected generated TexGen0 only, got %d", len(result.Main.TexGens))
	}

	resolution, ok := result.StageResolutions["Stage1"]
	if !ok {
		t.Fatalf("expected Stage1 resolution")
	}
	if resolution.Source != StageTextureSourceProcedural {
		t.Fatalf("unexpected Stage1 source: %s", resolution.Source)
	}
}

func TestGenerateSetAutoFillFromBaseTexture(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "testbox_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "testbox_nohq.paa"))
	mustTouchFile(t, filepath.Join(tmp, "testbox_as.paa"))
	mustTouchFile(t, filepath.Join(tmp, "testbox_smdi.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		OutputPath:          filepath.Join(tmp, "testbox.rvmat"),
		BaseTexture:         filepath.Join(tmp, "testbox_co.paa"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
		GenerateDamage:      true,
		GenerateDestruct:    true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}
	if result.Main == nil || result.Damage == nil || result.Destruct == nil {
		t.Fatalf("expected main+damage+destruct materials")
	}
	if result.DamageOutputPath == "" || result.DestructOutputPath == "" {
		t.Fatalf("expected output paths for damage/destruct")
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage4 := findMaterialStageByName(result.Main, "Stage4")
	stage5 := findMaterialStageByName(result.Main, "Stage5")
	stage2 := findMaterialStageByName(result.Main, "Stage2")
	if stage1 == nil || stage2 == nil || stage4 == nil || stage5 == nil {
		t.Fatalf("missing expected stages")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "testbox_nohq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage4.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "testbox_as.paa")) {
		t.Fatalf("unexpected Stage4 texture: %q", stage4.Texture.Raw)
	}
	if stage5.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "testbox_smdi.paa")) {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}
	if !stage2.Texture.IsProcedural() {
		t.Fatalf("expected procedural Stage2 fallback when dt is missing")
	}

	if result.StageResolutions["Stage1"].Source != StageTextureSourceAutoFill {
		t.Fatalf("expected autofill source for Stage1")
	}
	if result.StageResolutions["Stage2"].Source != StageTextureSourceProcedural {
		t.Fatalf("expected procedural source for Stage2")
	}
}

func TestGenerateSetAutoFillFromStageOverride(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "crate_nohq.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_as.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_smdi.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		TextureAutoFillMode: TextureAutoFillModeFromStageOverride,
		TextureOverrides: map[string]string{
			"nohq": filepath.Join(tmp, "crate_nohq.paa"),
		},
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage4 := findMaterialStageByName(result.Main, "Stage4")
	stage5 := findMaterialStageByName(result.Main, "Stage5")
	if stage1 == nil || stage4 == nil || stage5 == nil {
		t.Fatalf("missing expected stages")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_nohq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage4.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_as.paa")) {
		t.Fatalf("unexpected Stage4 texture: %q", stage4.Texture.Raw)
	}
	if stage5.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_smdi.paa")) {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}

	if result.StageResolutions["Stage1"].Source != StageTextureSourceExplicit {
		t.Fatalf("expected explicit source for Stage1")
	}
	if result.StageResolutions["Stage4"].Source != StageTextureSourceAutoFill {
		t.Fatalf("expected autofill source for Stage4")
	}
}

func TestGenerateSetStageOverrideHasPriorityOverRoleOverride(t *testing.T) {
	tmp := t.TempDir()
	stagePath := filepath.Join(tmp, "override_stage_nohq.paa")
	rolePath := filepath.Join(tmp, "override_role_nohq.paa")
	mustTouchFile(t, stagePath)
	mustTouchFile(t, rolePath)

	result, err := GenerateSet(GenerateSetOptions{
		TextureOverrides: map[string]string{
			"stage1": stagePath,
			"nohq":   rolePath,
		},
		DisableDamage:   true,
		DisableDestruct: true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil {
		t.Fatalf("missing Stage1")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(stagePath) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if result.StageResolutions["Stage1"].Source != StageTextureSourceExplicit {
		t.Fatalf("expected explicit source for Stage1")
	}
}

func TestGenerateSetForceProceduralOnly(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "obj_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "obj_nohq.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		BaseTexture:         filepath.Join(tmp, "obj_co.paa"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
		ForceProceduralOnly: true,
		TextureOverrides:    nil,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil || !stage1.Texture.IsProcedural() {
		t.Fatalf("expected procedural Stage1 texture in force-procedural mode")
	}
	if result.StageResolutions["Stage1"].Source != StageTextureSourceProcedural {
		t.Fatalf("expected procedural resolution source for Stage1")
	}
}

func TestGenerateSetCustomDamageAndDestructMacro(t *testing.T) {
	result, err := GenerateSet(GenerateSetOptions{
		GenerateDamage:       true,
		GenerateDestruct:     true,
		DamageMacroTexture:   `my\data\custom_damage_mc.paa`,
		DestructMacroTexture: `my\data\custom_destruct_mc.paa`,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}
	if result.Damage == nil || result.Destruct == nil {
		t.Fatalf("expected damage and destruct materials")
	}

	damageStage3 := findMaterialStageByName(result.Damage, "Stage3")
	destructStage3 := findMaterialStageByName(result.Destruct, "Stage3")
	if damageStage3 == nil || destructStage3 == nil {
		t.Fatalf("expected Stage3 in damage/destruct materials")
	}
	if damageStage3.Texture.Raw != `my\data\custom_damage_mc.paa` {
		t.Fatalf("unexpected damage Stage3 texture: %q", damageStage3.Texture.Raw)
	}
	if destructStage3.Texture.Raw != `my\data\custom_destruct_mc.paa` {
		t.Fatalf("unexpected destruct Stage3 texture: %q", destructStage3.Texture.Raw)
	}
}

func TestGenerateSetAutoDiscoversBaseTextureFromOutputPath(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "crate_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_nohq.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		OutputPath: filepath.Join(tmp, "crate"),
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}
	if result.MainOutputPath != filepath.Join(tmp, "crate.rvmat") {
		t.Fatalf("unexpected output path: %q", result.MainOutputPath)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil {
		t.Fatalf("expected Stage1")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_nohq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if result.StageResolutions["Stage1"].Source != StageTextureSourceAutoFill {
		t.Fatalf("expected Stage1 auto-fill source")
	}
}

func TestGenerateSetAutoFillTextureExtensionPriority(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "box_co.png"))
	mustTouchFile(t, filepath.Join(tmp, "box_nohq.png"))
	mustTouchFile(t, filepath.Join(tmp, "box_nohq.tga"))
	mustTouchFile(t, filepath.Join(tmp, "box_nohq.paa"))
	mustTouchFile(t, filepath.Join(tmp, "box_as.tga"))
	mustTouchFile(t, filepath.Join(tmp, "box_smdi.png"))

	result, err := GenerateSet(GenerateSetOptions{
		BaseTexture:         filepath.Join(tmp, "box_co.png"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage4 := findMaterialStageByName(result.Main, "Stage4")
	stage5 := findMaterialStageByName(result.Main, "Stage5")
	if stage1 == nil || stage4 == nil || stage5 == nil {
		t.Fatalf("expected Stage1, Stage4 and Stage5")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_nohq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage4.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_as.tga")) {
		t.Fatalf("unexpected Stage4 texture: %q", stage4.Texture.Raw)
	}
	if stage5.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_smdi.png")) {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}
}

func TestGenerateSetAutoFillTextureExtensionPriorityPAX(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "box_co.png"))
	mustTouchFile(t, filepath.Join(tmp, "box_nohq.tga"))
	mustTouchFile(t, filepath.Join(tmp, "box_nohq.pax"))

	result, err := GenerateSet(GenerateSetOptions{
		BaseTexture:         filepath.Join(tmp, "box_co.png"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil {
		t.Fatalf("expected Stage1")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_nohq.pax")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
}

func TestGenerateSetColorSuffixPreferenceByMaterial(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "box_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "box_ca.tga"))

	glassResult, err := GenerateSet(GenerateSetOptions{
		OutputPath:          filepath.Join(tmp, "box"),
		BaseMaterial:        BaseMaterialGlass,
		TextureAutoFillMode: TextureAutoFillModeDisabled,
		DisableDamage:       true,
		DisableDestruct:     true,
	})
	if err != nil {
		t.Fatalf("generate glass rvmat: %v", err)
	}

	glassStage1 := findMaterialStageByName(glassResult.Main, "Stage1")
	if glassStage1 == nil {
		t.Fatalf("expected Stage1 for glass material")
	}
	if glassStage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_nohq.tga")) {
		t.Fatalf("unexpected glass Stage1 texture: %q", glassStage1.Texture.Raw)
	}

	steelResult, err := GenerateSet(GenerateSetOptions{
		OutputPath:          filepath.Join(tmp, "box"),
		BaseMaterial:        BaseMaterialSteel,
		TextureAutoFillMode: TextureAutoFillModeDisabled,
		DisableDamage:       true,
		DisableDestruct:     true,
	})
	if err != nil {
		t.Fatalf("generate steel rvmat: %v", err)
	}

	steelStage1 := findMaterialStageByName(steelResult.Main, "Stage1")
	if steelStage1 == nil {
		t.Fatalf("expected Stage1 for steel material")
	}
	if steelStage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "box_nohq.paa")) {
		t.Fatalf("unexpected steel Stage1 texture: %q", steelStage1.Texture.Raw)
	}
}

func TestGenerateSetAutoFillSupportsTexconfigSuffixVariants(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "crate_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_nshq.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_detail.tga"))
	mustTouchFile(t, filepath.Join(tmp, "crate_ads.paa"))
	mustTouchFile(t, filepath.Join(tmp, "crate_dtsmdi.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		BaseTexture:         filepath.Join(tmp, "crate_co.paa"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage2 := findMaterialStageByName(result.Main, "Stage2")
	stage4 := findMaterialStageByName(result.Main, "Stage4")
	stage5 := findMaterialStageByName(result.Main, "Stage5")
	if stage1 == nil || stage2 == nil || stage4 == nil || stage5 == nil {
		t.Fatalf("expected Stage1/2/4/5")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_nshq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage2.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_detail.tga")) {
		t.Fatalf("unexpected Stage2 texture: %q", stage2.Texture.Raw)
	}
	if stage4.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_ads.paa")) {
		t.Fatalf("unexpected Stage4 texture: %q", stage4.Texture.Raw)
	}
	if stage5.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "crate_dtsmdi.paa")) {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}
}

func TestGenerateSetAutoFillFromStageOverrideSupportsNormalVariants(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "lamp_nshq.paa"))
	mustTouchFile(t, filepath.Join(tmp, "lamp_as.paa"))
	mustTouchFile(t, filepath.Join(tmp, "lamp_smdi.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		TextureAutoFillMode: TextureAutoFillModeFromStageOverride,
		TextureOverrides: map[string]string{
			"nohq": filepath.Join(tmp, "lamp_nshq.paa"),
		},
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage4 := findMaterialStageByName(result.Main, "Stage4")
	stage5 := findMaterialStageByName(result.Main, "Stage5")
	if stage1 == nil || stage4 == nil || stage5 == nil {
		t.Fatalf("missing expected stages")
	}
	if stage1.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "lamp_nshq.paa")) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage4.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "lamp_as.paa")) {
		t.Fatalf("unexpected Stage4 texture: %q", stage4.Texture.Raw)
	}
	if stage5.Texture.Raw != NormalizeGameTexturePath(filepath.Join(tmp, "lamp_smdi.paa")) {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}
}

func TestGenerateSetDisableTexGenUsesInlineUV(t *testing.T) {
	result, err := GenerateSet(GenerateSetOptions{
		DisableTexGen:   true,
		DisableDamage:   true,
		DisableDestruct: true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}
	if result.Main == nil {
		t.Fatalf("expected main material")
	}
	if len(result.Main.TexGens) != 0 {
		t.Fatalf("expected no texgens when DisableTexGen=true, got %d", len(result.Main.TexGens))
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	if stage1 == nil {
		t.Fatalf("expected Stage1")
	}
	if stage1.TexGen != "" {
		t.Fatalf("expected empty stage texGen, got %q", stage1.TexGen)
	}
	if stage1.UVSource != "tex" || stage1.UVTransform == nil {
		t.Fatalf("expected inline UV data for Stage1 when texgen is disabled")
	}
}

func TestGenerateSetTexturePrefix(t *testing.T) {
	tmp := t.TempDir()
	mustTouchFile(t, filepath.Join(tmp, "testbox_co.paa"))
	mustTouchFile(t, filepath.Join(tmp, "testbox_nohq.paa"))

	result, err := GenerateSet(GenerateSetOptions{
		BaseTexture:         filepath.Join(tmp, "testbox_co.paa"),
		TextureAutoFillMode: TextureAutoFillModeFromBaseTexture,
		TexturePrefix:       "my/some/mod",
		DisableDamage:       true,
		DisableDestruct:     true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	stage1 := findMaterialStageByName(result.Main, "Stage1")
	stage7 := findMaterialStageByName(result.Main, "Stage7")
	if stage1 == nil || stage7 == nil {
		t.Fatalf("expected Stage1 and Stage7")
	}
	if !strings.HasPrefix(stage1.Texture.Raw, `my\some\mod\`) {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if !strings.HasPrefix(stage7.Texture.Raw, `dz\`) {
		t.Fatalf("unexpected Stage7 texture: %q", stage7.Texture.Raw)
	}
}

func TestGenerateSetDisableDamageAndDestruct(t *testing.T) {
	result, err := GenerateSet(GenerateSetOptions{
		DisableDamage:   true,
		DisableDestruct: true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	if result.Damage != nil || result.Destruct != nil {
		t.Fatalf("expected no variants when both disable flags are set")
	}
	if result.DamageOutputPath != "" || result.DestructOutputPath != "" {
		t.Fatalf("expected empty variant output paths when both disable flags are set")
	}
}

func TestGenerateSetGenerateFlagsOverrideDisableFlags(t *testing.T) {
	result, err := GenerateSet(GenerateSetOptions{
		GenerateDamage:   true,
		GenerateDestruct: true,
		DisableDamage:    true,
		DisableDestruct:  true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	if result.Damage == nil || result.Destruct == nil {
		t.Fatalf("expected explicit generate flags to force both variants")
	}
}

func TestWriteGenerateSetUsesFormatIndent(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "testbox.rvmat")

	result, err := GenerateSet(GenerateSetOptions{
		OutputPath:      outPath,
		DisableDamage:   true,
		DisableDestruct: true,
	})
	if err != nil {
		t.Fatalf("generate rvmat: %v", err)
	}

	if err := WriteGenerateSet(result, &FormatOptions{Indent: "  "}); err != nil {
		t.Fatalf("write generated rvmat: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read generated rvmat: %v", err)
	}

	rendered := string(data)
	if !strings.Contains(rendered, "\n  texture=") {
		t.Fatalf("expected two-space indentation in generated output")
	}
	if strings.Contains(rendered, "\n\ttexture=") {
		t.Fatalf("did not expect tab indentation in generated output")
	}
}

func TestNormalizeGameTexturePath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "relative dot slash",
			in:   "./MyMod/assets/data/texture_nohq.paa",
			want: `mymod\assets\data\texture_nohq.paa`,
		},
		{
			name: "absolute slash",
			in:   "/MyMod/data/texture_nohq.paa",
			want: `mymod\data\texture_nohq.paa`,
		},
		{
			name: "windows drive",
			in:   `P:\MyMod\data\texture_nohq.paa`,
			want: `mymod\data\texture_nohq.paa`,
		},
		{
			name: "procedural unchanged",
			in:   "#(argb,8,8,3)color(0.5,0.5,1,1,nohq)",
			want: "#(argb,8,8,3)color(0.5,0.5,1,1,nohq)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeGameTexturePath(tt.in)
			if got != tt.want {
				t.Fatalf("NormalizeGameTexturePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// mustTouchFile creates an empty file and fails test on error.
func mustTouchFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatalf("touch file %s: %v", path, err)
	}
}
