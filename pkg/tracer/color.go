package tracer

import (
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

// Type definition for Color
type Color struct {
	geom.Vec3
}

// Creates a new Color, given r, g and b values
func NewColor(r, g, b float64) (c Color) {
	c.E[0] = r
	c.E[1] = g
	c.E[2] = b
	return
}

// R returns the red element of Color
func (c Color) R() float64 {
	return c.E[0]
}

// G returns the red element of Color
func (c Color) G() float64 {
	return c.E[1]
}

// B returns the red element of Color
func (c Color) B() float64 {
	return c.E[2]
}

// Plus returns the result of the sum of two Colors
func (c Color) Plus(c2 Color) Color {
	return Color{Vec3: c.Vec3.Plus(c2.Vec3)}
}

// Times returns the multiplication of two Colors
func (c Color) Times(c2 Color) Color {
	return Color{Vec3: c.Vec3.Times(c2.Vec3)}
}

// Scale returns an Color scaled by a factor n
func (c Color) Scale(n float64) Color {
	return Color{Vec3: c.Vec3.Scale(n)}
}

func (c Color) Clamp() Color {
	return NewColor(math.Min(1.0, c.R()), math.Min(1.0, c.G()), math.Min(1.0, c.B()))
}

// Gamma raises each of R, G, and B to 1/n
func (c Color) Gamma(n float64) Color {
	ni := 1 / n
	return NewColor(
		math.Pow(c.R(), ni),
		math.Pow(c.G(), ni),
		math.Pow(c.B(), ni),
	)
}
