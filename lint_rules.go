// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// lintRunValueByCodeKey stores grouped diagnostic map in run values.
	lintRunValueByCodeKey = "rvmat.lint.by_code"
)

var (
	// lintBindingState stores lazy-initialized code-catalog binding state.
	lintBindingState struct {
		// once guards one-time binding construction.
		once sync.Once

		// binding stores reusable register+attach helper.
		binding lint.CodeCatalogBinding[lint.Diagnostic]

		// err stores binding construction error.
		err error
	}

	// configurableLintCodes lists option-aware rules handled by custom runners.
	configurableLintCodes = []lint.Code{
		CodeValidateUnexpectedTextureExtension,
		CodeValidateUnknownTextureTag,
	}
)

// UnexpectedTextureExtensionRuleOptions configures RVMAT2008 behavior.
type UnexpectedTextureExtensionRuleOptions struct {
	// AllowedExtensions overrides allowed texture extensions.
	// Empty value keeps validator defaults.
	AllowedExtensions []string `json:"allowed_extensions,omitempty" yaml:"allowed_extensions,omitempty"`
}

// UnknownTextureTagRuleOptions configures RVMAT2028 behavior.
type UnknownTextureTagRuleOptions struct {
	// AllowedTags overrides allowed procedural color tags.
	// Empty value keeps validator defaults.
	AllowedTags []string `json:"allowed_tags,omitempty" yaml:"allowed_tags,omitempty"`
}

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
		binding lint.CodeCatalogBinding[lint.Diagnostic],
	) error {
		if err := binding.RegisterRules(registrar); err != nil {
			return err
		}

		return registerConfigurableRulesWithSource(registrar, configurableRuleRunners)
	})
}

// RegisterLintRulesByScope registers rvmat rules filtered by scope tokens.
func RegisterLintRulesByScope(
	registrar lint.RuleRegistrar,
	scopes ...string,
) error {
	return registerLintRulesWithBinding(registrar, func(
		binding lint.CodeCatalogBinding[lint.Diagnostic],
	) error {
		if err := binding.RegisterRulesByScope(registrar, scopes...); err != nil {
			return err
		}

		return registerConfigurableRulesWithSource(
			registrar,
			func() ([]lint.RuleRunner, error) {
				return configurableRuleRunnersByScope(scopes...)
			},
		)
	})
}

// RegisterLintRulesByStage registers rvmat rules filtered by stage tokens.
func RegisterLintRulesByStage(
	registrar lint.RuleRegistrar,
	stages ...lint.Stage,
) error {
	return registerLintRulesWithBinding(registrar, func(
		binding lint.CodeCatalogBinding[lint.Diagnostic],
	) error {
		if err := binding.RegisterRulesByStage(registrar, stages...); err != nil {
			return err
		}

		return registerConfigurableRulesWithSource(
			registrar,
			func() ([]lint.RuleRunner, error) {
				return configurableRuleRunnersByStage(stages...)
			},
		)
	})
}

