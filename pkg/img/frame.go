package img

import (
	"math"

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

// Render loops over the width and height and sets the pixels
func (f Frame) Render(pixels []byte, pitch int, s obj.Surface) {
	// Camera
	viewportHeight := 2.0
	viewportWidth := f.AR * viewportHeight
	focalLength := 1.0

	origin := geom.NewVec3(0.0, 0.0, 0.0)
	horizontal := geom.NewVec3(viewportWidth, 0, 0)
	vertical := geom.NewVec3(0, viewportHeight, 0)
	focalVec := geom.NewVec3(0, 0, focalLength)
	lowerLeft := origin.Minus(horizontal.Scale(0.5)).Minus(vertical.Scale(0.5)).Minus(focalVec)

	bpp := pitch / f.W // bytes-per-pixel
	for j := f.H - 1; j >= 0; j-- {
		for i := 0; i < f.W; i++ {
			ind := (j * pitch) + (i * bpp)

			u := float64(i) / float64(f.W-1)
			v := float64(j) / float64(f.H-1)

			r := geom.NewRay(
				origin,
				lowerLeft.Plus(horizontal.Scale(u)).Plus(vertical.Scale(v)).Minus(origin).Unit(),
			)
			pixelColor := color(r, s)
			f.WriteColor(ind, pixels, pixelColor)
		}
	}
}

func color(r geom.Ray, s obj.Surface) RGB {
	if t, _, n := s.Hit(r, 0, math.MaxFloat64); t > 0 {
		return NewRGB(n.X()+1.0, n.Y()+1.0, n.Z()+1.0).Scale(0.5)
	}
	t := 0.5 * (r.Dir.Y() + 1.0)
	c1 := NewRGB(1.0, 1.0, 1.0).Scale(1.0 - t)
	c2 := NewRGB(0.5, 0.7, 1.0).Scale(t)
	return c1.Plus(c2)
}
