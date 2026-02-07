# rvmat

Go library for parsing, validating, and writing
Real Virtuality `.rvmat` material files.

* Fast parser with tolerances for real-world data
  (case-insensitive keys, comments, relaxed numeric arrays).
* Deterministic writer (canonical formatting) with configurable indentation.
* Helpers for procedural textures, path resolution, and validation.
* Validator supports configurable checks (shader/stage allowlists are optional).
* Unknown fields are preserved internally and round-tripped.

## Install

```bash
go get github.com/woozymasta/rvmat
```

## Usage

```go
m, err := rvmat.DecodeFile("material.rvmat", nil)
if err != nil {
  // handle
}

// Validate with defaults
issues := rvmat.Validate(m, nil)

// Render canonical output
out, err := rvmat.Format(m, nil)
```

## Practical example

Create a minimal material with a procedural texture:

```go
m := &rvmat.Material{
  PixelShaderID:  "Super",
  VertexShaderID: "Super",
  Ambient:        []float64{1, 1, 1, 1},
  Diffuse:        []float64{1, 1, 1, 1},
  ForcedDiffuse:  []float64{0, 0, 0, 0},
  Emmisive:       []float64{0, 0, 0, 1},
  Specular:       []float64{0.75, 0.75, 0.75, 1},
}

m.Stages = []rvmat.Stage{
  {
    Name:     "Stage1",
    Texture:  rvmat.ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`),
    UVSource: "tex",
  },
}

issues := rvmat.Validate(m, nil)
out, err := rvmat.Format(m, nil)
```

## Read/write

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

### Parse options

Defaults are already tuned for real-world files.

```go
opt := &rvmat.ParseOptions{
  DisableCaseInsensitive: false,
  DisableComments:        false,
  DisableRelaxedNumbers:  false,
}

m, err := rvmat.DecodeFile(path, opt)
```

### Format options

```go
fmtOpt := &rvmat.FormatOptions{
  Indent: "    ", // tabs or spaces
}

err := rvmat.EncodeFile("out.rvmat", m, fmtOpt)
```

### Validate options

```go
valOpt := &rvmat.ValidateOptions{
  DisableFileCheck:       true,
  DisableExtensionsCheck: false,
  GameRoot:               "P:\\",
  DisableShaderNameCheck: true,
  ExcludePaths:           []string{`dz\vehicles\*`},
}

issues := rvmat.Validate(m, valOpt)
```

## Procedural textures

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
issues := tex.Validate(&rvmat.TextureValidateOptions{
  DisableProceduralFnCheck:   false,
  DisableProceduralArgsCheck: false,
  DisableTextureTagCheck:     false,
})
```

## Minimal structure example

```go
m := &rvmat.Material{
  PixelShaderID:  "Super",
  VertexShaderID: "Super",
  Ambient:        []float64{1, 1, 1, 1},
  Diffuse:        []float64{1, 1, 1, 1},
  ForcedDiffuse:  []float64{0, 0, 0, 0},
  Emmisive:       []float64{0, 0, 0, 1},
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

## Behavior and edge cases

* **Binary rvmat**:
  returns `ErrBinaryRVMAT`.
* **Relaxed numbers**:
  numeric arrays may contain strings/expressions; invalid entries become `0`.
* **Stage with texGen**:
  writer omits `uvSource` and `uvTransform` when `TexGen` is set.
* **Unknown fields**:
  preserved internally and round-tripped.

## Data helpers

* `TextureRef` represents either a file path or a procedural texture string.
* `ParseTextureRef` and `NewProcedural*` help create procedural references.
* `PathResolver` resolves texture paths against `GameRoot`.

## References

* <https://community.bistudio.com/wiki/RVMAT_basics>
* <https://community.bistudio.com/wiki/Multimaterial>
* <https://community.bistudio.com/wiki/DayZ:Projection_Layer>
* <https://community.bistudio.com/wiki/Multimaterial>
* <https://community.bistudio.com/wiki/Rvmat_File_Format>
