// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"sync"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// lintRunValueByCodeKey stores grouped issue map in run values.
	lintRunValueByCodeKey = "rvmat.lint.by_code"
)

var (
	// lintBindingState stores lazy-initialized code-catalog binding state.
	lintBindingState struct {
		// once guards one-time binding construction.
		once sync.Once

		// binding stores reusable register+attach helper.
		binding lint.CodeCatalogBinding[Issue]

		// err stores binding construction error.
		err error
	}
)

// LintRulesProvider registers rvmat diagnostic rules into any RuleRegistrar.
type LintRulesProvider struct{}

// RegisterRules adds provider-owned rules to target registrar.
func (provider LintRulesProvider) RegisterRules(
	registrar lint.RuleRegistrar,
) error {
	return RegisterLintRules(registrar)
}

// RegisterRulesByScope adds provider-owned rules filtered by scope tokens.
func (provider LintRulesProvider) RegisterRulesByScope(
	registrar lint.RuleRegistrar,
	scopes ...string,
) error {
	return RegisterLintRulesByScope(registrar, scopes...)
}

// RegisterRulesByStage adds provider-owned rules filtered by stage tokens.
func (provider LintRulesProvider) RegisterRulesByStage(
	registrar lint.RuleRegistrar,
	stages ...lint.Stage,
) error {
	return RegisterLintRulesByStage(registrar, stages...)
}

// RegisterLintRules registers stable rvmat diagnostic rules into registrar.
func RegisterLintRules(registrar lint.RuleRegistrar) error {
	return registerLintRulesWithBinding(registrar, func(
		binding lint.CodeCatalogBinding[Issue],
	) error {
		return binding.RegisterRules(registrar)
	})
}

// RegisterLintRulesByScope registers rvmat rules filtered by scope tokens.
func RegisterLintRulesByScope(
	registrar lint.RuleRegistrar,
	scopes ...string,
) error {
	return registerLintRulesWithBinding(registrar, func(
		binding lint.CodeCatalogBinding[Issue],
	) error {
		return binding.RegisterRulesByScope(registrar, scopes...)
	})
}

// RegisterLintRulesByStage registers rvmat rules filtered by stage tokens.
func RegisterLintRulesByStage(
	registrar lint.RuleRegistrar,
	stages ...lint.Stage,
) error {
	return registerLintRulesWithBinding(registrar, func(
		binding lint.CodeCatalogBinding[Issue],
	) error {
		return binding.RegisterRulesByStage(registrar, stages...)
	})
}

// registerLintRulesWithBinding validates registrar and executes binding callback.
func registerLintRulesWithBinding(
	registrar lint.RuleRegistrar,
	register func(binding lint.CodeCatalogBinding[Issue]) error,
) error {
	if registrar == nil {
		return ErrNilLintRuleRegistrar
	}

	binding, err := getLintBinding()
	if err != nil {
		return err
	}

	return register(binding)
}

// AttachLintDiagnostics stores issues in run context values.
func AttachLintDiagnostics(run *lint.RunContext, issues []Issue) {
	binding, err := getLintBinding()
	if err != nil {
		return
	}

	_ = binding.Attach(run, issues)
}

// getLintBinding returns lazy-initialized code-catalog binding helper.
func getLintBinding() (lint.CodeCatalogBinding[Issue], error) {
	lintBindingState.once.Do(func() {
		catalog, err := getDiagnosticCodeCatalog()
		if err != nil {
			lintBindingState.err = err
			return
		}

		lintBindingState.binding, lintBindingState.err = lint.NewCodeCatalogBinding(
			lint.CodeCatalogBindingConfig[Issue]{
				RunValueKey:        lintRunValueByCodeKey,
				Catalog:            catalog,
				CodeFromDiagnostic: lintIssueCode,
				DiagnosticToLint:   lintIssueDiagnostic,
				UnknownCodePolicy:  lint.UnknownCodeDrop,
			},
		)
	})

	if lintBindingState.err != nil {
		return lint.CodeCatalogBinding[Issue]{}, lintBindingState.err
	}

	return lintBindingState.binding, nil
}

// lintIssueCode extracts numeric code from internal issue item.
func lintIssueCode(item Issue) lint.Code {
	return item.Code
}

// lintIssueDiagnostic converts one rvmat issue into lintkit diagnostic.
func lintIssueDiagnostic(issue Issue) lint.Diagnostic {
	return lint.Diagnostic{
		Severity: issue.Level,
		Message:  issue.Message,
		Path:     issue.Path,
	}
}
