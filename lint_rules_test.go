// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"context"
	"testing"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
	"github.com/woozymasta/lintkit/linttest"
)

// lintRulesTestRegistrar captures registered runners for tests.
type lintRulesTestRegistrar struct {
	// runners stores registered runner instances.
	runners []lint.RuleRunner
}

// Register appends all provided runners into local test storage.
func (registrar *lintRulesTestRegistrar) Register(
	runners ...lint.RuleRunner,
) error {
	registrar.runners = append(registrar.runners, runners...)
	return nil
}

func TestLintRuleSpecsMatchCatalog(t *testing.T) {
	t.Parallel()

	linttest.AssertCatalogContract(
		t,
		LintModule,
		DiagnosticCatalog(),
		LintRuleSpecs(),
		LintRuleID,
	)
}

func TestRegisterLintRules(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	if err := RegisterLintRules(&registrar); err != nil {
		t.Fatalf("RegisterLintRules() error: %v", err)
	}

	catalog := DiagnosticCatalog()
	if len(registrar.runners) != len(catalog) {
		t.Fatalf("registered runners=%d, want %d", len(registrar.runners), len(catalog))
	}
}

func TestLintRulesProviderRegisterRules(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	provider := LintRulesProvider{}
	if err := provider.RegisterRules(&registrar); err != nil {
		t.Fatalf("LintRulesProvider.RegisterRules() error: %v", err)
	}

	catalog := DiagnosticCatalog()
	if len(registrar.runners) != len(catalog) {
		t.Fatalf("registered runners=%d, want %d", len(registrar.runners), len(catalog))
	}
}

func TestLintRulesProviderRegisterRulesByScope(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	provider := LintRulesProvider{}
	if err := provider.RegisterRulesByScope(&registrar, string(StageValidate)); err != nil {
		t.Fatalf("LintRulesProvider.RegisterRulesByScope() error: %v", err)
	}

	catalog := DiagnosticCatalog()
	if len(registrar.runners) != len(catalog)-1 {
		t.Fatalf("registered runners=%d, want %d", len(registrar.runners), len(catalog)-1)
	}
}

func TestLintRulesProviderRegisterRulesByStage(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	provider := LintRulesProvider{}
	if err := provider.RegisterRulesByStage(&registrar, StageNormalize); err != nil {
		t.Fatalf("LintRulesProvider.RegisterRulesByStage() error: %v", err)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}
}

func TestRegisterLintRulesNilRegistrar(t *testing.T) {
	t.Parallel()

	if err := RegisterLintRules(nil); err != ErrNilLintRuleRegistrar {
		t.Fatalf(
			"RegisterLintRules(nil) error=%v, want %v",
			err,
			ErrNilLintRuleRegistrar,
		)
	}
}

func TestRegisterLintRulesByScope(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	if err := RegisterLintRulesByScope(&registrar, string(StageNormalize)); err != nil {
		t.Fatalf("RegisterLintRulesByScope() error: %v", err)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}
}

func TestRegisterLintRulesByStage(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	if err := RegisterLintRulesByStage(&registrar, StageValidate); err != nil {
		t.Fatalf("RegisterLintRulesByStage() error: %v", err)
	}

	catalog := DiagnosticCatalog()
	if len(registrar.runners) != len(catalog)-1 {
		t.Fatalf("registered runners=%d, want %d", len(registrar.runners), len(catalog)-1)
	}
}

