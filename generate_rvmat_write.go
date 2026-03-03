// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteGenerateSet writes generated materials to result output paths.
func WriteGenerateSet(result *GenerateSetResult, opt *FormatOptions) error {
	if result == nil {
		return errors.New("write generated rvmat result: nil result")
	}

	if err := writeGeneratedMaterial(result.MainOutputPath, result.Main, opt); err != nil {
		return fmt.Errorf("write generated rvmat result main: %w", err)
	}
	if err := writeGeneratedMaterial(result.DamageOutputPath, result.Damage, opt); err != nil {
		return fmt.Errorf("write generated rvmat result damage: %w", err)
	}
	if err := writeGeneratedMaterial(result.DestructOutputPath, result.Destruct, opt); err != nil {
		return fmt.Errorf("write generated rvmat result destruct: %w", err)
	}

	return nil
}

// writeGeneratedMaterial writes one generated material to disk.
func writeGeneratedMaterial(path string, m *Material, opt *FormatOptions) error {
	if strings.TrimSpace(path) == "" || m == nil {
		return nil
	}

	clean := filepath.Clean(path)
	dir := filepath.Dir(clean)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	formatted, err := Format(m, opt)
	if err != nil {
		return err
	}

	return os.WriteFile(clean, formatted, 0o600)
}
