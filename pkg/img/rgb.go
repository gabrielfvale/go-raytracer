package img

import (
	"math"

	"github.com/gabrielfvale/go-traytracer/pkg/geom"
)

// Type definition for RGB
type RGB struct {
	geom.Vec3
}

// Creates a new RGB, given r, g and b values
func NewRGB(r, g, b float64) (c RGB) {
	c.E[0] = r
	c.E[1] = g
	c.E[2] = b
	return
}

// R returns the red element of RGB
func (c RGB) R() float64 {
	return c.E[0]
}

// G returns the red element of RGB
func (c RGB) G() float64 {
	return c.E[1]
}

// B returns the red element of RGB
func (c RGB) B() float64 {
	return c.E[2]
}

// Plus returns the result of the sum of two RGBs
func (c RGB) Plus(c2 RGB) RGB {
	return RGB{Vec3: c.Vec3.Plus(c2.Vec3)}
}

// Scale returns an RGB scaled by a factor n
func (c RGB) Scale(n float64) RGB {
	return RGB{Vec3: c.Vec3.Scale(n)}
}

// Gamma raises each of R, G, and B to 1/n
func (c RGB) Gamma(n float64) RGB {
	ni := 1 / n
	return NewRGB(
		math.Pow(c.R(), ni),
		math.Pow(c.G(), ni),
		math.Pow(c.B(), ni),
	)
}
