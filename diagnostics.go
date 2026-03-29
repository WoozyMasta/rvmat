// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import "github.com/woozymasta/lintkit/lint"

// warningDiagnostic builds one warning-level lint diagnostic.
func warningDiagnostic(
	code lint.Code,
	message string,
	path string,
) lint.Diagnostic {
	return diagnosticWithSeverity(code, lint.SeverityWarning, message, path)
}

// errorDiagnostic builds one error-level lint diagnostic.
func errorDiagnostic(
	code lint.Code,
	message string,
	path string,
) lint.Diagnostic {
	return diagnosticWithSeverity(code, lint.SeverityError, message, path)
}

// diagnosticWithSeverity builds one diagnostic from catalog code and payload.
func diagnosticWithSeverity(
	code lint.Code,
	severity lint.Severity,
	message string,
	path string,
) lint.Diagnostic {
	diagnostic := lint.Diagnostic{
		Severity: severity,
		Message:  message,
		Path:     path,
	}

	spec, ok := DiagnosticByCode(code)
	if !ok {
		return diagnostic
	}

	ruleSpec, err := DiagnosticRuleSpec(spec)
	if err != nil {
		return diagnostic
	}

	diagnostic.RuleID = ruleSpec.ID
	diagnostic.Code = ruleSpec.Code
	return diagnostic
}
