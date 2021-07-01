package img

import (
	"math"
	"math/rand"

	"github.com/gabrielfvale/go-traytracer/pkg/geom"
	"github.com/gabrielfvale/go-traytracer/pkg/obj"
)

// Type definition for Frame
type Frame struct {
	W, H int
	AR   float64
}

// NewFrame returns a Frame, given width, height and aspect ratio
func NewFrame(width, height int, aspect float64) Frame {
	return Frame{W: width, H: height, AR: aspect}
}

// WriteColor writes an RGB color to a pixel byte array
func (f Frame) WriteColor(index int, pixels []byte, c RGB) {
	r := uint8(255.99 * c.R())
	g := uint8(255.99 * c.G())
	b := uint8(255.99 * c.B())
	pixels[index] = b
	pixels[index+1] = g
	pixels[index+2] = r
}

// Render loops over the width and height, and for each sample
// taking the average of the samples and setting the R, G, B
// values in a pixel byte array.
func (f Frame) Render(pixels []byte, pitch int, h obj.Hitable, samples int) {
	cam := Camera{}

	bpp := pitch / f.W // bytes-per-pixel
	for j := f.H - 1; j >= 0; j-- {
		for i := 0; i < f.W; i++ {
			ind := (j * pitch) + (i * bpp)
			c := NewRGB(0.0, 0.0, 0.0)
			for s := 0; s < samples; s++ {
				u := (float64(i) + rand.Float64()) / float64(f.W-1)
				v := (float64(j) + rand.Float64()) / float64(f.H-1)
				r := cam.Ray(u, v)
				c = c.Plus(color(r, h))
			}
			c = c.Scale(1 / float64(samples))
			f.WriteColor(ind, pixels, c)
		}
	}
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func color(r geom.Ray, h obj.Hitable) RGB {
	if t, _, n := h.Hit(r, 0, math.MaxFloat64); t > 0 {
		return NewRGB(n.X()+1.0, n.Y()+1.0, n.Z()+1.0).Scale(0.5)
	}
	t := 0.5 * (r.Dir.Y() + 1.0)
	c1 := NewRGB(1.0, 1.0, 1.0).Scale(1.0 - t)
	c2 := NewRGB(0.5, 0.7, 1.0).Scale(t)
	return c1.Plus(c2)
}
