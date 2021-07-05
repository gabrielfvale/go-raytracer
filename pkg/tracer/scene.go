package tracer

import (
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"github.com/gabrielfvale/go-raytracer/pkg/util"
)

const bias = 0.001

// Type definition for Scene
type Scene struct {
	W, H    int
	Cam     Camera
	Objects []Hitable
	Lights  []Hitable
}

type result struct {
	row    int
	pixels []byte
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
	worker := func(jobs <-chan int, results chan<- result, rnd *rand.Rand) {
		for y := range jobs {
			res := result{row: y, pixels: make([]byte, bpp*scene.W)}
			for x := 0; x < scene.W; x++ {
				ind := (x * bpp)
				c := NewColor(0.0, 0.0, 0.0)
				for s := 0; s < samples; s++ {
					u := (float64(x) + rnd.Float64()) / float64(scene.W)
					v := (float64(y) + rnd.Float64()) / float64(scene.H)
					r := scene.Cam.Ray(u, v)
					c = c.Plus(scene.trace(r, 5, rnd))
				}
				c = c.Scale(1 / float64(samples)).Gamma(2)
				c = c.Clamp()
				scene.WriteColor(ind, res.pixels, c)
			}
			results <- res
		}
	}

	workers := runtime.NumCPU() + 1
	jobs := make(chan int, scene.H)
	results := make(chan result, workers+1)
	pending := make(map[int][]byte, 0)
	cursor := 0
	bar := util.NewProgress(0, scene.H)

	for w := 0; w < workers; w++ {
		go worker(jobs, results, rand.New(rand.NewSource(time.Now().Unix())))
	}
	for y := 0; y < scene.H; y++ {
		jobs <- y
	}

	close(jobs)

	for y := 0; y < scene.H; y++ {
		r := <-results
		bar.Tick()
		pending[r.row] = r.pixels
		for len(pending[cursor]) > 0 {
			copy(pixels[cursor*pitch:], pending[cursor])
			delete(pending, cursor)
			cursor++
		}
	}
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func (scene Scene) trace(r geom.Ray, depth int, rnd *rand.Rand) Color {
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
		// t := 0.5 * (r.Dir.Y() + 1.0)
		// c1 := NewColor(1.0, 1.0, 1.0).Scale(1.0 - t)
		// c2 := NewColor(0.5, 0.7, 1.0).Scale(t)
		// return c1.Plus(c2)
		return NewColor(0, 0, 0)
	}

	result := NewColor(0.0, 0.0, 0.0)
	incident := r.Dir.Unit()
	p := r.At(tNear)
	n, m := surf.Surface(p)
	// n = n.Unit()

	// "Normal" material
	if m.Normal {
		return NewColor(n.X()+0.5, n.Y()+0.5, n.Z()+0.5).Scale(0.5)
	}

	if m.Emittance > 0 {
		result = result.Plus(m.Color.Scale(m.Emittance))
	} else if m.Lambert { // Lambertian material
		scattered := n.Unit().Plus(geom.SampleHemisphere(rnd))
		if scattered.NearZero() {
			scattered = n
		}
		r2 := geom.NewRay(p, scattered)
		result = result.Plus(scene.trace(r2, depth-1, rnd).Times(m.Color))
	} else if m.Reflectivity > 0 { // Metalic material
		reflected := incident.Reflect(n)
		// Add roughness/fuzzyness
		reflected = reflected.Plus(geom.SampleHemisphereCos(rnd).Scale(m.Roughness))
		if reflected.Dot(n) > 0 {
			r2 := geom.NewRay(p, reflected)
			result = result.Plus(scene.trace(r2, depth-1, rnd).Times(m.Color).Scale(m.Reflectivity))
		}
	} else if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		result = result.Plus(scene.trace(r2, depth-1, rnd))
	} else {
		// calc diffuse
		for _, l := range scene.Lights {
			pos := l.Pos()
			dir := pos.Minus(p).Unit()
			fd := n.Dot(dir)
			if fd < 0 {
				fd = 0
			}
			// calculate shadow
			visible := 1.0
			tMin, tMax := bias, math.MaxFloat64
			tNear := tMax
			shadowRay := geom.NewRay(p, dir)
			for _, o := range scene.Objects {
				if ht, _ := o.Hit(shadowRay, tMin, tNear); ht > 0.0 {
					m := o.Material()
					if m.Emittance == 0 {
						visible = 0.5
					}
					tNear = ht
				}
			}
			result = result.Plus(m.Color.Scale(fd)).Scale(visible)
		}
	}
	return result

}
