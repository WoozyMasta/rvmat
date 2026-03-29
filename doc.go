// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

/*
Package rvmat provides parsing, writing, validation, normalization, and
generation for Real Virtuality RVMAT material files.

It also provides native lintkit integration with stable rule codes,
catalog export, and provider-based rule registration.

Reader example:

	m, err := rvmat.DecodeFile("material.rvmat", nil)
	if err != nil {
		// handle error
	}

Writer example:

	out, err := rvmat.Format(m, &rvmat.FormatOptions{Indent: "\t"})
	if err != nil {
		// handle error
	}
	_ = out

Validator example:

	diagnostics := rvmat.Validate(m, &rvmat.ValidateOptions{
		TexturePathMode:          rvmat.TexturePathModeIgnore,
		EnableShaderProfileCheck: true,
	})
	_ = diagnostics

lintkit catalog example:

	all := rvmat.DiagnosticCatalog()
	spec, ok := rvmat.DiagnosticByCode(rvmat.CodeValidateTextureFileNotFound)
	_, _, _ = all, spec, ok

TexGen effective UV example:

	stage := m.Stages[0]
	uvSource, err := rvmat.EffectiveUVSource(m, stage)
	if err != nil {
		// unknown texGen, broken base, or cycle
	}
	_ = uvSource

Normalization example:

	result, normalizeDiagnostics := rvmat.Normalize(m, &rvmat.NormalizeOptions{
		StageTextures: true,
		StageOrder:    true,
		TexGenOrder:   true,
		TexturePaths:  true,
	})
	_ = result
	_ = normalizeDiagnostics

High-level generator example:

	gen, err := rvmat.GenerateSet(rvmat.GenerateSetOptions{
		OutputPath:   `assets\data\testbox`,
		BaseMaterial: rvmat.BaseMaterialSteel,
		Finish:       rvmat.FinishGloss,
		Condition:    rvmat.ConditionWorn,
	})
	if err != nil {
		// handle error
	}

	err = rvmat.WriteGenerateSet(gen, &rvmat.FormatOptions{
		Indent: "\t",
	})
	if err != nil {
		// handle error
	}

Important format note: RVMAT key on disk is "emmisive[]". API field remains "Emissive".
*/
package rvmat
