# rvmat file foramt

This document describes the `.rvmat` format used by Real Virtuality.
It is format documentation, not package documentation.

An `.rvmat` file defines how a surface is rendered. It binds shader IDs,
material coefficients, and texture stages into one text config file.

## Core idea

In Real Virtuality, a material is built from stages. A stage is one shader
input layer. Different shaders expect different stage layouts, so the meaning
of `Stage1`, `Stage2`, and so on depends on shader type.

Shader selection is done with two top-level keys:

* `PixelShaderID`
* `VertexShaderID`

## File anatomy

A typical `.rvmat` contains top-level material properties, then class blocks
for stages, and optionally class blocks for TexGen UV reuse.

Most color-like arrays use 4 floats in `RGBA` order:

* index `0` = `R` (red)
* index `1` = `G` (green)
* index `2` = `B` (blue)
* index `3` = `A` (alpha)

In practical authoring, values are typically in the `0..1` range.

```cpp
ambient[] = {1,1,1,1};
diffuse[] = {1,1,1,1};
forcedDiffuse[] = {0,0,0,0};
emmisive[] = {0,0,0,1};
specular[] = {0.2,0.2,0.2,1};
specularPower = 80;
PixelShaderID = "Super";
VertexShaderID = "Super";
```

The key spelling `emmisive[]` is intentional and canonical for `.rvmat`.

## Stage and UV mapping

A stage usually defines a texture and UV source. UV can be written directly
inside the stage, or inherited through a `texGen` reference.

`TexGen` is a reusable UV block. It can also inherit from another TexGen.
That is how large files avoid repeating identical `uvTransform` blocks.

```cpp
class TexGen0
{
    uvSource = "tex";
    class uvTransform
    {
        aside[] = {1,0,0};
        up[] = {0,1,0};
        dir[] = {0,0,1};
        pos[] = {0,0,0};
    };
};

class Stage1
{
    texture = "mymod\data\item_nohq.paa";
    texGen = 0;
};
```

Common `uvSource` values in game data are `tex`, `tex1`, `none`, and
`WorldPos`.

## Super shader layout

`Super` is the most common general-purpose layout. In practice it is used with
`PixelShaderID="Super"` and `VertexShaderID="Super"` and this stage meaning:

1. `Stage1`: normal map input, usually `_NOHQ`
1. `Stage2`: detail map input, usually `_DT` or `_CDT`
1. `Stage3`: macro map input, usually `_MC`
1. `Stage4`: ambient shadow input, usually `_AS` / `_ADS` / `_ADSHQ`
1. `Stage5`: SMDI input (`_SMDI`, sometimes `_SM`/`_DTSMDI`)
1. `Stage6`: fresnel input, often procedural
1. `Stage7`: environment map input
1. `StageTI`: optional thermal stage

In many assets, the base color map (`_CO`/`_CA`) comes from the model texture
assignment and not from a dedicated Super stage.

## SMDI and AS details

SMDI should be treated as an RV-specific packing, not as direct metallic/PBR
data. The common target convention is "specular in green, gloss in blue".
In channel terms that means:

* `R`: diffuse inverse, usually near white
* `G`: specular intensity
* `B`: gloss or highlight sharpness

Avoid pure black values in gloss (`B`) for production assets, because black
gloss pixels can produce broken-looking highlights.

For `_AS` maps, the green channel is the important AO signal channel in common
workflows.

If source authoring is metallic/roughness, roughness can be converted to gloss
with inversion and curve remap. Metallic cannot be copied directly to SMDI
specular without calibration.

## Procedural textures

Procedural textures use this syntax:

```text
#(format,width,height,mip)function(arg1,arg2,...)
```

Typical examples:

* `#(argb,8,8,3)color(0.5,0.5,1,1,NOHQ)`
* `#(argb,8,8,3)color(0,0,0,0,MC)`
* `#(ai,64,1,1)fresnel(1.3,0.2)`
* `#(ai,64,1,1)fresnelGlass(1.7)`

For fresnel LUT usage, height `1` is common and efficient.

## Texture suffix conventions

Frequently used suffixes:

* `_CO`, `_CA` for base color
* `_NOHQ`, `_NO`, `_NS` for normal family
* `_DT`, `_CDT` for detail family
* `_MC` for macro
* `_AS`, `_ADS`, `_ADSHQ` for ambient-shadow family
* `_SMDI`, `_SM`, `_DTSMDI` for specular family
* `_TI` for thermal

For a broader practical suffix/alias set used in tooling, see
<https://github.com/WoozyMasta/paa/blob/master/texconfig/default_values.go>

## Other shader families

`Glass` commonly uses fresnel and environment reflection stages, and alpha has
a strong impact on transparency versus reflection balance.

`Multi` can use a broad stage layout (`Stage0..Stage14`) and multiple TexGen
slots. UV behavior is more shader-specific there than in simple Super files.

## Path style

Final game-facing texture paths are typically written with backslashes and
without drive letters.

Preferred style:

```text
mymod\data\item_nohq.paa
```

## Minimal complete Super example

```cpp
ambient[] = {1,1,1,1};
diffuse[] = {1,1,1,1};
forcedDiffuse[] = {0,0,0,0};
emmisive[] = {0,0,0,1};
specular[] = {0.2,0.2,0.2,1};
specularPower = 80;
PixelShaderID = "Super";
VertexShaderID = "Super";

class TexGen0
{
    uvSource = "tex";
    class uvTransform
    {
        aside[] = {1,0,0};
        up[] = {0,1,0};
        dir[] = {0,0,1};
        pos[] = {0,0,0};
    };
};

class Stage1 { texture = "mymod\data\item_nohq.paa"; texGen = 0; };
class Stage2 { texture = "#(argb,8,8,3)color(0.5,0.5,0.5,1,DT)"; texGen = 0; };
class Stage3 { texture = "#(argb,8,8,3)color(0,0,0,0,MC)"; texGen = 0; };
class Stage4 { texture = "mymod\data\item_as.paa"; texGen = 0; };
class Stage5 { texture = "mymod\data\item_smdi.paa"; texGen = 0; };
class Stage6 { texture = "#(ai,64,1,1)fresnel(1.3,0.2)"; texGen = 0; };
class Stage7 { texture = "dz\data\data\env_land_co.paa"; texGen = 0; };
```

## Useful links

* <https://community.bistudio.com/wiki/Rvmat_File_Format>
* <https://community.bistudio.com/wiki/RVMAT_basics>
* <https://community.bistudio.com/wiki/Material_Templates>
* <https://community.bistudio.com/wiki/Multimaterial>
* <https://community.bistudio.com/wiki/Super_shader>
* <https://community.bistudio.com/wiki/Skin_shader>
