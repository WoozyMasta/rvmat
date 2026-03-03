package rvmat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateGolden(t *testing.T) {
	tests := []struct {
		name       string
		goldenPath string
		opts       GenerateOptions
	}{
		{
			name:       "textile_default",
			goldenPath: filepath.Join("testdata", "generate", "textile_default.rvmat"),
			opts: GenerateOptions{
				BaseMaterial: BaseMaterialTextile,
				UseTexGen:    true,
			},
		},
		{
			name:       "steel_gloss_worn",
			goldenPath: filepath.Join("testdata", "generate", "steel_gloss_worn.rvmat"),
			opts: GenerateOptions{
				BaseMaterial: BaseMaterialSteel,
				Finish:       FinishGloss,
				Condition:    ConditionWorn,
				UseTexGen:    true,
			},
		},
		{
			name:       "glass_default",
			goldenPath: filepath.Join("testdata", "generate", "glass_default.rvmat"),
			opts: GenerateOptions{
				BaseMaterial: BaseMaterialGlass,
				UseTexGen:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat, err := Generate(tt.opts)
			if err != nil {
				t.Fatalf("generate material: %v", err)
			}

			assertMaterialGolden(t, tt.goldenPath, mat)
		})
	}
}

func TestGenerateVariantGolden(t *testing.T) {
	base, err := Generate(GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		UseTexGen:    true,
	})
	if err != nil {
		t.Fatalf("generate base material: %v", err)
	}

	damage, err := GenerateDamage(base)
	if err != nil {
		t.Fatalf("generate damage variant: %v", err)
	}
	assertMaterialGolden(
		t,
		filepath.Join("testdata", "generate", "steel_damage.rvmat"),
		damage,
	)

	destruct, err := GenerateDestruct(base)
	if err != nil {
		t.Fatalf("generate destruct variant: %v", err)
	}
	assertMaterialGolden(
		t,
		filepath.Join("testdata", "generate", "steel_destruct.rvmat"),
		destruct,
	)
}

func assertMaterialGolden(t *testing.T, goldenPath string, m *Material) {
	t.Helper()

	got, err := Format(m, nil)
	if err != nil {
		t.Fatalf("format material: %v", err)
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v", goldenPath, err)
	}

	if string(got) != string(want) {
		t.Fatalf("golden mismatch for %s", goldenPath)
	}
}
