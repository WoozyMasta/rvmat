// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	prettyFloatDigits = 4
)

// formatPrettyFloat formats float values without common IEEE tail noise.
func formatPrettyFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return strconv.FormatFloat(v, 'g', -1, 64)
	}

	short := strconv.FormatFloat(v, 'f', prettyFloatDigits, 64)
	short = strings.TrimRight(short, "0")
	short = strings.TrimRight(short, ".")
	if short == "" || short == "-0" {
		short = "0"
	}

	return short
}

// Encode writes a Material to writer.
func Encode(w io.Writer, m *Material, opt *FormatOptions) error {
	fopt := opt.normalize()
	bw, owned := toBufferedWriter(w)
	wr := &writer{
		w:             bw,
		indent:        fopt.Indent,
		compactStages: fopt.CompactStages,
	}
	if err := wr.writeMaterial(m); err != nil {
		return err
	}

	if !owned {
		return nil
	}

	return bw.Flush()
}

// EncodeFile writes a Material to a file.
func EncodeFile(path string, m *Material, opt *FormatOptions) error {
	b, err := Format(m, opt)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o600)
}

// Format renders a Material to bytes.
func Format(m *Material, opt *FormatOptions) ([]byte, error) {
	fopt := opt.normalize()
	var buf bytes.Buffer
	wr := &writer{
		w:             &buf,
		indent:        fopt.Indent,
		compactStages: fopt.CompactStages,
	}
	if err := wr.writeMaterial(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// toBufferedWriter reuses a buffered writer when available.
func toBufferedWriter(w io.Writer) (*bufio.Writer, bool) {
	if bw, ok := w.(*bufio.Writer); ok {
		return bw, false
	}

	return bufio.NewWriter(w), true
}

// writer writes a Material to a writer.
type writer struct {
	w             io.Writer // Writer to write to
	indent        string    // Indentation string
	level         int       // Current nesting level
	compactStages bool      // Compact formatting for simple StageN classes
}

// writeMaterial writes a Material to the writer.
func (w *writer) writeMaterial(m *Material) error {
	writeArray := func(name string, vals []float64) error {
		if len(vals) == 0 {
			return nil
		}
		if err := w.writeString(name); err != nil {
			return err
		}
		if err := w.writeString("[]="); err != nil {
			return err
		}
		if err := w.writeFloatArray(vals); err != nil {
			return err
		}
		return w.writeString(";\n")
	}

	// Write color arrays
	if err := writeArray("ambient", m.Ambient); err != nil {
		return err
	}
	if err := writeArray("diffuse", m.Diffuse); err != nil {
		return err
	}
	if err := writeArray("forcedDiffuse", m.ForcedDiffuse); err != nil {
		return err
	}
	if err := writeArray("emmisive", m.Emissive); err != nil {
		return err
	}
	if err := writeArray("specular", m.Specular); err != nil {
		return err
	}

	// Write specular power
	if m.SpecularPower != nil {
		if err := w.writeString("specularPower="); err != nil {
			return err
		}
		if err := w.writeNumber(*m.SpecularPower); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write pixel shader ID
	if m.PixelShaderID != "" {
		if err := w.writeString("PixelShaderID="); err != nil {
			return err
		}
		if err := w.writeQuoted(m.PixelShaderID); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write vertex shader ID
	if m.VertexShaderID != "" {
		if err := w.writeString("VertexShaderID="); err != nil {
			return err
		}
		if err := w.writeQuoted(m.VertexShaderID); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write texture generators
	for _, tg := range m.TexGens {
		if err := w.writeTexGen(tg); err != nil {
			return err
		}
	}

	// Write stages
	for _, st := range m.Stages {
		if err := w.writeStage(st); err != nil {
			return err
		}
	}

	// Write extras
	for _, n := range m.extras {
		if err := w.writeNode(n); err != nil {
			return err
		}
	}

	return nil
}

// writeStage writes a Stage to the writer.
func (w *writer) writeStage(s Stage) error {
	name := s.Name
	if name == "" {
		name = "Stage"
	}

	if w.compactStages && isCompactStageCandidate(name, s) {
		return w.writeCompactStage(name, s)
	}

	// Write stage class
	if err := w.writeString("class "); err != nil {
		return err
	}
	if err := w.writeString(name); err != nil {
		return err
	}
	if err := w.writeString("\n{\n"); err != nil {
		return err
	}

	// Write stage body
	w.level++
	if s.Texture.Raw != "" {
		if err := w.writeAssign("texture", value{Kind: valueString, Str: s.Texture.Raw}, false); err != nil {
			return err
		}
	}
	if s.UVSource != "" && s.TexGen == "" {
		if err := w.writeAssign("uvSource", value{Kind: valueString, Str: s.UVSource}, false); err != nil {
			return err
		}
	}
	if s.TexGen != "" {
		if err := w.writeAssign("texGen", value{Kind: valueString, Str: s.TexGen}, false); err != nil {
			return err
		}
	}
	if s.UVTransform != nil && s.TexGen == "" {
		if err := w.writeUVTransform(*s.UVTransform); err != nil {
			return err
		}
	}

	// Write extras
	for _, n := range s.extras {
		if err := w.writeNode(n); err != nil {
			return err
		}
	}
	w.level--

	// Write stage end
	return w.writeString("};\n")
}

// isCompactStageCandidate reports whether StageN can be rendered in one line.
func isCompactStageCandidate(stageName string, stage Stage) bool {
	if strings.TrimSpace(stage.Texture.Raw) == "" {
		return false
	}
	if strings.TrimSpace(stage.TexGen) == "" {
		return false
	}
	if strings.TrimSpace(stage.UVSource) != "" || stage.UVTransform != nil {
		return false
	}
	if len(stage.extras) > 0 {
		return false
	}
	if _, ok := parseIndexedName(stageName, "stage"); !ok {
		return false
	}

	return true
}

// writeCompactStage writes one-line StageN class with texture and texGen.
func (w *writer) writeCompactStage(stageName string, stage Stage) error {
	if err := w.writeString("class "); err != nil {
		return err
	}
	if err := w.writeString(stageName); err != nil {
		return err
	}
	if err := w.writeString(" { texture = "); err != nil {
		return err
	}
	if err := w.writeQuoted(NormalizeGameTexturePath(stage.Texture.Raw)); err != nil {
		return err
	}
	if err := w.writeString("; texGen = "); err != nil {
		return err
	}

	texGen := strings.TrimSpace(stage.TexGen)
	if n, err := strconv.Atoi(texGen); err == nil {
		if err := w.writeString(strconv.Itoa(n)); err != nil {
			return err
		}
	} else {
		if err := w.writeQuoted(texGen); err != nil {
			return err
		}
	}

	return w.writeString("; };\n")
}

// writeTexGen writes a TexGen to the writer.
func (w *writer) writeTexGen(t TexGen) error {
	name := t.Name
	if name == "" {
		name = "TexGen"
	}

	// Write texgen class with base or without base
	if t.Base != "" {
		if err := w.writeString("class "); err != nil {
			return err
		}
		if err := w.writeString(name); err != nil {
			return err
		}
		if err := w.writeString(" : "); err != nil {
			return err
		}
		if err := w.writeString(t.Base); err != nil {
			return err
		}
		if err := w.writeString("\n"); err != nil {
			return err
		}
	} else {
		if err := w.writeString("class "); err != nil {
			return err
		}
		if err := w.writeString(name); err != nil {
			return err
		}
		if err := w.writeString("\n"); err != nil {
			return err
		}
	}

	// Write texgen body
	if err := w.writeString("{\n"); err != nil {
		return err
	}
	w.level++
	if t.UVSource != "" {
		if err := w.writeAssign("uvSource", value{Kind: valueString, Str: t.UVSource}, false); err != nil {
			return err
		}
	}
	if t.UVTransform != nil {
		if err := w.writeUVTransform(*t.UVTransform); err != nil {
			return err
		}
	}

	// Write extras
	for _, n := range t.extras {
		if err := w.writeNode(n); err != nil {
			return err
		}
	}

	// Write texgen end
	w.level--
	return w.writeString("};\n")
}

// writeUVTransform writes a UVTransform to the writer.
func (w *writer) writeUVTransform(uv UVTransform) error {
	// Write uvTransform class
	if err := w.writeIndent(); err != nil {
		return err
	}
	if err := w.writeString("class uvTransform\n"); err != nil {
		return err
	}
	if err := w.writeIndent(); err != nil {
		return err
	}
	if err := w.writeString("{\n"); err != nil {
		return err
	}

	// Write uvTransform body
	w.level++
	if len(uv.Aside) > 0 {
		if err := w.writeIndent(); err != nil {
			return err
		}
		if err := w.writeString("aside[]="); err != nil {
			return err
		}
		if err := w.writeFloatArray(uv.Aside); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write up array
	if len(uv.Up) > 0 {
		if err := w.writeIndent(); err != nil {
			return err
		}
		if err := w.writeString("up[]="); err != nil {
			return err
		}
		if err := w.writeFloatArray(uv.Up); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write dir array
	if len(uv.Dir) > 0 {
		if err := w.writeIndent(); err != nil {
			return err
		}
		if err := w.writeString("dir[]="); err != nil {
			return err
		}
		if err := w.writeFloatArray(uv.Dir); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write pos array
	if len(uv.Pos) > 0 {
		if err := w.writeIndent(); err != nil {
			return err
		}
		if err := w.writeString("pos[]="); err != nil {
			return err
		}
		if err := w.writeFloatArray(uv.Pos); err != nil {
			return err
		}
		if err := w.writeString(";\n"); err != nil {
			return err
		}
	}

	// Write uvTransform end
	w.level--
	if err := w.writeIndent(); err != nil {
		return err
	}

	return w.writeString("};\n")
}

// writeNode writes a node to the writer.
func (w *writer) writeNode(n node) error {
	switch t := n.(type) {
	case assignNode:
		return w.writeAssign(t.Name, t.Value, t.IsArray)
	case classNode:
		return w.writeClass(t)
	default:
		return nil
	}
}

// writeClass writes a classNode to the writer.
func (w *writer) writeClass(c classNode) error {
	if err := w.writeIndent(); err != nil {
		return err
	}

	// Write class with base or without base
	if c.Base != "" {
		if err := w.writeString("class "); err != nil {
			return err
		}
		if err := w.writeString(c.Name); err != nil {
			return err
		}
		if err := w.writeString(" : "); err != nil {
			return err
		}
		if err := w.writeString(c.Base); err != nil {
			return err
		}
		if err := w.writeString("\n"); err != nil {
			return err
		}
	} else {
		if err := w.writeString("class "); err != nil {
			return err
		}
		if err := w.writeString(c.Name); err != nil {
			return err
		}
		if err := w.writeString("\n"); err != nil {
			return err
		}
	}

	// Write class body
	if err := w.writeIndent(); err != nil {
		return err
	}
	if err := w.writeString("{\n"); err != nil {
		return err
	}
	w.level++
	for _, n := range c.Body {
		if err := w.writeNode(n); err != nil {
			return err
		}
	}
	w.level--
	if err := w.writeIndent(); err != nil {
		return err
	}

	return w.writeString("};\n")
}

// writeAssign writes an assignNode to the writer.
func (w *writer) writeAssign(name string, val value, isArray bool) error {
	if val.Kind == valueString && matchKey(name, "texture", true) {
		val.Str = NormalizeGameTexturePath(val.Str)
	}

	if err := w.writeIndent(); err != nil {
		return err
	}

	// Write assign as array
	if isArray {
		if err := w.writeString(name); err != nil {
			return err
		}
		if err := w.writeString("[]="); err != nil {
			return err
		}
		if err := w.writeValue(val); err != nil {
			return err
		}
		return w.writeString(";\n")
	}

	// Write assign as single value
	if err := w.writeString(name); err != nil {
		return err
	}
	if err := w.writeString("="); err != nil {
		return err
	}
	if err := w.writeValue(val); err != nil {
		return err
	}

	return w.writeString(";\n")
}

// writeIndent writes the current indentation level to the writer.
func (w *writer) writeIndent() error {
	if w.level <= 0 || w.indent == "" {
		return nil
	}

	for i := 0; i < w.level; i++ {
		if err := w.writeString(w.indent); err != nil {
			return err
		}
	}

	return nil
}

// writeNumber writes a float64 value to the writer.
func (w *writer) writeNumber(v float64) error {
	_, err := io.WriteString(w.w, formatPrettyFloat(v))

	return err
}

// writeQuoted writes a quoted string to the writer.
func (w *writer) writeQuoted(s string) error {
	if err := w.writeString("\""); err != nil {
		return err
	}
	if err := w.writeString(s); err != nil {
		return err
	}

	return w.writeString("\"")
}

// writeString writes a string to the writer.
func (w *writer) writeString(s string) error {
	_, err := io.WriteString(w.w, s)
	return err
}

// writeValue writes a value to the writer.
func (w *writer) writeValue(v value) error {
	switch v.Kind {
	case valueNumber:
		return w.writeNumber(v.Num)
	case valueString:
		return w.writeQuoted(v.Str)
	case valueIdent:
		return w.writeString(v.Str)
	case valueArray:
		return w.writeArray(v.Array)
	default:
		return nil
	}
}

// writeArray writes an array of values to the writer.
func (w *writer) writeArray(vals []value) error {
	if err := w.writeString("{"); err != nil {
		return err
	}

	// Write array values
	for i, v := range vals {
		if i > 0 {
			if err := w.writeString(", "); err != nil {
				return err
			}
		}
		if err := w.writeValue(v); err != nil {
			return err
		}
	}

	// Write array end
	return w.writeString("}")
}

// writeFloatArray writes a slice of float64 values to the writer.
func (w *writer) writeFloatArray(vals []float64) error {
	if err := w.writeString("{"); err != nil {
		return err
	}

	// Write float array values
	for i, v := range vals {
		if i > 0 {
			if err := w.writeString(", "); err != nil {
				return err
			}
		}
		if err := w.writeNumber(v); err != nil {
			return err
		}
	}

	// Write float array end
	return w.writeString("}")
}
