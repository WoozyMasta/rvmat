// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import "github.com/woozymasta/lintkit/lint"

// withDescription attaches optional documentation text to one catalog spec.
func withDescription(spec lint.CodeSpec, description string) lint.CodeSpec {
	spec.Description = description
	return spec
}

// diagnosticCatalog stores stable diagnostics metadata table.
var diagnosticCatalog = []lint.CodeSpec{
	withDescription(
		lint.ErrorCodeSpec(
			CodeNormalizeNilMaterial,
			StageNormalize,
			"normalization failed: material input is nil",
		),
		"Ensure parsed material object is created before normalization or "+
			"validation.",
	),
	lint.WarningCodeSpec(
		CodeValidatePixelShaderMissing,
		StageValidate,
		"pixel shader ID is missing",
	),
	lint.WarningCodeSpec(
		CodeValidateVertexShaderMissing,
		StageValidate,
		"vertex shader ID is missing",
	),
	lint.WarningCodeSpec(
		CodeValidateUnknownPixelShaderID,
		StageValidate,
		"pixel shader ID is unknown",
	),
	lint.WarningCodeSpec(
		CodeValidateUnknownVertexShaderID,
		StageValidate,
		"vertex shader ID is unknown",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateShaderProfileMissingRequiredStage,
			StageValidate,
			"shader profile is missing required stage",
		),
		"Selected profile expects one or more required stages that are not "+
			"present in material.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateShaderProfileMissingCommonStage,
			StageValidate,
			"shader profile is missing common stage",
		),
		"Selected profile usually includes this stage. Missing it can be valid, "+
			"but often indicates incomplete material setup.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateShaderProfileStageSetMismatch,
			StageValidate,
			"shader profile stage set mismatch",
		),
		"Defined stage set does not match expected layout for selected profile.",
	),
	lint.WithCodeOptions(
		lint.WarningCodeSpec(
			CodeValidateUnexpectedTextureExtension,
			StageValidate,
			"unexpected texture extension",
		),
		UnexpectedTextureExtensionRuleOptions{
			AllowedExtensions: defaultTextureExtensions,
		},
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateTexturePathParentTraversal,
			StageValidate,
			"texture path contains parent traversal (`..`)",
		),
		"This may break packing rules and cross-platform path normalization.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateTextureFileNotFound,
			StageValidate,
			"texture file not found",
		),
		"Verify file exists and path mapping is correct for current search roots.",
	),
	lint.WarningCodeSpec(
		CodeValidateUnknownStageName,
		StageValidate,
		"stage name is unknown",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateStageUnknownTexGen,
			StageValidate,
			"stage references unknown texGen entry",
		),
		"Define referenced `texGen` entry or fix `texGen` reference in `stage`.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateTexGenBaseNotFound,
			StageValidate,
			"texGen inheritance base entry was not found",
		),
		"Fix base name or define missing parent `texGen` entry.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateTexGenCycle,
			StageValidate,
			"texGen inheritance cycle detected",
		),
		"Remove recursive parent chain so effective `texGen` values can be "+
			"resolved.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateTexGenResolutionFailed,
			StageValidate,
			"texGen resolution failed",
		),
		"Usually caused by invalid inheritance graph or malformed `texGen` "+
			"data.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateStageMissingEffectiveUVSource,
			StageValidate,
			"stage has no effective uvSource",
		),
		"Define `uvSource` directly or via `texGen` chain.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateStageMissingEffectiveUVTransform,
			StageValidate,
			"stage has no effective uvTransform",
		),
		"Provide transform values directly or via `texGen` chain.",
	),
	lint.ErrorCodeSpec(
		CodeValidateDuplicateStageName,
		StageValidate,
		"duplicate stage name",
	),
	lint.ErrorCodeSpec(
		CodeValidateColorComponentCount,
		StageValidate,
		"color vector must have 4 components",
	),
	lint.ErrorCodeSpec(
		CodeValidateUVTransformVectorRequired,
		StageValidate,
		"uvTransform vector is required",
	),
	lint.ErrorCodeSpec(
		CodeValidateUVTransformVectorComponentCount,
		StageValidate,
		"uvTransform vector must have 3 components",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateProceduralTextureParseFailed,
			StageValidate,
			"failed to parse procedural texture header",
		),
		"Verify syntax and argument format of procedural texture header.",
	),
	lint.WarningCodeSpec(
		CodeValidateUnknownProceduralFunction,
		StageValidate,
		"unknown procedural function",
	),
	lint.WarningCodeSpec(
		CodeValidateUnknownProceduralTextureFormat,
		StageValidate,
		"unknown procedural texture format",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeValidateInvalidProceduralTextureHeaderDimensions,
			StageValidate,
			"invalid procedural texture header dimensions",
		),
		"Width and height must be valid positive values.",
	),
	lint.WarningCodeSpec(
		CodeValidateUnexpectedProceduralArgumentCount,
		StageValidate,
		"procedural argument count is unexpected",
	),
	lint.WarningCodeSpec(
		CodeValidateInvalidProceduralNumericArguments,
		StageValidate,
		"procedural numeric argument values are invalid",
	),
	withDescription(
		lint.WithCodeOptions(
			lint.WarningCodeSpec(
				CodeValidateUnknownTextureTag,
				StageValidate,
				"unknown texture tag",
			),
			UnknownTextureTagRuleOptions{
				AllowedTags: orderedKnownTextureTags(),
			},
		),
		"Use known engine tag prefix or absolute project-relative path.",
	),
}
