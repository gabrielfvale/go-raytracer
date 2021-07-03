package tracer

import (
	"math"
	"math/rand"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

const bias = 0.001

// Type definition for Scene
type Scene struct {
	W, H    int
	Cam     Camera
	Objects []Hitable
	Lights  []Hitable
}

// NewScene returns a Scene, given width, height and object slice
func NewScene(width, height int, cam Camera, objects []Hitable) Scene {
	var lights []Hitable
	// pre compute lights
	for _, o := range objects {
		m := o.Material()
		if m.Emittance > 0 {
			lights = append(lights, o)
		}
	}
	return Scene{W: width, H: height, Cam: cam, Objects: objects, Lights: lights}
}

// WriteColor writes a Color to a pixel byte array
func (scene Scene) WriteColor(index int, pixels []byte, c Color) {
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
func (scene Scene) Render(pixels []byte, pitch int, samples int) {
	bpp := pitch / scene.W // bytes-per-pixel
	for j := scene.H - 1; j >= 0; j-- {
		for i := 0; i < scene.W; i++ {
			ind := (j * pitch) + (i * bpp)
			c := NewColor(0.0, 0.0, 0.0)
			for s := 0; s < samples; s++ {
				u := (float64(i) + rand.Float64()) / float64(scene.W-1)
				v := (float64(j) + rand.Float64()) / float64(scene.H-1)
				r := scene.Cam.Ray(u, v)
				c = c.Plus(scene.trace(r, 50))
			}
			c = c.Scale(1 / float64(samples)).Gamma(2)
			scene.WriteColor(ind, pixels, c)
		}
	}
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func (scene Scene) trace(r geom.Ray, depth int) Color {
	if depth <= 0 {
		return NewColor(0.0, 0.0, 0.0)
	}

	tMin, tMax := bias, math.MaxFloat64
	tNear := tMax
	var surf Surface
	hasHit := false
	for _, s := range scene.Objects {
		if ht, hs := s.Hit(r, tMin, tNear); ht > 0.0 {
			hasHit = true
			tNear = ht
			surf = hs
		}
	}

	if !hasHit {
		t := 0.5 * (r.Dir.Y() + 1.0)
		c1 := NewColor(1.0, 1.0, 1.0).Scale(1.0 - t)
		c2 := NewColor(0.5, 0.7, 1.0).Scale(t)
		return c1.Plus(c2)
	}

	result := NewColor(0.0, 0.0, 0.0)
	incident := r.Dir.Unit()
	p := r.At(tNear)
	n, m := surf.Surface(p)

	if m.Emittance > 0 {
		result = result.Plus(m.Color.Scale(m.Emittance))
	}
	if m.Lambert { // Lambertian material
		scattered := n.Unit().Plus(geom.SampleHemisphereCos())
		if scattered.NearZero() {
			scattered = n
		}
		r2 := geom.NewRay(p, scattered)
		result = result.Plus(scene.trace(r2, depth-1).Times(m.Color))
	}
	if m.Reflectivity > 0 { // Metalic material
		reflected := incident.Reflect(n)
		// Add roughness/fuzzyness
		reflected = reflected.Plus(geom.SampleHemisphereCos().Scale(m.Roughness))
		if reflected.Dot(n) > 0 {
			r2 := geom.NewRay(p, reflected)
			result = result.Plus(scene.trace(r2, depth-1).Times(m.Color).Scale(m.Reflectivity))
		}
	}
	if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		result = result.Plus(scene.trace(r2, depth-1))
	}
	// calc diffuse
	return result

}