func TestLintRuleRunnerCheck(t *testing.T) {
	t.Parallel()

	var registrar lintRulesTestRegistrar
	if err := RegisterLintRules(&registrar); err != nil {
		t.Fatalf("RegisterLintRules() error: %v", err)
	}

	runner, ok := findRunnerByRuleID(
		registrar.runners,
		LintRuleID(CodeValidateDuplicateStageName),
	)
	if !ok {
		t.Fatalf(
			"runner for %q not found",
			LintRuleID(CodeValidateDuplicateStageName),
		)
	}

	runContext := lint.RunContext{
		TargetPath: "material.rvmat",
	}
	AttachLintDiagnostics(&runContext, []lint.Diagnostic{
		{
			Code:     mustRuleCode(t, CodeValidateDuplicateStageName),
			Severity: lint.SeverityError,
			Message:  "duplicate Stage name",
			Path:     "Stage1",
		},
		{
			Code:     mustRuleCode(t, CodeValidateDuplicateStageName),
			Severity: lint.SeverityError,
			Message:  "duplicate Stage name",
			Path:     "Stage2",
		},
		{
			Code:     mustRuleCode(t, CodeValidateUnknownTextureTag),
			Severity: lint.SeverityWarning,
			Message:  "unknown texture tag",
			Path:     "co",
		},
	})

	diagnostics := make([]lint.Diagnostic, 0, 2)
	err := runner.Check(
		context.Background(),
		&runContext,
		func(diagnostic lint.Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	if err != nil {
		t.Fatalf("runner.Check() error: %v", err)
	}

	if len(diagnostics) != 2 {
		t.Fatalf("len(Diagnostics)=%d, want 2", len(diagnostics))
	}

	for index := range diagnostics {
		if diagnostics[index].RuleID != LintRuleID(CodeValidateDuplicateStageName) {
			t.Fatalf("Diagnostics[%d].RuleID=%q", index, diagnostics[index].RuleID)
		}
	}
}

func TestLintRuleOptionsUnexpectedTextureExtension(t *testing.T) {
	t.Parallel()

	engine := linting.NewEngine()
	if err := RegisterLintRules(engine); err != nil {
		t.Fatalf("RegisterLintRules() error: %v", err)
	}

	runContext := lint.RunContext{
		TargetPath: "material.rvmat",
	}
	AttachLintDiagnostics(&runContext, []lint.Diagnostic{
		{
			Code:     mustRuleCode(t, CodeValidateUnexpectedTextureExtension),
			Severity: lint.SeverityWarning,
			Message:  "unexpected texture extension",
			Path:     `dz\gear\test\foo.jpg`,
		},
	})

	policy := linting.RunPolicy{
		Rules: map[string]linting.RuleSettings{
			LintRuleID(CodeValidateUnexpectedTextureExtension): {
				Options: UnexpectedTextureExtensionRuleOptions{
					AllowedExtensions: []string{".jpg"},
				},
			},
		},
	}

	result, err := engine.Run(context.Background(), runContext, &linting.RunOptions{
		Policy: &policy,
	})
	if err != nil {
		t.Fatalf("engine.Run() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(result.Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

func TestLintRuleOptionsUnknownTextureTag(t *testing.T) {
	t.Parallel()

	engine := linting.NewEngine()
	if err := RegisterLintRules(engine); err != nil {
		t.Fatalf("RegisterLintRules() error: %v", err)
	}

	runContext := lint.RunContext{
		TargetPath: "material.rvmat",
	}
	AttachLintDiagnostics(&runContext, []lint.Diagnostic{
		{
			Code:     mustRuleCode(t, CodeValidateUnknownTextureTag),
			Severity: lint.SeverityWarning,
			Message:  "unknown texture tag",
			Path:     "wat",
		},
	})

	policy := linting.RunPolicy{
		Rules: map[string]linting.RuleSettings{
			LintRuleID(CodeValidateUnknownTextureTag): {
				Options: UnknownTextureTagRuleOptions{
					AllowedTags: []string{"wat"},
				},
			},
		},
	}

	result, err := engine.Run(context.Background(), runContext, &linting.RunOptions{
		Policy: &policy,
	})
	if err != nil {
		t.Fatalf("engine.Run() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(result.Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

// findRunnerByRuleID returns runner by stable rule id.
func findRunnerByRuleID(
	runners []lint.RuleRunner,
	ruleID string,
) (lint.RuleRunner, bool) {
	for index := range runners {
		if runners[index].RuleSpec().ID == ruleID {
			return runners[index], true
		}
	}

	return nil, false
}

// mustRuleCode returns exported rule code token for one numeric code.
func mustRuleCode(t *testing.T, code lint.Code) string {
	t.Helper()

	spec, ok := DiagnosticByCode(code)
	if !ok {
		t.Fatalf("missing diagnostic catalog entry for code %d", code)
	}

	ruleSpec, err := DiagnosticRuleSpec(spec)
	if err != nil {
		t.Fatalf("DiagnosticRuleSpec() error: %v", err)
	}

	return ruleSpec.Code
}
