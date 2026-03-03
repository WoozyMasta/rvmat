package rvmat

import (
	"strings"
	"testing"
)

func TestFormatNormalizesTexturePaths(t *testing.T) {
	mat := &Material{
		Stages: []Stage{
			{
				Name:    "Stage1",
				Texture: ParseTextureRef("./wiki/test/testbox_nohq.paa"),
			},
			{
				Name:    "Stage2",
				Texture: ParseTextureRef(`P:\wiki\test\testbox_dt.paa`),
			},
			{
				Name:    "Stage6",
				Texture: ParseTextureRef("#(ai,64,1,1)fresnel(1.32,0.94)"),
			},
		},
	}

	out, err := Format(mat, nil)
	if err != nil {
		t.Fatalf("format material: %v", err)
	}

	rendered := string(out)
	if !strings.Contains(rendered, `texture="wiki\test\testbox_nohq.paa";`) {
		t.Fatalf("normalized relative path not found:\n%s", rendered)
	}
	if !strings.Contains(rendered, `texture="wiki\test\testbox_dt.paa";`) {
		t.Fatalf("normalized drive path not found:\n%s", rendered)
	}
	if !strings.Contains(rendered, `texture="#(ai,64,1,1)fresnel(1.32,0.94)";`) {
		t.Fatalf("procedural texture should stay unchanged:\n%s", rendered)
	}
}

func TestFormatPrettyFloat(t *testing.T) {
	power := 5.280000000000001
	mat := &Material{
		SpecularPower: &power,
		Stages: []Stage{
			{
				Name:    "Stage1",
				Texture: NewProceduralColor("argb", 8, 8, 3, 1, 0.3333333333333333, 0.5, 1, "smdi"),
			},
		},
	}

	out, err := Format(mat, nil)
	if err != nil {
		t.Fatalf("format material: %v", err)
	}

	rendered := string(out)
	if !strings.Contains(rendered, "specularPower=5.28;") {
		t.Fatalf("expected pretty specularPower, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "specularPower=5.280000000000001;") {
		t.Fatalf("unexpected IEEE tail in specularPower, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "0.3333") {
		t.Fatalf("expected pretty procedural float, got:\n%s", rendered)
	}
}

func TestFormatCompactStages(t *testing.T) {
	mat := &Material{
		Stages: []Stage{
			{
				Name:    "Stage1",
				Texture: ParseTextureRef(`./mymod/data/item_nohq.paa`),
				TexGen:  "0",
			},
			{
				Name:    "Stage2",
				Texture: ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1,DT)`),
				TexGen:  "TexGen0",
			},
			{
				Name:    "StageTI",
				Texture: ParseTextureRef(`dz/data/thermal_ti.paa`),
				TexGen:  "0",
			},
		},
	}

	out, err := Format(mat, &FormatOptions{
		CompactStages: true,
	})
	if err != nil {
		t.Fatalf("format material: %v", err)
	}

	rendered := string(out)
	if !strings.Contains(rendered, `class Stage1 { texture = "mymod\data\item_nohq.paa"; texGen = 0; };`) {
		t.Fatalf("expected compact Stage1, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, `class Stage2 { texture = "#(argb,8,8,3)color(0.5,0.5,0.5,1,DT)"; texGen = "TexGen0"; };`) {
		t.Fatalf("expected compact Stage2, got:\n%s", rendered)
	}
	if strings.Contains(rendered, `class StageTI { texture =`) {
		t.Fatalf("did not expect compact StageTI, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "class StageTI\n{\n") {
		t.Fatalf("expected regular multi-line StageTI block, got:\n%s", rendered)
	}
}
