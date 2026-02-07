package rvmat

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseSamples(t *testing.T) {
	files := []string{
		"basic.rvmat",
		"multi.rvmat",
		"minimal.rvmat",
	}
	for _, f := range files {
		m, err := DecodeFile(filepath.Join("testdata", f), nil)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		if f == "multi.rvmat" {
			if len(m.Stages) == 0 || len(m.TexGens) == 0 {
				t.Fatalf("expected stages and texgens in %s", f)
			}
		}
	}
}

func TestRoundTrip(t *testing.T) {
	m, err := DecodeFile(filepath.Join("testdata", "multi.rvmat"), nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	b, err := Format(m, nil)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	m2, err := Parse(b, nil)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if len(m2.Stages) != len(m.Stages) {
		t.Fatalf("stage count mismatch: %d vs %d", len(m2.Stages), len(m.Stages))
	}
}

func TestProceduralTextureParse(t *testing.T) {
	in := "#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)"
	tr := ParseTextureRef(in)
	if !tr.IsProcedural() || !tr.ParsedOK || tr.Procedural == nil {
		t.Fatalf("expected parsed procedural texture")
	}
	if tr.Procedural.Format != "argb" || tr.Procedural.Func != "color" {
		t.Fatalf("unexpected parse result")
	}
}

func TestPathResolver(t *testing.T) {
	resolver := PathResolver{GameRoot: "P:\\"}
	raw := "dz\\vehicles\\wheeled\\offroad_02\\data\\offroad_02_roof_co.paa"
	got := resolver.ResolvePath(raw)
	want := filepath.Clean(filepath.Join("P:\\", raw))
	if got != want {
		t.Fatalf("resolve mismatch: %q != %q", got, want)
	}
}

func TestRoundTripMinimalMaterial(t *testing.T) {
	var out []byte
	var err error

	want := &Material{
		PixelShaderID:  "Super",
		VertexShaderID: "Super",
		Ambient:        []float64{1, 1, 1, 1},
		Diffuse:        []float64{1, 1, 1, 1},
		ForcedDiffuse:  []float64{0, 0, 0, 0},
		Emmisive:       []float64{0, 0, 0, 1},
		Specular:       []float64{0.75, 0.75, 0.75, 1},
		Stages: []Stage{
			{
				Name:     "Stage1",
				Texture:  ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
				UVSource: "tex",
				UVTransform: &UVTransform{
					Aside: []float64{1, 0, 0},
					Up:    []float64{0, 1, 0},
					Dir:   []float64{0, 0, 0},
					Pos:   []float64{0, 0, 0},
				},
			},
		},
	}
	out, err = Format(want, nil)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	got, err := Parse(out, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	vopt := &ValidateOptions{GameRoot: `P:\`}
	if os.Getenv("CI") != "" {
		vopt.DisableFileCheck = true
	} else if !vopt.IsGameRootExist() {
		vopt.DisableFileCheck = true
	}
	issues := Validate(got, vopt)
	if len(issues) != 0 {
		t.Fatalf("unexpected validation issues: %v", issues)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestRoundTripFullMaterial(t *testing.T) {
	var out []byte
	var err error

	want := &Material{
		PixelShaderID:  "Multi",
		VertexShaderID: "Multi",
		Ambient:        []float64{4, 4, 4, 1},
		Diffuse:        []float64{4, 4, 4, 1},
		ForcedDiffuse:  []float64{0, 0, 0, 0},
		Emmisive:       []float64{0, 0, 0, 1},
		Specular:       []float64{0.2, 0.2, 0.2, 1},
		SpecularPower:  floatPtr(1500),
		TexGens: []TexGen{
			{
				Name:     "TexGen0",
				UVSource: "tex",
				UVTransform: &UVTransform{
					Aside: []float64{1, 0, 0},
					Up:    []float64{0, 1, 0},
					Dir:   []float64{0, 0, 1},
					Pos:   []float64{0, 0, 0},
				},
			},
		},
		Stages: []Stage{
			{
				Name:    "Stage0",
				Texture: ParseTextureRef(`dz\vehicles\wheeled\offroad_02\data\offroad_02_roof_co.paa`),
				TexGen:  "0",
			},
			{
				Name:    "Stage1",
				Texture: NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1.0, "co"),
				TexGen:  "1",
			},
			{
				Name:    "Stage2",
				Texture: NewProceduralFresnel("ai", 32, 128, 1, 1.96, 0.01),
				TexGen:  "2",
			},
			{
				Name:    "Stage3",
				Texture: NewProceduralFresnelGlass("ai", 32, 128, 1, 1.7, 0, false),
				TexGen:  "3",
			},
			{
				Name:    "Stage4",
				Texture: NewProceduralIrradiance("ai", 32, 128, 1, 3),
				TexGen:  "4",
			},
			{
				Name:     "Stage5",
				Texture:  ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
				UVSource: "tex",
				UVTransform: &UVTransform{
					Aside: []float64{1, 0, 0},
					Up:    []float64{0, 1, 0},
					Dir:   []float64{0, 0, 0},
					Pos:   []float64{0, 0, 0},
				},
			},
		},
	}
	out, err = Format(want, nil)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	got, err := Parse(out, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	vopt := &ValidateOptions{GameRoot: `P:\`}
	if os.Getenv("CI") != "" {
		vopt.DisableFileCheck = true
	} else if !vopt.IsGameRootExist() {
		vopt.DisableFileCheck = true
	}
	issues := ValidateWithTextureOptions(got, vopt, &TextureValidateOptions{
		DisableProceduralFnCheck:   false,
		DisableProceduralArgsCheck: false,
		DisableTextureTagCheck:     false,
	})
	if len(issues) != 0 {
		t.Fatalf("unexpected validation issues: %v", issues)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch")
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestValidateTable(t *testing.T) {
	tests := []struct {
		name     string
		mat      *Material
		opt      *ValidateOptions
		wantWarn int
		wantErr  int
	}{
		{
			name: "ok_minimal",
			mat: &Material{
				PixelShaderID:  "Super",
				VertexShaderID: "Super",
				Ambient:        []float64{1, 1, 1, 1},
				Diffuse:        []float64{1, 1, 1, 1},
				ForcedDiffuse:  []float64{0, 0, 0, 0},
				Emmisive:       []float64{0, 0, 0, 1},
				Specular:       []float64{0.75, 0.75, 0.75, 1},
				Stages: []Stage{
					{
						Name:     "Stage1",
						Texture:  ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
						UVSource: "tex",
						UVTransform: &UVTransform{
							Aside: []float64{1, 0, 0},
							Up:    []float64{0, 1, 0},
							Dir:   []float64{0, 0, 0},
							Pos:   []float64{0, 0, 0},
						},
					},
				},
			},
			opt:      &ValidateOptions{DisableFileCheck: true},
			wantWarn: 0,
			wantErr:  0,
		},
		{
			name: "missing_uv_without_texgen",
			mat: &Material{
				PixelShaderID:  "Super",
				VertexShaderID: "Super",
				Stages: []Stage{
					{
						Name:    "Stage1",
						Texture: ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
					},
				},
			},
			opt:      &ValidateOptions{DisableFileCheck: true, DisableShaderNameCheck: true},
			wantWarn: 2,
			wantErr:  0,
		},
		{
			name: "duplicate_stage_name",
			mat: &Material{
				PixelShaderID:  "Super",
				VertexShaderID: "Super",
				Stages: []Stage{
					{Name: "Stage1"},
					{Name: "Stage1"},
				},
			},
			opt:      &ValidateOptions{DisableFileCheck: true, DisableShaderNameCheck: true},
			wantWarn: 4,
			wantErr:  1,
		},
		{
			name: "unknown_shader_names",
			mat: &Material{
				PixelShaderID:  "UnknownPS",
				VertexShaderID: "UnknownVS",
			},
			opt:      &ValidateOptions{DisableFileCheck: true, DisableShaderNameCheck: false},
			wantWarn: 2,
			wantErr:  0,
		},
		{
			name: "extension_warning",
			mat: &Material{
				PixelShaderID:  "Super",
				VertexShaderID: "Super",
				Stages: []Stage{
					{
						Name:     "Stage1",
						Texture:  ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.png`),
						UVSource: "tex",
						UVTransform: &UVTransform{
							Aside: []float64{1, 0, 0},
							Up:    []float64{0, 1, 0},
							Dir:   []float64{0, 0, 0},
							Pos:   []float64{0, 0, 0},
						},
					},
				},
			},
			opt:      &ValidateOptions{DisableFileCheck: true, DisableExtensionsCheck: false},
			wantWarn: 1,
			wantErr:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := Validate(tt.mat, tt.opt)
			var warns, errs int
			for _, it := range issues {
				switch it.Level {
				case IssueWarning:
					warns++
				case IssueError:
					errs++
				}
			}
			if warns != tt.wantWarn || errs != tt.wantErr {
				t.Fatalf("unexpected issues: warnings=%d errors=%d issues=%v", warns, errs, issues)
			}
		})
	}
}

