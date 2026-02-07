package rvmat

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// TextureKind indicates texture reference type.
type TextureKind string

const (
	// TextureKindPath represents a file path texture reference.
	TextureKindPath TextureKind = "path"
	// TextureKindProcedural represents a procedural texture reference.
	TextureKindProcedural TextureKind = "procedural"
)

// TextureRef represents a texture reference string.
type TextureRef struct {
	Procedural *ProceduralTexture `json:"procedural,omitempty" yaml:"procedural,omitempty"` // Parsed procedural texture expression
	Raw        string             `json:"raw,omitempty" yaml:"raw,omitempty"`               // Raw texture reference string
	Kind       TextureKind        `json:"kind,omitempty" yaml:"kind,omitempty"`             // Texture reference type
	ParsedOK   bool               `json:"parsedOk,omitempty" yaml:"parsedOk,omitempty"`     // Whether the texture reference was parsed successfully
}

// Validate validates this texture reference.
func (t TextureRef) Validate(opt *TextureValidateOptions) []Issue {
	vopt := opt.normalize()
	return validateTexture(t, vopt)
}

// ProceduralTexture is procedural texture expression.
//
// Example: "#(argb,8,8,3)color(0.5,0.5,0.5,1.0,co)"
// - Format/Width/Height/Mip come from the header "#(argb,8,8,3)".
// - Func/Args come from "color(...)".
type ProceduralTexture struct {
	// Color is procedural texture color(r,g,b,a[,tag]).
	Color *ProceduralColor `json:"color,omitempty" yaml:"color,omitempty"`
	// Fresnel is procedural texture fresnel(a,b) and fresnelGlass(a,b?).
	Fresnel *ProceduralFresnel `json:"fresnel,omitempty" yaml:"fresnel,omitempty"`
	// Irradiance is procedural texture irradiance(x).
	Irradiance *ProceduralIrradiance `json:"irradiance,omitempty" yaml:"irradiance,omitempty"`

	Format string   `json:"format,omitempty" yaml:"format,omitempty"` // Procedural texture format
	Func   string   `json:"func,omitempty" yaml:"func,omitempty"`     // Procedural function name
	Args   []string `json:"args,omitempty" yaml:"args,omitempty"`     // Procedural function arguments
	Width  int      `json:"width,omitempty" yaml:"width,omitempty"`   // Texture width
	Height int      `json:"height,omitempty" yaml:"height,omitempty"` // Texture height
	Mip    int      `json:"mip,omitempty" yaml:"mip,omitempty"`       // Texture mip level

}

// ProceduralColor is procedural texture color(r,g,b,a[,tag]).
type ProceduralColor struct {
	Tag string  `json:"tag,omitempty" yaml:"tag,omitempty"` // Texture tag
	R   float64 `json:"r,omitempty" yaml:"r,omitempty"`     // Red color component
	G   float64 `json:"g,omitempty" yaml:"g,omitempty"`     // Green color component
	B   float64 `json:"b,omitempty" yaml:"b,omitempty"`     // Blue color component
	A   float64 `json:"a,omitempty" yaml:"a,omitempty"`     // Alpha color component
}

// ProceduralFresnel is procedural texture fresnel(a,b) and fresnelGlass(a,b?).
type ProceduralFresnel struct {
	A float64 `json:"a,omitempty" yaml:"a,omitempty"` // Fresnel parameter a
	B float64 `json:"b,omitempty" yaml:"b,omitempty"` // Fresnel parameter b
}

// ProceduralIrradiance is procedural texture irradiance(x).
type ProceduralIrradiance struct {
	Value float64 `json:"value,omitempty" yaml:"value,omitempty"` // Irradiance value
}

// NewProcedural creates a procedural texture reference from parts.
// Args can be strings or numbers; numeric args are formatted consistently.
func NewProcedural(format string, width, height, mip int, fn string, args ...any) TextureRef {
	raw := buildProceduralRaw(format, width, height, mip, fn, formatProceduralArgs(args...))
	return ParseTextureRef(raw)
}

// NewProceduralColor creates a color(...) procedural texture reference.
func NewProceduralColor(format string, width, height, mip int, r, g, b, a float64, tag string) TextureRef {
	args := []any{r, g, b, a}
	if tag != "" {
		args = append(args, tag)
	}
	return NewProcedural(format, width, height, mip, "color", args...)
}

// NewProceduralFresnel creates a fresnel(a,b) procedural texture reference.
func NewProceduralFresnel(format string, width, height, mip int, a, b float64) TextureRef {
	return NewProcedural(format, width, height, mip, "fresnel", a, b)
}

