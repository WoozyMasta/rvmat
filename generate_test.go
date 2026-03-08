package rvmat

import (
	"errors"
	"testing"
)

func TestGenerateBaseMaterials(t *testing.T) {
	tests := []struct {
		name       string
		material   BaseMaterial
		wantPower  float64
		wantFn     string
		wantStages int
	}{
		{
			name:       "textile",
			material:   BaseMaterialTextile,
			wantPower:  55,
			wantFn:     "fresnel",
			wantStages: 7,
		},
		{
			name:       "steel",
			material:   BaseMaterialSteel,
			wantPower:  80,
			wantFn:     "fresnel",
			wantStages: 7,
		},
		{
			name:       "glass",
			material:   BaseMaterialGlass,
			wantPower:  500,
			wantFn:     "fresnelGlass",
			wantStages: 7,
		},
		{
			name:       "concrete",
			material:   BaseMaterialConcrete,
			wantPower:  80,
			wantFn:     "fresnel",
			wantStages: 7,
		},
		{
			name:       "skin",
			material:   BaseMaterialSkin,
			wantPower:  80,
			wantFn:     "fresnel",
			wantStages: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat, err := Generate(GenerateOptions{
				BaseMaterial: tt.material,
			})
			if err != nil {
				t.Fatalf("generate material: %v", err)
			}
			if mat.SpecularPower == nil || *mat.SpecularPower != tt.wantPower {
				t.Fatalf("unexpected specular power: %v", mat.SpecularPower)
			}
			if len(mat.Stages) != tt.wantStages {
				t.Fatalf("unexpected stage count: %d", len(mat.Stages))
			}

			stage6 := findMaterialStageByName(mat, "Stage6")
			if stage6 == nil || stage6.Texture.Procedural == nil {
				t.Fatalf("expected procedural Stage6 texture")
			}
			if stage6.Texture.Procedural.Func != tt.wantFn {
				t.Fatalf("unexpected Stage6 procedural fn: %q", stage6.Texture.Procedural.Func)
			}
		})
	}
}

func TestGenerateTextureDeriveAndTexGen(t *testing.T) {
	mat, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		BaseTexture:  `my/path/item_co.paa`,
		UseTexGen:    true,
	})
	if err != nil {
		t.Fatalf("generate material: %v", err)
	}

	stage1 := findMaterialStageByName(mat, "Stage1")
	stage5 := findMaterialStageByName(mat, "Stage5")
	stage7 := findMaterialStageByName(mat, "Stage7")
	if stage1 == nil || stage5 == nil || stage7 == nil {
		t.Fatalf("missing generated stages")
	}
	if stage1.Texture.Raw != `my/path/item_nohq.paa` {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage5.Texture.Raw != `my/path/item_smdi.paa` {
		t.Fatalf("unexpected Stage5 texture: %q", stage5.Texture.Raw)
	}
	if stage7.Texture.Raw != DefaultEnvironmentTexture {
		t.Fatalf("unexpected Stage7 texture: %q", stage7.Texture.Raw)
	}
	if stage1.TexGen != "0" {
		t.Fatalf("expected Stage1 texGen=0, got %q", stage1.TexGen)
	}
	if len(mat.TexGens) != 1 {
		t.Fatalf("unexpected texgen count: %d", len(mat.TexGens))
	}
}

