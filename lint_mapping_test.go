// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

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
		if issues[index].Code == "" {
			t.Fatalf("issues[%d] has empty code: %+v", index, issues[index])
		}

		code, ok := lint.ParsePublicCode(issues[index].Code)
		if !ok {
			t.Fatalf("issues[%d] has invalid code %q", index, issues[index].Code)
		}

		if _, ok = DiagnosticByCode(code); !ok {
			t.Fatalf("issues[%d] has unknown code %q", index, issues[index].Code)
		}
	}
}

func TestNormalizeIssuesHaveKnownCodes(t *testing.T) {
	t.Parallel()

	_, issues := Normalize(nil, nil)
	if len(issues) == 0 {
		t.Fatal("expected normalize lint.Diagnostic for nil material")
	}

	for index := range issues {
		if issues[index].Code == "" {
			t.Fatalf("issues[%d] has empty code: %+v", index, issues[index])
		}

		code, ok := lint.ParsePublicCode(issues[index].Code)
		if !ok {
			t.Fatalf("issues[%d] has invalid code %q", index, issues[index].Code)
		}

		if _, ok = DiagnosticByCode(code); !ok {
			t.Fatalf("issues[%d] has unknown code %q", index, issues[index].Code)
		}
	}
}