// registerLintRulesWithBinding validates registrar and executes binding callback.
func registerLintRulesWithBinding(
	registrar lint.RuleRegistrar,
	register func(binding lint.CodeCatalogBinding[lint.Diagnostic]) error,
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

// registerConfigurableRulesWithSource registers configurable runners from source.
func registerConfigurableRulesWithSource(
	registrar lint.RuleRegistrar,
	runnersSource func() ([]lint.RuleRunner, error),
) error {
	runners, err := runnersSource()
	if err != nil {
		return err
	}

	if len(runners) == 0 {
		return nil
	}

	return registrar.Register(runners...)
}

// AttachLintDiagnostics stores diagnostics in run context values.
func AttachLintDiagnostics(run *lint.RunContext, diagnostics []lint.Diagnostic) {
	binding, err := getLintBinding()
	if err == nil {
		_ = binding.Attach(run, diagnostics)
	}
}

// getLintBinding returns lazy-initialized code-catalog binding helper.
func getLintBinding() (lint.CodeCatalogBinding[lint.Diagnostic], error) {
	lintBindingState.once.Do(func() {
		coreCatalog, err := coreDiagnosticCodeCatalog()
		if err != nil {
			lintBindingState.err = err
			return
		}

		lintBindingState.binding, lintBindingState.err = lint.NewCodeCatalogBinding(
			lint.CodeCatalogBindingConfig[lint.Diagnostic]{
				RunValueKey:        lintRunValueByCodeKey,
				Catalog:            coreCatalog,
				CodeFromDiagnostic: lintDiagnosticCode,
				DiagnosticToLint: func(diagnostic lint.Diagnostic) lint.Diagnostic {
					return diagnostic
				},
				UnknownCodePolicy: lint.UnknownCodeKeep,
			},
		)
	})

	if lintBindingState.err != nil {
		return lint.CodeCatalogBinding[lint.Diagnostic]{}, lintBindingState.err
	}

	return lintBindingState.binding, nil
}

// coreDiagnosticCodeCatalog returns catalog without configurable rules.
func coreDiagnosticCodeCatalog() (lint.CodeCatalog, error) {
	specs := DiagnosticCatalog()
	coreSpecs := make([]lint.CodeSpec, 0, len(specs))
	for index := range specs {
		code := specs[index].Code
		if code == CodeValidateUnexpectedTextureExtension ||
			code == CodeValidateUnknownTextureTag {
			continue
		}

		coreSpecs = append(coreSpecs, specs[index])
	}

	return lint.NewCodeCatalog(diagnosticCodeCatalogConfig, coreSpecs)
}

// lintDiagnosticCode extracts numeric code from one lint diagnostic item.
func lintDiagnosticCode(item lint.Diagnostic) lint.Code {
	code, ok := lint.ParsePublicCode(item.Code)
	if !ok {
		return 0
	}

	return code
}

// configurableRuleRunners returns runners for configurable lint rules.
func configurableRuleRunners() ([]lint.RuleRunner, error) {
	return configurableRuleRunnersByScope()
}

// configurableRuleRunnersByScope returns configurable runners filtered by scope.
func configurableRuleRunnersByScope(scopes ...string) ([]lint.RuleRunner, error) {
	catalog, err := getDiagnosticCodeCatalog()
	if err != nil {
		return nil, err
	}

	runners := make([]lint.RuleRunner, 0, 2)
	for _, code := range configurableLintCodes {
		spec, ok := catalog.ByCode(code)
		if !ok {
			continue
		}

		ruleSpec := catalog.RuleSpec(spec)
		if !scopeAllowed(ruleSpec.Scope, scopes...) {
			continue
		}

		runners = append(runners, configurableDiagnosticRuleRunner{
			code: code,
			spec: ruleSpec,
		})
	}

	return runners, nil
}

// configurableRuleRunnersByStage returns configurable runners filtered by stage.
func configurableRuleRunnersByStage(stages ...lint.Stage) ([]lint.RuleRunner, error) {
	if len(stages) == 0 {
		return configurableRuleRunners()
	}

	scopes := make([]string, 0, len(stages))
	for index := range stages {
		scopes = append(scopes, string(stages[index]))
	}

	return configurableRuleRunnersByScope(scopes...)
}

// scopeAllowed reports whether scope passes optional scope filters.
func scopeAllowed(scope string, scopes ...string) bool {
	if len(scopes) == 0 {
		return true
	}

	current := strings.TrimSpace(scope)
	for index := range scopes {
		if current == strings.TrimSpace(scopes[index]) {
			return true
		}
	}

	return false
}

// configurableDiagnosticRuleRunner executes option-aware filtering by code.
type configurableDiagnosticRuleRunner struct {
	// spec stores stable metadata for this runner.
	spec lint.RuleSpec
	// code stores stable lint code for this runner.
	code lint.Code
}

// RuleSpec returns stable metadata descriptor for current runner.
func (runner configurableDiagnosticRuleRunner) RuleSpec() lint.RuleSpec {
	return runner.spec
}

// Check runs one configurable rule against attached diagnostics index.
func (runner configurableDiagnosticRuleRunner) Check(
	_ context.Context,
	run *lint.RunContext,
	emit lint.DiagnosticEmit,
) error {
	diagnosticsByCode, ok := lint.GetIndexedByCode[lint.Diagnostic, lint.Code](
		run,
		lintRunValueByCodeKey,
	)
	if !ok || len(diagnosticsByCode) == 0 {
		return nil
	}

	diagnostics := diagnosticsByCode[runner.code]
	for diagnosticIndex := range diagnostics {
		diagnostic := diagnostics[diagnosticIndex]
		if !runner.shouldEmitDiagnostic(run, diagnostic) {
			continue
		}

		emit(lint.Diagnostic{
			RuleID:   runner.spec.ID,
			Code:     runner.spec.Code,
			Severity: diagnostic.Severity,
			Message:  diagnostic.Message,
			Path:     diagnostic.Path,
			Start:    diagnostic.Start,
			End:      diagnostic.End,
		})
	}

	return nil
}

// shouldEmitDiagnostic applies optional rule-level filtering for configurable rules.
func (runner configurableDiagnosticRuleRunner) shouldEmitDiagnostic(
	run *lint.RunContext,
	diagnostic lint.Diagnostic,
) bool {
	switch runner.code {
	case CodeValidateUnexpectedTextureExtension:
		return shouldEmitUnexpectedTextureExtension(run, diagnostic)
	case CodeValidateUnknownTextureTag:
		return shouldEmitUnknownTextureTag(run, diagnostic)
	default:
		return true
	}
}

// shouldEmitUnexpectedTextureExtension applies RVMAT2008 rule options.
func shouldEmitUnexpectedTextureExtension(
	run *lint.RunContext,
	diagnostic lint.Diagnostic,
) bool {
	allowed := defaultTextureExtensions
	if options, ok := lint.GetCurrentRuleOptions[UnexpectedTextureExtensionRuleOptions](run); ok {
		if len(options.AllowedExtensions) > 0 {
			allowed = options.AllowedExtensions
		}
	}

	allowedSet := buildAllowedTextureExtensionSet(allowed)
	extension := strings.ToLower(filepath.Ext(diagnostic.Path))
	_, isAllowed := allowedSet[extension]
	return !isAllowed
}

// shouldEmitUnknownTextureTag applies RVMAT2028 rule options.
func shouldEmitUnknownTextureTag(
	run *lint.RunContext,
	diagnostic lint.Diagnostic,
) bool {
	allowed := orderedKnownTextureTags()
	if options, ok := lint.GetCurrentRuleOptions[UnknownTextureTagRuleOptions](run); ok {
		if len(options.AllowedTags) > 0 {
			allowed = options.AllowedTags
		}
	}

	allowedSet := buildAllowedTextureTagSet(allowed)
	tag := strings.ToLower(strings.TrimSpace(diagnostic.Path))
	_, isAllowed := allowedSet[tag]
	return !isAllowed
}

// buildAllowedTextureTagSet returns normalized allowed texture-tag set.
func buildAllowedTextureTagSet(tags []string) map[string]struct{} {
	out := make(map[string]struct{}, len(tags))
	for index := range tags {
		tag := strings.ToLower(strings.TrimSpace(tags[index]))
		if tag == "" {
			continue
		}

		out[tag] = struct{}{}
	}

	return out
}
