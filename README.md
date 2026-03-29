# rvmat

Go library for parsing, validating, and writing
[Real Virtuality `.rvmat` material files](./RVMAT.md).

* Fast parser with tolerances for real-world data
  (case-insensitive keys, comments, relaxed numeric arrays).
* Deterministic writer (canonical formatting) with configurable indentation.
* Helpers for procedural textures, path resolution, and validation.
* Normalization API for stage order, TexGen order, fallback textures,
  and texture path cleanup.
* Generation APIs for baseline `Super` materials, damage/destruct variants,
  and high-level material set output.
* Validator supports configurable checks (shader/stage allowlists are optional).
* Unknown fields are preserved internally and round-tripped.
* Lint diagnostics registry [RULES.md](RULES.md).

## Install

```bash
go get github.com/woozymasta/rvmat
```

## Quick Start

```go
m, err := rvmat.DecodeFile("material.rvmat", nil)
if err != nil {
  // handle
}

// Validate with defaults
diagnostics := rvmat.Validate(m, nil)

// Render canonical output
out, err := rvmat.Format(m, nil)
```

## Examples

Create a minimal material with a procedural texture:

```go
m := &rvmat.Material{
  PixelShaderID:  "Super",
  VertexShaderID: "Super",
  Ambient:        []float64{1, 1, 1, 1},
  Diffuse:        []float64{1, 1, 1, 1},
  ForcedDiffuse:  []float64{0, 0, 0, 0},
  Emissive:       []float64{0, 0, 0, 1},
  Specular:       []float64{0.75, 0.75, 0.75, 1},
}

m.Stages = []rvmat.Stage{
  {
    Name:     "Stage1",
    Texture:  rvmat.ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`),
    UVSource: "tex",
  },
}

diagnostics := rvmat.Validate(m, nil)
out, err := rvmat.Format(m, nil)
```

### Read And Write

Read from file:

```go
m, err := rvmat.DecodeFile("material.rvmat", nil)
```

Read from bytes or stream:

```go
m, err := rvmat.Parse(data, nil) // []byte

// or
f, _ := os.Open("material.rvmat")
defer f.Close()
m, err := rvmat.Decode(f, nil)
```

Write to file:

```go
err := rvmat.EncodeFile("out.rvmat", m, nil)
```

Write to bytes:

```go
out, err := rvmat.Format(m, nil)
```

#### Parse Options

Defaults are already tuned for real-world files.

```go
opt := &rvmat.ParseOptions{
  DisableCaseInsensitive: false,
  DisableComments:        false,
  DisableRelaxedNumbers:  false,
}

m, err := rvmat.DecodeFile(path, opt)
```

#### Format Options

```go
fmtOpt := &rvmat.FormatOptions{
  Indent:        "    ", // tabs or spaces
  CompactStages: true,   // one-line StageN blocks for texture+texGen
}

err := rvmat.EncodeFile("out.rvmat", m, fmtOpt)
```

### Parse And Procedural Textures

Parse a procedural texture:

```go
tex := rvmat.ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`)
if tex.IsProcedural() && tex.ParsedOK {
  _ = tex.Procedural
}
```

Create a procedural texture:

```go
tex := rvmat.NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1.0, "co")
raw := tex.Raw
```

Validate a procedural texture:

```go
diagnostics := tex.Validate(&rvmat.TextureValidateOptions{
  DisableProceduralFnCheck:   false,
  DisableProceduralArgsCheck: false,
  DisableTextureTagCheck:     false,
})
```

### Data Helpers

* `TextureRef` represents either a file path or a procedural texture string.
* `ParseTextureRef` and `NewProcedural*` help create procedural references.
* `PathResolver` resolves texture paths against `GameRoot`.

### Full Material Example

```go
m := &rvmat.Material{
  PixelShaderID:  "Super",
  VertexShaderID: "Super",
  Ambient:        []float64{1, 1, 1, 1},
  Diffuse:        []float64{1, 1, 1, 1},
  ForcedDiffuse:  []float64{0, 0, 0, 0},
  Emissive:       []float64{0, 0, 0, 1},
  Specular:       []float64{0.75, 0.75, 0.75, 1},
  Stages: []rvmat.Stage{
    {
      Name:     "Stage1",
      Texture:  rvmat.ParseTextureRef(`dz\gear\cooking\data\cauldron_nohq.paa`),
      UVSource: "tex",
      UVTransform: &rvmat.UVTransform{
        Aside: []float64{1, 0, 0},
        Up:    []float64{0, 1, 0},
        Dir:   []float64{0, 0, 0},
        Pos:   []float64{0, 0, 0},
      },
    },
  },
}
```