// NewProceduralFresnelGlass creates a fresnelGlass(a[,b]) procedural texture reference.
func NewProceduralFresnelGlass(format string, width, height, mip int, a, b float64, hasB bool) TextureRef {
	if hasB {
		return NewProcedural(format, width, height, mip, "fresnelGlass", a, b)
	}
	return NewProcedural(format, width, height, mip, "fresnelGlass", a)
}

// NewProceduralIrradiance creates an irradiance(x) procedural texture reference.
func NewProceduralIrradiance(format string, width, height, mip int, value float64) TextureRef {
	return NewProcedural(format, width, height, mip, "irradiance", value)
}

// ParseTextureRef parses a texture reference string.
func ParseTextureRef(raw string) TextureRef {
	raw = NormalizeTextureRaw(raw)

	// Initialize TextureRef
	tr := TextureRef{Raw: raw}

	// Check if the texture reference is a procedural texture
	if strings.HasPrefix(raw, "#(") {
		tr.Kind = TextureKindProcedural
		if pt, ok := parseProcedural(raw); ok {
			tr.Procedural = pt
			tr.ParsedOK = true
		}

		return tr
	}

	tr.Kind = TextureKindPath
	return tr
}

// IsProcedural reports whether the texture is procedural.
func (t TextureRef) IsProcedural() bool { return t.Kind == TextureKindProcedural }

// IsPath reports whether the texture is a file path.
func (t TextureRef) IsPath() bool { return t.Kind == TextureKindPath }

// PathResolver resolves texture paths relative to GameRoot.
type PathResolver struct {
	GameRoot string
}

// ResolveTexturePath resolves a texture path against GameRoot.
// Returns empty string for procedural textures.
func (r PathResolver) ResolveTexturePath(tex TextureRef) string {
	if tex.IsProcedural() {
		return ""
	}

	return r.ResolvePath(tex.Raw)
}

// ResolvePath resolves a raw path against GameRoot.
func (r PathResolver) ResolvePath(raw string) string {
	if raw == "" {
		return ""
	}

	norm := normalizeOSPath(raw)
	if filepath.IsAbs(norm) || hasVolume(norm) {
		return filepath.Clean(norm)
	}

	if r.GameRoot == "" {
		return filepath.Clean(norm)
	}

	return filepath.Clean(filepath.Join(r.GameRoot, norm))
}

// hasVolume checks if the path has a volume.
func hasVolume(p string) bool {
	if len(p) >= 2 && p[1] == ':' {
		return true
	}
	return false
}

// normalizeOSPath normalizes a path for OS-specific separators.
func normalizeOSPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	return filepath.FromSlash(p)
}

func parseProcedural(raw string) (*ProceduralTexture, bool) {
	// Parse minimal procedural form: "#(fmt,w,h,mip)func(args...)".
	if !strings.HasPrefix(raw, "#(") {
		return nil, false
	}

	closeIdx := strings.Index(raw, ")")
	if closeIdx < 0 {
		return nil, false
	}

	head := raw[2:closeIdx]
	headParts := splitCSV(head)
	if len(headParts) < 4 {
		return nil, false
	}

	// Parse width
	w, err := strconv.Atoi(strings.TrimSpace(headParts[1]))
	if err != nil {
		return nil, false
	}

	// Parse height
	h, err := strconv.Atoi(strings.TrimSpace(headParts[2]))
	if err != nil {
		return nil, false
	}

	// Parse mip level
	mip, err := strconv.Atoi(strings.TrimSpace(headParts[3]))
	if err != nil {
		return nil, false
	}

	// Parse remaining string
	remain := strings.TrimSpace(raw[closeIdx+1:])
	if remain == "" {
		return nil, false
	}

	// Parse function name and arguments
	funcName, args, ok := parseFunc(remain)
	if !ok {
		return nil, false
	}

	// Initialize ProceduralTexture
	pt := &ProceduralTexture{
		Format: strings.TrimSpace(headParts[0]),
		Width:  w,
		Height: h,
		Mip:    mip,
		Func:   funcName,
		Args:   args,
	}
	parseProceduralArgs(pt)

	return pt, true
}

// NormalizeTextureRaw attempts to clean malformed texture strings seen in the wild.
func NormalizeTextureRaw(raw string) string {
	s := strings.TrimSpace(raw)
	ls := strings.ToLower(s)

	if strings.HasPrefix(ls, "texture=\"") {
		s = s[len(`texture="`):]
		s = strings.TrimSuffix(s, "\";")
		s = strings.TrimSuffix(s, "\"")
		return strings.TrimSpace(s)
	}

	s = strings.TrimSuffix(s, "\";")
	return strings.TrimSpace(s)
}

