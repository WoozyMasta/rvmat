<!-- Automatically generated file, do not modify! -->

# Lint Rules Registry

This document contains the current registry of lint rules.

Total rules: 29.

## rvmat

Real Virtuality Materials

> Lint rules for .rvmat normalization and validation flows.

Rule groups for `rvmat`:

* [normalize](#normalize)
* [validate](#validate)

### normalize

> Normalization diagnostics.

#### `RVMAT1001`

Normalization failed: material input is nil

> Ensure parsed material object is created before normalization or validation.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.normalize.normalization-failed-material-input-is-nil` |
| Scope | `normalize` |
| Severity | `error` |
| Enabled | `true` (implicit) |

### validate

> Semantic validation diagnostics.

Codes:
[RVMAT2001](#rvmat2001),
[RVMAT2002](#rvmat2002),
[RVMAT2003](#rvmat2003),
[RVMAT2004](#rvmat2004),
[RVMAT2005](#rvmat2005),
[RVMAT2006](#rvmat2006),
[RVMAT2007](#rvmat2007),
[RVMAT2008](#rvmat2008),
[RVMAT2009](#rvmat2009),
[RVMAT2010](#rvmat2010),
[RVMAT2011](#rvmat2011),
[RVMAT2012](#rvmat2012),
[RVMAT2013](#rvmat2013),
[RVMAT2014](#rvmat2014),
[RVMAT2015](#rvmat2015),
[RVMAT2016](#rvmat2016),
[RVMAT2017](#rvmat2017),
[RVMAT2018](#rvmat2018),
[RVMAT2019](#rvmat2019),
[RVMAT2020](#rvmat2020),
[RVMAT2021](#rvmat2021),
[RVMAT2022](#rvmat2022),
[RVMAT2023](#rvmat2023),
[RVMAT2024](#rvmat2024),
[RVMAT2025](#rvmat2025),
[RVMAT2026](#rvmat2026),
[RVMAT2027](#rvmat2027),
[RVMAT2028](#rvmat2028),

#### `RVMAT2001`

Pixel shader ID is missing

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.pixel-shader-id-is-missing` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2002`

Vertex shader ID is missing

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.vertex-shader-id-is-missing` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2003`

Pixel shader ID is unknown

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.pixel-shader-id-is-unknown` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2004`

Vertex shader ID is unknown

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.vertex-shader-id-is-unknown` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2005`

Shader profile is missing required stage

> Selected profile expects one or more required stages that are not present in
> material.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.shader-profile-is-missing-required-stage` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2006`

Shader profile is missing common stage

> Selected profile usually includes this stage. Missing it can be valid, but
> often indicates incomplete material setup.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.shader-profile-is-missing-common-stage` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2007`

Shader profile stage set mismatch

> Defined stage set does not match expected layout for selected profile.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.shader-profile-stage-set-mismatch` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2008`

Unexpected texture extension

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.unexpected-texture-extension` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

Default options:
```json
{
  "allowed_extensions": [
    ".paa",
    ".pax",
    ".tga",
    ".png"
  ]
}
```

#### `RVMAT2009`

Texture path contains parent traversal (`..`)

> This may break packing rules and cross-platform path normalization.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.texture-path-contains-parent-traversal` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2010`

Texture file not found

> Verify file exists and path mapping is correct for current search roots.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.texture-file-not-found` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2011`

Stage name is unknown

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.stage-name-is-unknown` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2012`

Stage references unknown texGen entry

> Define referenced `texGen` entry or fix `texGen` reference in `stage`.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.stage-references-unknown-texgen-entry` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2013`

TexGen inheritance base entry was not found

> Fix base name or define missing parent `texGen` entry.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.texgen-inheritance-base-entry-was-not-found` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2014`

TexGen inheritance cycle detected

> Remove recursive parent chain so effective `texGen` values can be resolved.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.texgen-inheritance-cycle-detected` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2015`

TexGen resolution failed

> Usually caused by invalid inheritance graph or malformed `texGen` data.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.texgen-resolution-failed` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2016`

Stage has no effective uvSource

> Define `uvSource` directly or via `texGen` chain.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.stage-has-no-effective-uvsource` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2017`

Stage has no effective uvTransform

> Provide transform values directly or via `texGen` chain.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.stage-has-no-effective-uvtransform` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2018`

Duplicate stage name

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.duplicate-stage-name` |
| Scope | `validate` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `RVMAT2019`

Color vector must have 4 components

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.color-vector-must-have-4-components` |
| Scope | `validate` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `RVMAT2020`

UvTransform vector is required

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.uvtransform-vector-is-required` |
| Scope | `validate` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `RVMAT2021`

UvTransform vector must have 3 components

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.uvtransform-vector-must-have-3-components` |
| Scope | `validate` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `RVMAT2022`

Failed to parse procedural texture header

> Verify syntax and argument format of procedural texture header.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.failed-to-parse-procedural-texture-header` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2023`

Unknown procedural function

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.unknown-procedural-function` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2024`

Unknown procedural texture format

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.unknown-procedural-texture-format` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2025`

Invalid procedural texture header dimensions

> Width and height must be valid positive values.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.invalid-procedural-texture-header-dimensions` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2026`

Procedural argument count is unexpected

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.procedural-argument-count-is-unexpected` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2027`

Procedural numeric argument values are invalid

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.procedural-numeric-argument-values-are-invalid` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `RVMAT2028`

Unknown texture tag

> Use known engine tag prefix or absolute project-relative path.

| Field | Value |
| --- | --- |
| Rule ID | `rvmat.validate.unknown-texture-tag` |
| Scope | `validate` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

Default options:
```json
{
  "allowed_tags": [
    "ads",
    "adshq",
    "as",
    "ca",
    "cat",
    "cdt",
    "co",
    "draftlco",
    "dt",
    "dtsmdi",
    "gs",
    "lca",
    "lco",
    "mask",
    "mc",
    "mca",
    "mco",
    "no",
    "noex",
    "nof",
    "nofex",
    "nofhq",
    "nohq",
    "non",
    "nopx",
    "normalmap",
    "novhq",
    "nsex",
    "nshq",
    "pr",
    "raw",
    "sky",
    "sm",
    "smdi"
  ]
}
```

---

> Generated with
> [lintkit](https://github.com/woozymasta/lintkit)
> version `dev`
> commit `unknown`

<!-- Automatically generated file, do not modify! -->