### TexGen Resolution

Use these helpers to get effective UV data from a stage, including inherited
`TexGen` chains:

```go
st := m.Stages[0]

resolved, err := rvmat.ResolveStageTexGen(m, st)
uvSource, err := rvmat.EffectiveUVSource(m, st)
uvTransform, err := rvmat.EffectiveUVTransform(m, st)

_, _, _ = resolved, uvSource, uvTransform
```

### Normalize

`Normalize` applies safe, package-level normalization to an in-memory
material:

```go
result, diagnostics := rvmat.Normalize(m, &rvmat.NormalizeOptions{
  StageTextures: true,
  StageOrder:    true,
  TexGenOrder:   true,
  TexturePaths:  true,
})

_, _ = result, diagnostics
```

### Validate

Use `Validate` to run configurable checks for paths, shader names, and
extensions:

```go
valOpt := &rvmat.ValidateOptions{
  TexturePathMode:        rvmat.TexturePathModeTrust,
  DisableExtensionsCheck: false,
  GameRoot:               "P:\\",
  DisableShaderNameCheck: true,
  ExcludePaths:           []string{`dz\vehicles\*`},
}

diagnostics := rvmat.Validate(m, valOpt)
```

Catalog API:

```go
all := rvmat.DiagnosticCatalog()
spec, ok := rvmat.DiagnosticByCode(rvmat.CodeValidateTextureFileNotFound)
_, _, _ = all, spec, ok
```

### lintkit Integration

Lint diagnostics docs:

* machine-readable snapshot: [rules.yaml](rules.yaml)
* human-readable table: [RULES.md](RULES.md)

### Generate

Use `Generate` for low-level generation and `GenerateSet`
for high-level generation with output orchestration.

Available base material profiles are:
`textile`, `steel`, `rust`, `wood`, `glass`, `plastic`, `rubber`,
`leather`, `earth`, `paper`, `concrete`, `stone`, `skin`.

#### Modifier Effects

* `Finish` changes reflectivity and highlight sharpness:
  * `FinishMatte` reduces specular and `specularPower`,
  * `FinishDefault`/`FinishSatin` keep baseline,
  * `FinishGloss` increases both,
  * `FinishPolished` increases both the most (with power clamp).
* `Condition` simulates surface wear:
  * `ConditionDefault`/`ConditionClean` keep baseline,
  * `ConditionWorn` reduces specular and power slightly,
  * `ConditionDirty` reduces them more,
  * `ConditionOxidized` reduces them for aged/oxidized look.

#### Low-Level Example

```go
mat, err := rvmat.Generate(rvmat.GenerateOptions{
  BaseMaterial: rvmat.BaseMaterialSteel,
  Finish:       rvmat.FinishGloss,
  Condition:    rvmat.ConditionWorn,
  UseTexGen:    true,
})
if err != nil {
  // handle
}

damage, _ := rvmat.GenerateDamage(mat)
destruct, _ := rvmat.GenerateDestruct(mat)
_, _ = damage, destruct
```

#### High-Level Example

```go
result, err := rvmat.GenerateSet(rvmat.GenerateSetOptions{
  OutputPath:   `assets\data\box`,
  BaseMaterial: rvmat.BaseMaterialWood,
  Finish:       rvmat.FinishMatte,
  Condition:    rvmat.ConditionDirty,
  BaseTexture:  `assets\data\box_co.paa`,
})
if err != nil {
  // handle
}

err = rvmat.WriteGenerateSet(result, &rvmat.FormatOptions{Indent: "\t"})
if err != nil {
  // handle
}
```

#### Generation Notes

* `GenerateSet` creates main, damage, and destruct by default.
* Disable variants with `DisableDamage` and `DisableDestruct`.
* Override Stage3 macros with `DamageMacroTexture` and `DestructMacroTexture`.
* Texture overrides accept both stage and role keys (`stage1`, `nohq`, etc.);
  stage keys have priority when both target the same stage.
* Use `StageIndexForTextureRole` to resolve role key to stage index.
* Texture auto-fill prefers existing sibling files
  by role suffix and extension priority (`.paa`, `.pax`, `.tga`, `.png`).

### Behavior And Edge Cases

* **Binary rvmat**: returns `ErrBinaryRVMAT`.
  For binarize/debinarize workflows, see <https://github.com/WoozyMasta/rap>.
* **Relaxed numbers**:
  numeric arrays may contain strings/expressions; invalid entries become `0`.
* **Stage with texGen**:
  writer omits `uvSource` and `uvTransform` when `TexGen` is set.
* **Unknown fields**: preserved internally and round-tripped.
