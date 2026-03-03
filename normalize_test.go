package rvmat

import (
	"reflect"
	"testing"
)

func TestNormalizeNil(t *testing.T) {
	result, issues := Normalize(nil, nil)

	if result.Changed {
		t.Fatalf("unexpected changed result: %+v", result)
	}
	if len(issues) != 1 {
		t.Fatalf("unexpected issues count: %d (%v)", len(issues), issues)
	}
	if issues[0].Level != IssueError {
		t.Fatalf("unexpected issue level: %v", issues[0])
	}
}

func TestNormalizeFillMissingStageTexture(t *testing.T) {
	mat := &Material{
		Stages: []Stage{
			{Name: "Stage1"},
			{Name: "Stage2"},
			{Name: "Stage6"},
			{Name: "CustomStage"},
		},
	}

	result, issues := Normalize(mat, &NormalizeOptions{
		StageTextures: true,
	})

	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %v", issues)
	}
	if result.StageTexturesFilled != 3 {
		t.Fatalf("unexpected stage texture normalize count: %+v", result)
	}
	if mat.Stages[0].Texture.Raw == "" || mat.Stages[1].Texture.Raw == "" || mat.Stages[2].Texture.Raw == "" {
		t.Fatalf("expected fallback textures on known stages")
	}
	if mat.Stages[3].Texture.Raw != "" {
		t.Fatalf("unexpected fallback texture on custom stage: %q", mat.Stages[3].Texture.Raw)
	}
}

func TestNormalizeOrder(t *testing.T) {
	mat := &Material{
		Stages: []Stage{
			{Name: "Stage3"},
			{Name: "StageTI"},
			{Name: "Custom"},
			{Name: "Stage1"},
			{Name: "Stage2"},
		},
		TexGens: []TexGen{
			{Name: "TexGen2"},
			{Name: "Custom"},
			{Name: "TexGen0"},
			{Name: "TexGen1"},
		},
	}

	result, issues := Normalize(mat, &NormalizeOptions{
		StageOrder:  true,
		TexGenOrder: true,
	})

	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %v", issues)
	}
	if !result.StageOrderNormalized || !result.TexGenOrderNormalized {
		t.Fatalf("unexpected normalize result: %+v", result)
	}

	stageNames := make([]string, 0, len(mat.Stages))
	for _, st := range mat.Stages {
		stageNames = append(stageNames, st.Name)
	}
	if !reflect.DeepEqual(stageNames, []string{"Stage1", "Stage2", "Stage3", "StageTI", "Custom"}) {
		t.Fatalf("unexpected stage order: %v", stageNames)
	}

	texGenNames := make([]string, 0, len(mat.TexGens))
	for _, tg := range mat.TexGens {
		texGenNames = append(texGenNames, tg.Name)
	}
	if !reflect.DeepEqual(texGenNames, []string{"TexGen0", "TexGen1", "TexGen2", "Custom"}) {
		t.Fatalf("unexpected texgen order: %v", texGenNames)
	}
}

func TestNormalizeTexturePaths(t *testing.T) {
	mat := &Material{
		Stages: []Stage{
			{Name: "Stage1", Texture: ParseTextureRef("./assets/data/texture_nohq.paa")},
			{Name: "Stage2", Texture: ParseTextureRef("/assets/data/texture_dt.paa")},
			{Name: "Stage3", Texture: ParseTextureRef(`P:\assets\data\texture_mc.paa`)},
			{Name: "Stage6", Texture: ParseTextureRef("#(ai,64,1,1)fresnel(1.32,0.94)")},
		},
	}

	result, issues := Normalize(mat, &NormalizeOptions{
		TexturePaths: true,
	})

	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %v", issues)
	}
	if !result.Changed {
		t.Fatalf("expected changed result")
	}
	if result.TexturePathsNormalized != 3 {
		t.Fatalf("unexpected normalized path count: %+v", result)
	}
	if mat.Stages[0].Texture.Raw != `assets\data\texture_nohq.paa` {
		t.Fatalf("unexpected Stage1 texture: %q", mat.Stages[0].Texture.Raw)
	}
	if mat.Stages[1].Texture.Raw != `assets\data\texture_dt.paa` {
		t.Fatalf("unexpected Stage2 texture: %q", mat.Stages[1].Texture.Raw)
	}
	if mat.Stages[2].Texture.Raw != `assets\data\texture_mc.paa` {
		t.Fatalf("unexpected Stage3 texture: %q", mat.Stages[2].Texture.Raw)
	}
	if mat.Stages[3].Texture.Raw != "#(ai,64,1,1)fresnel(1.32,0.94)" {
		t.Fatalf("unexpected Stage6 texture: %q", mat.Stages[3].Texture.Raw)
	}
}
