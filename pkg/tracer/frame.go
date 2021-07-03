package tracer

import (
	"math"
	"math/rand"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

const bias = 0.001

// Type definition for Frame
type Frame struct {
	W, H int
	AR   float64
}

// NewFrame returns a Frame, given width, height and aspect ratio
func NewFrame(width, height int, aspect float64) Frame {
	return Frame{W: width, H: height, AR: aspect}
}

// WriteColor writes a Color to a pixel byte array
func (f Frame) WriteColor(index int, pixels []byte, c Color) {
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
func (f Frame) Render(pixels []byte, pitch int, h Hitable, samples int) {
	cam := NewCamera(geom.NewVec3(-2, 2, 1), geom.NewVec3(0, 0, -1), geom.NewVec3(0, 1, 0), 90, f.AR)

	bpp := pitch / f.W // bytes-per-pixel
	for j := f.H - 1; j >= 0; j-- {
		for i := 0; i < f.W; i++ {
			ind := (j * pitch) + (i * bpp)
			c := NewColor(0.0, 0.0, 0.0)
			for s := 0; s < samples; s++ {
				u := (float64(i) + rand.Float64()) / float64(f.W-1)
				v := (float64(j) + rand.Float64()) / float64(f.H-1)
				r := cam.Ray(u, v)
				c = c.Plus(color(r, h, 50))
			}
			c = c.Scale(1 / float64(samples)).Gamma(2)
			f.WriteColor(ind, pixels, c)
		}
	}
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func color(r geom.Ray, h Hitable, depth int) Color {
	if depth <= 0 {
		return NewColor(0.0, 0.0, 0.0)
	}
	if t, s := h.Hit(r, bias, math.MaxFloat64); t > 0 {
		p := r.At(t)
		n, m := s.Surface(p)
		if scattered, outRay, attenuation := m.Scatter(r, p, n); scattered {
			return color(outRay, h, depth-1).Times(attenuation)
		}
		return NewColor(0.0, 0.0, 0.0)
	}
	t := 0.5 * (r.Dir.Y() + 1.0)
	c1 := NewColor(1.0, 1.0, 1.0).Scale(1.0 - t)
	c2 := NewColor(0.5, 0.7, 1.0).Scale(t)
	return c1.Plus(c2)
}
