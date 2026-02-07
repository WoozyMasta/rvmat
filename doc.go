/*
Package rvmat provides parsing, writing, and validation for Real Virtuality RVMAT
material files.

It focuses on fast parsing and deterministic formatting, extracting common
fields (ambient/diffuse/specular/etc.), stages, and texgens. Procedural textures
are supported via TextureRef and ProceduralTexture.

Reader example:

	m, err := rvmat.DecodeFile("material.rvmat", nil)
	if err != nil {
		// handle error
	}

Writer example:

	out, err := rvmat.Format(m, nil)
	if err != nil {
		// handle error
	}

Validator example:

	issues := rvmat.Validate(m, nil)
	if len(issues) != 0 {
		// handle validation issues
	}

Texture reader example:

	tex := rvmat.ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`)
	if tex.IsProcedural() && tex.ParsedOK {
		_ = tex.Procedural
	}

Texture writer example:

	tex := rvmat.NewProceduralColor("argb", 8, 8, 3, 0.5, 0.5, 0.5, 1.0, "co")
	_ = tex.Raw

Procedural color validation example:

	tex := rvmat.ParseTextureRef(`#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)`)
	issues := tex.Validate(&rvmat.TextureValidateOptions{
		DisableProceduralFnCheck:   false,
		DisableProceduralArgsCheck: false,
		DisableTextureTagCheck:     false,
	})
	if len(issues) != 0 {
		// handle validation issues
	}
*/
package rvmat
