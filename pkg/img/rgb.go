package img

import "github.com/gabrielfvale/go-traytracer/pkg/linalg"

// Type definition for RGB
type RGB struct {
	linalg.Vec3
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