func TestValidateTextureTable(t *testing.T) {
	tests := []struct {
		name     string
		tex      TextureRef
		opt      *TextureValidateOptions
		wantWarn int
		wantErr  int
	}{
		{
			name:     "path_texture_no_checks",
			tex:      ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
			opt:      &TextureValidateOptions{},
			wantWarn: 0,
			wantErr:  0,
		},
		{
			name:     "valid_color",
			tex:      ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`),
			opt:      &TextureValidateOptions{},
			wantWarn: 0,
			wantErr:  0,
		},
		{
			name: "unknown_fn",
			tex:  ParseTextureRef(`#(argb,8,8,3)unknown(1,2)`),
			opt: &TextureValidateOptions{
				DisableProceduralFnCheck: false,
			},
			wantWarn: 1,
			wantErr:  0,
		},
		{
			name: "bad_args_count",
			tex:  ParseTextureRef(`#(argb,8,8,3)color(1,1,1)`),
			opt: &TextureValidateOptions{
				DisableProceduralArgsCheck: false,
			},
			wantWarn: 1,
			wantErr:  0,
		},
		{
			name: "unknown_tag",
			tex:  ParseTextureRef(`#(argb,8,8,3)color(1,1,1,1,wat)`),
			opt: &TextureValidateOptions{
				DisableTextureTagCheck: false,
			},
			wantWarn: 1,
			wantErr:  0,
		},
		{
			name: "parse_failed_reports",
			tex:  ParseTextureRef(`#(argb,8,8,3)color(1,1,1,`),
			opt: &TextureValidateOptions{
				DisableProceduralFnCheck:   false,
				DisableProceduralArgsCheck: false,
				DisableTextureTagCheck:     false,
			},
			wantWarn: 1,
			wantErr:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.tex.Validate(tt.opt)
			var warns, errs int
			for _, it := range issues {
				switch it.Level {
				case IssueWarning:
					warns++
				case IssueError:
					errs++
				}
			}
			if warns != tt.wantWarn || errs != tt.wantErr {
				t.Fatalf("unexpected issues: warnings=%d errors=%d issues=%v", warns, errs, issues)
			}
		})
	}
}

