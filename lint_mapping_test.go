// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import "testing"

func TestValidateIssuesHaveKnownCodes(t *testing.T) {
	t.Parallel()

	material := &Material{
		PixelShaderID:  "UnknownPS",
		VertexShaderID: "UnknownVS",
		Stages: []Stage{
			{
				Name:    "Stage1",
				Texture: ParseTextureRef(`dz\bad\..\tex.jpg`),
			},
			{
				Name:    "Stage1",
				Texture: ParseTextureRef(`#(rgba,0,8,3)unknown(1)`),
			},
		},
	}

	issues := Validate(material, &ValidateOptions{
		TexturePathMode:          TexturePathModeIgnore,
		DisableExtensionsCheck:   false,
		DisableShaderNameCheck:   false,
		EnableShaderProfileCheck: true,
	})
	if len(issues) == 0 {
		t.Fatal("expected non-empty validation issues")
	}

	for index := range issues {
		if issues[index].Code == 0 {
			t.Fatalf("issues[%d] has empty code: %+v", index, issues[index])
		}

		if _, ok := DiagnosticByCode(issues[index].Code); !ok {
			t.Fatalf("issues[%d] has unknown code %q", index, issues[index].Code)
		}
	}
}

func TestNormalizeIssuesHaveKnownCodes(t *testing.T) {
	t.Parallel()

	_, issues := Normalize(nil, nil)
	if len(issues) == 0 {
		t.Fatal("expected normalize issue for nil material")
	}

	for index := range issues {
		if issues[index].Code == 0 {
			t.Fatalf("issues[%d] has empty code: %+v", index, issues[index])
		}

		if _, ok := DiagnosticByCode(issues[index].Code); !ok {
			t.Fatalf("issues[%d] has unknown code %q", index, issues[index].Code)
		}
	}
}
