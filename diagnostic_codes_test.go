// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"reflect"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestDiagnosticCatalogIntegrity(t *testing.T) {
	t.Parallel()

	catalog := DiagnosticCatalog()
	if len(catalog) == 0 {
		t.Fatal("expected non-empty diagnostics catalog")
	}

	seen := make(map[lint.Code]struct{}, len(catalog))
	for _, spec := range catalog {
		if spec.Code == 0 {
			t.Fatal("diagnostic catalog contains empty code")
		}

		if spec.Stage == "" {
			t.Fatalf("diagnostic %d has empty stage", spec.Code)
		}

		if spec.Severity == "" {
			t.Fatalf("diagnostic %d has empty severity", spec.Code)
		}

		if spec.Message == "" {
			t.Fatalf("diagnostic %d has empty message", spec.Code)
		}

		if _, ok := seen[spec.Code]; ok {
			t.Fatalf("diagnostic catalog has duplicate code %d", spec.Code)
		}

		seen[spec.Code] = struct{}{}
		lookup, ok := DiagnosticByCode(spec.Code)
		if !ok {
			t.Fatalf("DiagnosticByCode(%d) returned not found", spec.Code)
		}

		if !reflect.DeepEqual(lookup, spec) {
			t.Fatalf(
				"DiagnosticByCode(%d) returned different spec: %+v != %+v",
				spec.Code,
				lookup,
				spec,
			)
		}
	}
}

func TestDiagnosticByCodeUnknown(t *testing.T) {
	t.Parallel()

	if _, ok := DiagnosticByCode(0); ok {
		t.Fatal("expected unknown diagnostic code lookup to fail")
	}
}