func parseFunc(s string) (string, []string, bool) {
	open := strings.Index(s, "(")
	closeIdx := strings.LastIndex(s, ")")
	if open <= 0 || closeIdx <= open {
		return "", nil, false
	}

	// Function name is everything before '('; args are comma-separated.
	name := strings.TrimSpace(s[:open])
	argsRaw := s[open+1 : closeIdx]
	args := splitCSV(argsRaw)
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}

	return name, args, true
}

// splitCSV splits a CSV string into a slice of strings.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out
}

// parseProceduralArgs parses the arguments of a procedural texture.
func parseProceduralArgs(pt *ProceduralTexture) {
	switch strings.ToLower(pt.Func) {
	case "color":
		parseProceduralColor(pt)
	case "fresnel":
		parseProceduralFresnel(pt)
	case "fresnelglass":
		parseProceduralFresnel(pt)
	case "irradiance":
		parseProceduralIrradiance(pt)
	}
}

// parseProceduralColor parses a color(...) procedural texture.
func parseProceduralColor(pt *ProceduralTexture) {
	if len(pt.Args) != 4 && len(pt.Args) != 5 {
		return
	}
	r, ok := parseFloatArg(pt.Args[0])
	if !ok {
		return
	}
	g, ok := parseFloatArg(pt.Args[1])
	if !ok {
		return
	}
	b, ok := parseFloatArg(pt.Args[2])
	if !ok {
		return
	}
	a, ok := parseFloatArg(pt.Args[3])
	if !ok {
		return
	}

	tag := ""
	if len(pt.Args) == 5 {
		tag = pt.Args[4]
	}

	pt.Color = &ProceduralColor{R: r, G: g, B: b, A: a, Tag: tag}
}

// parseProceduralFresnel parses a fresnel(a,b) procedural texture.
func parseProceduralFresnel(pt *ProceduralTexture) {
	if len(pt.Args) != 1 && len(pt.Args) != 2 {
		return
	}

	a, ok := parseFloatArg(pt.Args[0])
	if !ok {
		return
	}

	b := 0.0
	if len(pt.Args) == 2 {
		var ok2 bool
		b, ok2 = parseFloatArg(pt.Args[1])
		if !ok2 {
			return
		}
	}

	pt.Fresnel = &ProceduralFresnel{A: a, B: b}
}

// parseProceduralIrradiance parses an irradiance(x) procedural texture.
func parseProceduralIrradiance(pt *ProceduralTexture) {
	if len(pt.Args) != 1 {
		return
	}

	v, ok := parseFloatArg(pt.Args[0])
	if !ok {
		return
	}

	pt.Irradiance = &ProceduralIrradiance{Value: v}
}

// parseFloatArg parses a string to a float64 value.
func parseFloatArg(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

// formatFloat formats a float64 value to a string.
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// formatProceduralArgs formats a slice of any type to a slice of strings.
func formatProceduralArgs(args ...any) []string {
	if len(args) == 0 {
		return nil
	}

	out := make([]string, 0, len(args))
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			out = append(out, v)
		case float64:
			out = append(out, formatFloat(v))
		case float32:
			out = append(out, formatFloat(float64(v)))
		case int:
			out = append(out, strconv.Itoa(v))
		case int64:
			out = append(out, strconv.FormatInt(v, 10))
		case int32:
			out = append(out, strconv.FormatInt(int64(v), 10))
		case uint:
			out = append(out, strconv.FormatUint(uint64(v), 10))
		case uint64:
			out = append(out, strconv.FormatUint(v, 10))
		case uint32:
			out = append(out, strconv.FormatUint(uint64(v), 10))
		default:
			out = append(out, fmt.Sprint(v))
		}
	}

	return out
}

// buildProceduralRaw builds a procedural texture reference string from parts.
func buildProceduralRaw(format string, width, height, mip int, fn string, args []string) string {
	var b strings.Builder
	b.Grow(64 + len(fn) + len(format) + len(args)*8)
	b.WriteString("#(")
	b.WriteString(format)
	b.WriteByte(',')
	b.WriteString(strconv.Itoa(width))
	b.WriteByte(',')
	b.WriteString(strconv.Itoa(height))
	b.WriteByte(',')
	b.WriteString(strconv.Itoa(mip))
	b.WriteByte(')')
	b.WriteString(fn)
	b.WriteByte('(')
	b.WriteString(strings.Join(args, ","))
	b.WriteByte(')')

	return b.String()
}
