// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"strconv"
	"strings"
)

var (
	// superTextureRoleByStageName maps canonical Super stage names to role keys.
	superTextureRoleByStageName = map[string]string{
		"stage1": "nohq",
		"stage2": "dt",
		"stage3": "mc",
		"stage4": "as",
		"stage5": "smdi",
		"stage6": "fresnel",
		"stage7": "env",
	}
	// superStageIndexByTextureRole maps canonical texture role keys to stage index.
	superStageIndexByTextureRole = map[string]int{
		"nohq":    1,
		"dt":      2,
		"mc":      3,
		"as":      4,
		"smdi":    5,
		"fresnel": 6,
		"env":     7,
	}
)

// StageIndexForTextureRole returns Super stage index for a texture role key.
//
// Examples:
// - "nohq" -> 1
// - "smdi" -> 5
// - "env" -> 7
func StageIndexForTextureRole(role string) (int, bool) {
	idx, ok := superStageIndexByTextureRole[strings.ToLower(strings.TrimSpace(role))]
	return idx, ok
}

// textureRoleForStageName returns canonical role key for stage name.
func textureRoleForStageName(stageName string) (string, bool) {
	role, ok := superTextureRoleByStageName[strings.ToLower(strings.TrimSpace(stageName))]
	return role, ok
}

// stageNameForTextureRole returns canonical stage key for role key.
func stageNameForTextureRole(role string) (string, bool) {
	idx, ok := StageIndexForTextureRole(role)
	if !ok {
		return "", false
	}

	return "stage" + strconv.Itoa(idx), true
}

// normalizeOverrideKeysToStage normalizes override keys to stage keys.
//
// Stage keys have priority over role keys. For example, when both "stage1" and
// "nohq" are present, "stage1" wins.
func normalizeOverrideKeysToStage(in map[string]string) map[string]string {
	out := map[string]string{}
	if len(in) == 0 {
		return out
	}

	type pendingItem struct {
		key string
		raw string
	}

	pending := make([]pendingItem, 0, len(in))

	for key, raw := range in {
		normKey := strings.ToLower(strings.TrimSpace(key))
		normRaw := strings.TrimSpace(raw)
		if normKey == "" || normRaw == "" {
			continue
		}

		if isStageOverrideKey(normKey) {
			out[normKey] = normRaw
			continue
		}

		pending = append(pending, pendingItem{key: normKey, raw: normRaw})
	}

	for _, item := range pending {
		stageName, ok := stageNameForTextureRole(item.key)
		if ok {
			if _, exists := out[stageName]; !exists {
				out[stageName] = item.raw
			}
			continue
		}

		out[item.key] = item.raw
	}

	return out
}

// isStageOverrideKey reports whether key is a stage-like override key.
func isStageOverrideKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "stageti" {
		return true
	}

	if !strings.HasPrefix(k, "stage") || len(k) == len("stage") {
		return false
	}

	n, err := strconv.Atoi(k[len("stage"):])
	if err != nil {
		return false
	}

	return n >= 0
}
