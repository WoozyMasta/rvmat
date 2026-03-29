// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"github.com/woozymasta/lintkit/lint"
)

const (
	// LintModule is stable lint module namespace for rvmat rules.
	LintModule = "rvmat"
)

const (
	// StageValidate marks validation diagnostics.
	StageValidate lint.Stage = "validate"

	// StageNormalize marks normalization diagnostics.
	StageNormalize lint.Stage = "normalize"
)

const (
	// CodeNormalizeNilMaterial reports nil material in normalize flow.
	CodeNormalizeNilMaterial lint.Code = 1001
)

const (
	// CodeValidatePixelShaderMissing reports missing pixel shader id.
	CodeValidatePixelShaderMissing lint.Code = 2001

	// CodeValidateVertexShaderMissing reports missing vertex shader id.
	CodeValidateVertexShaderMissing lint.Code = 2002

	// CodeValidateUnknownPixelShaderID reports unknown pixel shader id value.
	CodeValidateUnknownPixelShaderID lint.Code = 2003

	// CodeValidateUnknownVertexShaderID reports unknown vertex shader id value.
	CodeValidateUnknownVertexShaderID lint.Code = 2004

	// CodeValidateShaderProfileMissingRequiredStage reports required profile stage miss.
	CodeValidateShaderProfileMissingRequiredStage lint.Code = 2005

	// CodeValidateShaderProfileMissingCommonStage reports common profile stage miss.
	CodeValidateShaderProfileMissingCommonStage lint.Code = 2006

	// CodeValidateShaderProfileStageSetMismatch reports strict stage set mismatch.
	CodeValidateShaderProfileStageSetMismatch lint.Code = 2007

	// CodeValidateUnexpectedTextureExtension reports unexpected texture extension.
	CodeValidateUnexpectedTextureExtension lint.Code = 2008

	// CodeValidateTexturePathParentTraversal reports parent-traversal path segments.
	CodeValidateTexturePathParentTraversal lint.Code = 2009

	// CodeValidateTextureFileNotFound reports unresolved texture file path.
	CodeValidateTextureFileNotFound lint.Code = 2010

	// CodeValidateUnknownStageName reports unknown stage name.
	CodeValidateUnknownStageName lint.Code = 2011

	// CodeValidateStageUnknownTexGen reports stage texgen reference miss.
	CodeValidateStageUnknownTexGen lint.Code = 2012

	// CodeValidateTexGenBaseNotFound reports broken texgen inheritance base.
	CodeValidateTexGenBaseNotFound lint.Code = 2013

	// CodeValidateTexGenCycle reports texgen inheritance cycle.
	CodeValidateTexGenCycle lint.Code = 2014

	// CodeValidateTexGenResolutionFailed reports unresolved texgen chain.
	CodeValidateTexGenResolutionFailed lint.Code = 2015

	// CodeValidateStageMissingEffectiveUVSource reports missing effective uvSource.
	CodeValidateStageMissingEffectiveUVSource lint.Code = 2016

	// CodeValidateStageMissingEffectiveUVTransform reports missing effective uvTransform.
	CodeValidateStageMissingEffectiveUVTransform lint.Code = 2017

	// CodeValidateDuplicateStageName reports duplicate stage names.
	CodeValidateDuplicateStageName lint.Code = 2018

	// CodeValidateColorComponentCount reports invalid color vector component count.
	CodeValidateColorComponentCount lint.Code = 2019

	// CodeValidateUVTransformVectorRequired reports missing uvTransform vector.
	CodeValidateUVTransformVectorRequired lint.Code = 2020

	// CodeValidateUVTransformVectorComponentCount reports bad uvTransform vector width.
	CodeValidateUVTransformVectorComponentCount lint.Code = 2021

	// CodeValidateProceduralTextureParseFailed reports failed procedural parse.
	CodeValidateProceduralTextureParseFailed lint.Code = 2022

	// CodeValidateUnknownProceduralFunction reports unknown procedural texture function.
	CodeValidateUnknownProceduralFunction lint.Code = 2023

	// CodeValidateUnknownProceduralTextureFormat reports unknown procedural format.
	CodeValidateUnknownProceduralTextureFormat lint.Code = 2024

	// CodeValidateInvalidProceduralTextureHeaderDimensions reports invalid dimensions.
	CodeValidateInvalidProceduralTextureHeaderDimensions lint.Code = 2025

	// CodeValidateUnexpectedProceduralArgumentCount reports invalid arg count.
	CodeValidateUnexpectedProceduralArgumentCount lint.Code = 2026

	// CodeValidateInvalidProceduralNumericArguments reports invalid numeric args.
	CodeValidateInvalidProceduralNumericArguments lint.Code = 2027

	// CodeValidateUnknownTextureTag reports unknown texture tag.
	CodeValidateUnknownTextureTag lint.Code = 2028
)

var diagnosticCodeCatalogConfig = lint.CodeCatalogConfig{
	Module:            LintModule,
	CodePrefix:        "RVMAT",
	ModuleName:        "Real Virtuality Materials",
	ModuleDescription: "Lint rules for .rvmat normalization and validation flows.",
	ScopeDescriptions: map[lint.Stage]string{
		StageValidate:  "Semantic validation diagnostics.",
		StageNormalize: "Normalization diagnostics.",
	},
}

var diagnosticCodeCatalogHandle = lint.NewCodeCatalogHandle(
	diagnosticCodeCatalogConfig,
	diagnosticCatalog,
)

// getDiagnosticCodeCatalog returns lazy-initialized code catalog helper.
func getDiagnosticCodeCatalog() (lint.CodeCatalog, error) {
	return diagnosticCodeCatalogHandle.Catalog()
}

// DiagnosticRuleSpec converts one diagnostic spec into lint rule metadata.
func DiagnosticRuleSpec(spec lint.CodeSpec) (lint.RuleSpec, error) {
	return diagnosticCodeCatalogHandle.RuleSpec(spec)
}

// LintRuleID returns lint rule ID mapped from stable rvmat diagnostic code.
func LintRuleID(code lint.Code) string {
	return diagnosticCodeCatalogHandle.RuleIDOrUnknown(code)
}

// DiagnosticCatalog returns stable diagnostics metadata list.
func DiagnosticCatalog() []lint.CodeSpec {
	return diagnosticCodeCatalogHandle.CodeSpecs()
}

// DiagnosticByCode returns diagnostic metadata for code.
func DiagnosticByCode(code lint.Code) (lint.CodeSpec, bool) {
	return diagnosticCodeCatalogHandle.ByCode(code)
}

// LintRuleSpecs returns deterministic lint rule specs from diagnostics catalog.
func LintRuleSpecs() []lint.RuleSpec {
	return diagnosticCodeCatalogHandle.RuleSpecs()
}
