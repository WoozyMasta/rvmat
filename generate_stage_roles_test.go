package rvmat

import "testing"

func TestStageIndexForTextureRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantIdx int
		wantOK  bool
	}{
		{name: "nohq", role: "nohq", wantIdx: 1, wantOK: true},
		{name: "smdi uppercase", role: "SMDI", wantIdx: 5, wantOK: true},
		{name: "env spaced", role: " env ", wantIdx: 7, wantOK: true},
		{name: "unknown", role: "foo", wantIdx: 0, wantOK: false},
		{name: "empty", role: "", wantIdx: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdx, gotOK := StageIndexForTextureRole(tt.role)
			if gotOK != tt.wantOK {
				t.Fatalf("StageIndexForTextureRole(%q) ok=%t, want %t", tt.role, gotOK, tt.wantOK)
			}
			if gotIdx != tt.wantIdx {
				t.Fatalf("StageIndexForTextureRole(%q) idx=%d, want %d", tt.role, gotIdx, tt.wantIdx)
			}
		})
	}
}

func TestNormalizeOverrideKeysToStage(t *testing.T) {
	in := map[string]string{
		"stage1": "a_stage_nohq.paa",
		"nohq":   "a_role_nohq.paa",
		"smdi":   "a_role_smdi.paa",
		"stage4": "a_stage_as.paa",
		"as":     "a_role_as.paa",
		"custom": "custom_value",
	}

	got := normalizeOverrideKeysToStage(in)

	if got["stage1"] != "a_stage_nohq.paa" {
		t.Fatalf("expected stage1 override priority, got %q", got["stage1"])
	}
	if got["stage5"] != "a_role_smdi.paa" {
		t.Fatalf("expected role->stage mapping for smdi, got %q", got["stage5"])
	}
	if got["stage4"] != "a_stage_as.paa" {
		t.Fatalf("expected stage4 override priority, got %q", got["stage4"])
	}
	if got["custom"] != "custom_value" {
		t.Fatalf("expected unknown key passthrough, got %q", got["custom"])
	}
	if _, ok := got["nohq"]; ok {
		t.Fatalf("expected role key nohq to be normalized into stage1")
	}
	if _, ok := got["as"]; ok {
		t.Fatalf("expected role key as to be normalized into stage4")
	}
}
