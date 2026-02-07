package rvmat

// Color represents RGBA color.
type Color struct {
	R float64 `json:"r,omitempty" yaml:"r,omitempty"` // Red channel component
	G float64 `json:"g,omitempty" yaml:"g,omitempty"` // Green channel component
	B float64 `json:"b,omitempty" yaml:"b,omitempty"` // Blue channel component
	A float64 `json:"a,omitempty" yaml:"a,omitempty"` // Alpha channel component
}

// Clamp01 clamps v to [0,1].
func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// SetColorRGBA creates a Color from RGBA values.
func SetColorRGBA(r, g, b, a float64) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// SetColorRGB creates a Color with alpha=1.
func SetColorRGB(r, g, b float64) Color {
	return Color{R: r, G: g, B: b, A: 1}
}

// ToArray converts color to float array.
func (c Color) ToArray() []float64 {
	return []float64{c.R, c.G, c.B, c.A}
}