func TestGenerateStageOverrideHasPriorityOverRoleOverride(t *testing.T) {
	mat, err := Generate(GenerateOptions{
		TextureOverrides: map[string]string{
			"stage1": `my/path/stage_nohq.paa`,
			"nohq":   `my/path/role_nohq.paa`,
		},
	})
	if err != nil {
		t.Fatalf("generate material: %v", err)
	}

	stage1 := findMaterialStageByName(mat, "Stage1")
	if stage1 == nil {
		t.Fatalf("missing Stage1")
	}
	if stage1.Texture.Raw != `my/path/stage_nohq.paa` {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
}

func TestGenerateDamageAndDestructVariants(t *testing.T) {
	base, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
	})
	if err != nil {
		t.Fatalf("generate base material: %v", err)
	}
	baseStage3 := findMaterialStageByName(base, "Stage3")
	if baseStage3 == nil {
		t.Fatalf("base Stage3 missing")
	}
	baseRaw := baseStage3.Texture.Raw

	damage, err := GenerateDamage(base)
	if err != nil {
		t.Fatalf("generate damage variant: %v", err)
	}
	damageStage3 := findMaterialStageByName(damage, "Stage3")
	if damageStage3 == nil {
		t.Fatalf("damage Stage3 missing")
	}
	if damageStage3.Texture.Raw != DefaultDamageMacroTexture {
		t.Fatalf("unexpected damage Stage3 texture: %q", damageStage3.Texture.Raw)
	}

	destruct, err := GenerateDestruct(base)
	if err != nil {
		t.Fatalf("generate destruct variant: %v", err)
	}
	destructStage3 := findMaterialStageByName(destruct, "Stage3")
	if destructStage3 == nil {
		t.Fatalf("destruct Stage3 missing")
	}
	if destructStage3.Texture.Raw != DefaultDestructMacroTexture {
		t.Fatalf("unexpected destruct Stage3 texture: %q", destructStage3.Texture.Raw)
	}

	// Base material must stay unchanged.
	if baseStage3.Texture.Raw != baseRaw {
		t.Fatalf("base material was modified, got %q want %q", baseStage3.Texture.Raw, baseRaw)
	}
}

func TestGenerateVariantFlags(t *testing.T) {
	if _, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		WithDamage:   true,
		WithDestruct: true,
	}); err == nil {
		t.Fatalf("expected error for both variant flags")
	}

	damage, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		WithDamage:   true,
	})
	if err != nil {
		t.Fatalf("generate damage via flags: %v", err)
	}
	if findMaterialStageByName(damage, "Stage3").Texture.Raw != DefaultDamageMacroTexture {
		t.Fatalf("expected damage Stage3 texture")
	}
}

func TestGenerateUnknownBaseMaterial(t *testing.T) {
	_, err := Generate(GenerateOptions{BaseMaterial: BaseMaterial(255)})
	if err == nil {
		t.Fatalf("expected error for unknown base material")
	}
	if !errors.Is(err, ErrUnknownBaseMaterial) {
		t.Fatalf("expected ErrUnknownBaseMaterial, got %v", err)
	}
}

func TestGenerateFallbackProceduralDefaults(t *testing.T) {
	mat, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialTextile,
	})
	if err != nil {
		t.Fatalf("generate material: %v", err)
	}

	stage1 := findMaterialStageByName(mat, "Stage1")
	stage3 := findMaterialStageByName(mat, "Stage3")
	if stage1 == nil || stage3 == nil {
		t.Fatalf("expected Stage1 and Stage3")
	}

	if stage1.Texture.Raw != `#(argb,8,8,3)color(0.5,0.5,1,1,NOHQ)` {
		t.Fatalf("unexpected Stage1 texture: %q", stage1.Texture.Raw)
	}
	if stage3.Texture.Raw != `#(argb,8,8,3)color(0.5,0.5,0.5,0,MC)` {
		t.Fatalf("unexpected Stage3 texture: %q", stage3.Texture.Raw)
	}
}

