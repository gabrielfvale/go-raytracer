package img

import "github.com/gabrielfvale/go-traytracer/pkg/geom"

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

func (c RGB) Plus(c2 RGB) RGB {
	return RGB{Vec3: c.Vec3.Plus(c2.Vec3)}
}

func (c RGB) Scale(n float64) RGB {
	return RGB{Vec3: c.Vec3.Scale(n)}
}

func WriteColor(index int, pixels []byte, c RGB) {
	r := uint8(255.99 * c.R())
	g := uint8(255.99 * c.G())
	b := uint8(255.99 * c.B())
	pixels[index] = b
	pixels[index+1] = g
	pixels[index+2] = r
}
