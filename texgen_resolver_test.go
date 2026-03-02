package rvmat

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestResolveStageTexGen_InheritanceFromFixture(t *testing.T) {
	m, err := DecodeFile(
		filepath.Join("wiki", "game_data_examples", "structures", "bed_stacked.rvmat"),
		nil,
	)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	stage, ok := findStageByName(m.Stages, "Stage7")
	if !ok {
		t.Fatalf("stage not found")
	}

	got, err := ResolveStageTexGen(m, stage)
	if err != nil {
		t.Fatalf("ResolveStageTexGen: %v", err)
	}
	if got == nil {
		t.Fatalf("resolved texgen is nil")
	}

	if got.UVSource != "tex" {
		t.Fatalf("unexpected uvSource: %q", got.UVSource)
	}
	if got.UVTransform == nil {
		t.Fatalf("expected uvTransform")
	}

	if len(got.UVTransform.Aside) != 3 || got.UVTransform.Aside[0] != 10 {
		t.Fatalf("unexpected aside: %v", got.UVTransform.Aside)
	}
}

func TestResolveStageTexGen_OverrideMerge(t *testing.T) {
	m := &Material{
		TexGens: []TexGen{
			{
				Name:     "TexGen0",
				UVSource: "tex",
				UVTransform: &UVTransform{
					Aside: []float64{1, 0, 0},
					Up:    []float64{0, 1, 0},
					Dir:   []float64{0, 0, 1},
					Pos:   []float64{0, 0, 0},
				},
			},
			{
				Name:     "TexGen1",
				Base:     "TexGen0",
				UVSource: "tex1",
				UVTransform: &UVTransform{
					Aside: []float64{2, 0, 0},
				},
			},
		},
		Stages: []Stage{
			{Name: "Stage1", TexGen: "1"},
		},
	}

	got, err := ResolveStageTexGen(m, m.Stages[0])
	if err != nil {
		t.Fatalf("ResolveStageTexGen: %v", err)
	}
	if got == nil {
		t.Fatalf("resolved texgen is nil")
	}

	if got.UVSource != "tex1" {
		t.Fatalf("unexpected uvSource: %q", got.UVSource)
	}
	if got.UVTransform == nil {
		t.Fatalf("expected uvTransform")
	}

	if got.UVTransform.Aside[0] != 2 {
		t.Fatalf("unexpected aside: %v", got.UVTransform.Aside)
	}
	if got.UVTransform.Up[1] != 1 {
		t.Fatalf("base up was not inherited: %v", got.UVTransform.Up)
	}
}

func TestResolveStageTexGen_MissingReference(t *testing.T) {
	m := &Material{
		Stages: []Stage{
			{Name: "Stage1", TexGen: "2"},
		},
	}

	_, err := ResolveStageTexGen(m, m.Stages[0])
	if !errors.Is(err, ErrTexGenNotFound) {
		t.Fatalf("expected ErrTexGenNotFound, got %v", err)
	}
}

func TestResolveStageTexGen_MissingBase(t *testing.T) {
	m := &Material{
		TexGens: []TexGen{
			{Name: "TexGen0", Base: "TexGenX"},
		},
		Stages: []Stage{
			{Name: "Stage1", TexGen: "0"},
		},
	}

	_, err := ResolveStageTexGen(m, m.Stages[0])
	if !errors.Is(err, ErrTexGenBaseNotFound) {
		t.Fatalf("expected ErrTexGenBaseNotFound, got %v", err)
	}
}

func TestResolveStageTexGen_Cycle(t *testing.T) {
	m := &Material{
		TexGens: []TexGen{
			{Name: "TexGen0", Base: "TexGen1"},
			{Name: "TexGen1", Base: "TexGen0"},
		},
		Stages: []Stage{
			{Name: "Stage1", TexGen: "0"},
		},
	}

	_, err := ResolveStageTexGen(m, m.Stages[0])
	if !errors.Is(err, ErrTexGenCycle) {
		t.Fatalf("expected ErrTexGenCycle, got %v", err)
	}
}

func TestResolveStageTexGen_ByNameReference(t *testing.T) {
	m := &Material{
		TexGens: []TexGen{
			{Name: "TexGen0", UVSource: "tex"},
		},
		Stages: []Stage{
			{Name: "Stage1", TexGen: "TexGen0"},
		},
	}

	got, err := ResolveStageTexGen(m, m.Stages[0])
	if err != nil {
		t.Fatalf("ResolveStageTexGen: %v", err)
	}
	if got == nil || got.UVSource != "tex" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestEffectiveUVFallbackWithoutTexGen(t *testing.T) {
	stage := Stage{
		Name:     "Stage1",
		UVSource: "tex",
		UVTransform: &UVTransform{
			Aside: []float64{1, 0, 0},
		},
	}

	src, err := EffectiveUVSource(nil, stage)
	if err != nil {
		t.Fatalf("EffectiveUVSource: %v", err)
	}
	if src != "tex" {
		t.Fatalf("unexpected uvSource: %q", src)
	}

	uv, err := EffectiveUVTransform(nil, stage)
	if err != nil {
		t.Fatalf("EffectiveUVTransform: %v", err)
	}
	if uv == nil {
		t.Fatalf("expected uvTransform")
	}

	uv.Aside[0] = 42
	if stage.UVTransform.Aside[0] == 42 {
		t.Fatalf("expected clone, got shared pointer")
	}
}

// findStageByName returns stage by name.
func findStageByName(stages []Stage, name string) (Stage, bool) {
	for _, st := range stages {
		if st.Name == name {
			return st, true
		}
	}

	return Stage{}, false
}