func TestGenerateFallbackASAndSMDIAreMaterialAware(t *testing.T) {
	textile, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialTextile,
	})
	if err != nil {
		t.Fatalf("generate textile material: %v", err)
	}
	metal, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
	})
	if err != nil {
		t.Fatalf("generate metal material: %v", err)
	}

	textileAS := findMaterialStageByName(textile, "Stage4")
	metalAS := findMaterialStageByName(metal, "Stage4")
	textileSMDI := findMaterialStageByName(textile, "Stage5")
	metalSMDI := findMaterialStageByName(metal, "Stage5")
	if textileAS == nil || metalAS == nil || textileSMDI == nil || metalSMDI == nil {
		t.Fatalf("expected Stage4/Stage5 in both materials")
	}

	textileASColor := textileAS.Texture.Procedural.Color
	metalASColor := metalAS.Texture.Procedural.Color
	if textileASColor == nil || metalASColor == nil {
		t.Fatalf("expected procedural color fallback for Stage4")
	}
	if !(metalASColor.G > textileASColor.G) {
		t.Fatalf("expected metal AS to be brighter than textile AS: metal=%f textile=%f", metalASColor.G, textileASColor.G)
	}

	textileSMDIColor := textileSMDI.Texture.Procedural.Color
	metalSMDIColor := metalSMDI.Texture.Procedural.Color
	if textileSMDIColor == nil || metalSMDIColor == nil {
		t.Fatalf("expected procedural color fallback for Stage5")
	}
	if textileSMDIColor.R != 1 || metalSMDIColor.R != 1 {
		t.Fatalf("expected SMDI.R=1, got textile=%f metal=%f", textileSMDIColor.R, metalSMDIColor.R)
	}
	if !(metalSMDIColor.G > textileSMDIColor.G) {
		t.Fatalf("expected metal SMDI spec (G) > textile: metal=%f textile=%f", metalSMDIColor.G, textileSMDIColor.G)
	}
	if !(metalSMDIColor.B > textileSMDIColor.B) {
		t.Fatalf("expected metal SMDI gloss (B) > textile: metal=%f textile=%f", metalSMDIColor.B, textileSMDIColor.B)
	}
}

func TestGenerateFallbackMCAlphaIsZero(t *testing.T) {
	mat, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
	})
	if err != nil {
		t.Fatalf("generate material: %v", err)
	}

	stage3 := findMaterialStageByName(mat, "Stage3")
	if stage3 == nil {
		t.Fatalf("expected Stage3")
	}
	if stage3.Texture.Procedural == nil || stage3.Texture.Procedural.Color == nil {
		t.Fatalf("expected procedural color in Stage3")
	}
	if stage3.Texture.Procedural.Color.A != 0 {
		t.Fatalf("expected Stage3 alpha=0, got %v", stage3.Texture.Procedural.Color.A)
	}
}

func TestGenerateEmissiveIntensity(t *testing.T) {
	mat, err := Generate(GenerateOptions{
		BaseMaterial:      BaseMaterialSteel,
		EmissiveIntensity: 0.35,
	})
	if err != nil {
		t.Fatalf("generate material: %v", err)
	}
	if len(mat.Emissive) < 3 {
		t.Fatalf("unexpected emissive length: %d", len(mat.Emissive))
	}
	if mat.Emissive[0] != 0.35 || mat.Emissive[1] != 0.35 || mat.Emissive[2] != 0.35 {
		t.Fatalf("unexpected emissive values: %#v", mat.Emissive)
	}
}

func TestGenerateInvalidFinishAndCondition(t *testing.T) {
	_, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		Finish:       Finish(255),
	})
	if err == nil {
		t.Fatalf("expected error for invalid finish")
	}
	if !errors.Is(err, ErrInvalidGenerateOption) {
		t.Fatalf("expected ErrInvalidGenerateOption for finish, got %v", err)
	}

	_, err = Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		Condition:    Condition(255),
	})
	if err == nil {
		t.Fatalf("expected error for invalid condition")
	}
	if !errors.Is(err, ErrInvalidGenerateOption) {
		t.Fatalf("expected ErrInvalidGenerateOption for condition, got %v", err)
	}

	_, err = Generate(GenerateOptions{
		BaseMaterial: BaseMaterial(255),
	})
	if err == nil {
		t.Fatalf("expected error for invalid base material")
	}
	if !errors.Is(err, ErrUnknownBaseMaterial) {
		t.Fatalf("expected ErrUnknownBaseMaterial for base material, got %v", err)
	}

}