func TestParseComments(t *testing.T) {
	input := `// top comment
ambient[] = { 1, 1, 1, 1 }; /* mid */
PixelShaderID = "Super";
VertexShaderID = "Super";
class Stage1 {
    texture = "dz\\x\\y.paa"; // end
    uvSource = "tex";
    class uvTransform { aside[] = { 1, 0, 0 }; up[] = { 0, 1, 0 }; dir[] = { 0, 0, 0 }; pos[] = { 0, 0, 0 }; };
};
`
	if _, err := Parse([]byte(input), nil); err != nil {
		t.Fatalf("parse with comments: %v", err)
	}
	if _, err := Parse([]byte(input), &ParseOptions{DisableComments: true}); err == nil {
		t.Fatalf("expected error with comments disabled")
	}
}

func TestCaseInsensitiveKeys(t *testing.T) {
	input := `DiFfUse[] = { 1, 1, 1, 1 };
PixelShaderID = "Super";
VertexShaderID = "Super";
`
	m, err := Parse([]byte(input), nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(m.Diffuse) != 4 {
		t.Fatalf("expected diffuse parsed, got %v", m.Diffuse)
	}

	m2, err := Parse([]byte(input), &ParseOptions{DisableCaseInsensitive: true})
	if err != nil {
		t.Fatalf("parse case-sensitive: %v", err)
	}
	if len(m2.Diffuse) != 0 {
		t.Fatalf("expected diffuse empty with case-sensitive parsing")
	}
}

func TestRelaxedNumbers(t *testing.T) {
	input := `diffuse[] = { 0.75, 1.5, "1.25.1", 0.0 };
PixelShaderID = "Super";
VertexShaderID = "Super";
`
	if _, err := Parse([]byte(input), nil); err != nil {
		t.Fatalf("parse relaxed: %v", err)
	}
	if _, err := Parse([]byte(input), &ParseOptions{DisableRelaxedNumbers: true}); err == nil {
		t.Fatalf("expected error with relaxed numbers disabled")
	}
}

func TestUnknownFieldsRoundTrip(t *testing.T) {
	input := `mainLight = "Sun";
ambient[] = { 1, 1, 1, 1 };
PixelShaderID = "Super";
VertexShaderID = "Super";
class Stage1 {
    texture = "dz\\x\\y.paa";
    uvSource = "tex";
    class uvTransform { aside[] = { 1, 0, 0 }; up[] = { 0, 1, 0 }; dir[] = { 0, 0, 0 }; pos[] = { 0, 0, 0 }; };
    class ExtraBlock { foo = 1; };
};
class CustomTop { bar = 2; };
`
	m, err := Parse([]byte(input), nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := Format(m, nil)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "mainLight") {
		t.Fatalf("expected mainLight in output")
	}
	if !strings.Contains(s, "class CustomTop") {
		t.Fatalf("expected CustomTop in output")
	}
	if !strings.Contains(s, "class ExtraBlock") {
		t.Fatalf("expected ExtraBlock in output")
	}
}
