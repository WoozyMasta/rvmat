// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"errors"

	"github.com/woozymasta/lintkit/lint"
)

var (
	// ErrBinaryRVMAT indicates the file is not a text RVMAT (likely binary surface data).
	ErrBinaryRVMAT = errors.New("binary rvmat")

	// ErrLex indicates a lexer failure.
	ErrLex = errors.New("lex error")

	// ErrParse indicates a parser failure.
	ErrParse = errors.New("parse error")

	// ErrTexGenNotFound indicates stage texGen reference points to unknown TexGen.
	ErrTexGenNotFound = errors.New("texgen not found")

	// ErrTexGenBaseNotFound indicates TexGen inheritance points to unknown base.
	ErrTexGenBaseNotFound = errors.New("texgen base not found")

	// ErrTexGenCycle indicates a cycle in TexGen inheritance chain.
	ErrTexGenCycle = errors.New("texgen inheritance cycle")

	// ErrUnknownBaseMaterial indicates unknown material generator base material.
	ErrUnknownBaseMaterial = errors.New("unknown base material")

	// ErrInvalidGenerateOption indicates invalid value in GenerateOptions.
	ErrInvalidGenerateOption = errors.New("invalid generate option")

	// ErrMaterialNotFound indicates required material input is missing.
	ErrMaterialNotFound = errors.New("material not found")

	// ErrStageNotFound indicates required stage is missing.
	ErrStageNotFound = errors.New("stage not found")

	// ErrNilLintRuleRegistrar indicates nil lint rule registrar in registration.
	ErrNilLintRuleRegistrar = lint.ErrNilRuleRegistrar
)
