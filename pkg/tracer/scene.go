package tracer

import (
	"fmt"
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
	W, H      int
	Cam       Camera
	Objects   []Hitable
	Lights    []Hitable
	lightArea float64
	pmap      *PhotonMap
}

type result struct {
	row    int
	pixels []byte
}

// NewScene returns a Scene, given width, height and object slice
func NewScene(width, height int, cam Camera, objects []Hitable, pmap *PhotonMap) Scene {
	var lights []Hitable
	var lightArea float64 = 0.0
	// pre compute lights
	for _, o := range objects {
		m := o.Material()
		if m.Emittance > 0 {
			lights = append(lights, o)
			e := o.Material().Color
			lightArea += e.R() + e.G() + e.B()
		}
	}
	return Scene{W: width, H: height, Cam: cam, Objects: objects, Lights: lights, lightArea: lightArea, pmap: pmap}
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

func randSample(n geom.Vec3) geom.Vec3 {
	/*
		Vec nl(0,-1,0);
		float r1=2*M_PI*myrand(), r2=myrand(), r2s=sqrt(r2);
		Vec w=nl, u=((fabs(w.x)>.1?Vec(0,1,0):Vec(1,0,0))%w).norm(), v=w%u;
		Vec d = (u*cos(r1)*r2s + v*sin(r1)*r2s + w*sqrt(1-r2)).norm();
	*/
	r1 := 2 * math.Pi * rand.Float64()
	r2 := rand.Float64()
	r2s := math.Sqrt(r2)
	w := n
	u := geom.NewVec3(1, 0, 0)
	if math.Abs(w.X()) > 0.1 {
		u = geom.NewVec3(0, 1, 0)
	}
	u = u.Cross(w).Unit()
	v := w.Cross(u)

	uc := u.Scale(math.Cos(r1) * r2s)
	vc := v.Scale(math.Sin(r1) * r2s)
	wc := w.Scale(math.Sqrt(1 - r2))
	return uc.Plus(vc).Plus(wc).Unit()
	// return u.Scale(math.Cos(r1) * r2s).Plus(v.Scale(math.Sin(r1) * r2s)).Plus(w.Scale(math.Sqrt(1 - r2))).Unit()
}

// Render loops over the width and height, and for each sample
// taking the average of the samples and setting the R, G, B
// values in a pixel byte array.
func (scene Scene) Render(pixels []byte, pitch int, samples int) {
	rnd1 := rand.New(rand.NewSource(time.Now().Unix()))
	pmap := scene.pmap
	// pbar := util.NewProgress(0, pmap.maxPhotons)
	fmt.Println("tracing photons")
	for _, l := range scene.Lights {
		e := l.Material().Color
		area := e.R() + e.G() + e.B()
		for pmap.storedPhotons < pmap.maxPhotons*int(scene.lightArea/area) {
			pos := l.Pos()
			col := NewColor(10.0, 10.0, 10.0)
			nl := geom.NewVec3(0, -1, 0)
			rp := geom.NewRay(pos, randSample(nl))
			scene.tracePhotons(rp, 4, col, rnd1)
		}
	}

	pmap.Balance()
	pmap.ScalePhotonPower(1.0 / float64(pmap.maxPhotons))

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

func (scene Scene) intersect(r geom.Ray) (hit bool, t float64, s Surface) {
	tMin, tMax := bias, math.MaxFloat64
	t = tMax
	hit = false
	for _, o := range scene.Objects {
		if ht, hs := o.Hit(r, tMin, t); ht > 0.0 {
			hit = true
			t = ht
			s = hs
		}
	}
	return
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func (scene Scene) trace(r geom.Ray, depth int, rnd *rand.Rand) Color {
	if depth <= 0 {
		return NewColor(0.0, 0.0, 0.0)
	}

	hit, tNear, surf := scene.intersect(r)

	if !hit {
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
	orientedN := n
	if n.Dot(incident) >= 0.0 {
		orientedN = orientedN.Inv()
	}
	// n = n.Unit()

	irrad := scene.pmap.IrradianceEst(p.E, orientedN.E, 10, 100)
	// fmt.Println(irrad)
	return NewColor(irrad[0], irrad[1], irrad[2])
	result = result.Plus(NewColor(irrad[0], irrad[1], irrad[2]))

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
		irrad := scene.pmap.IrradianceEst(p.E, orientedN.E, 5, 100)
		result = result.Plus(NewColor(irrad[0], irrad[1], irrad[2]))
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
						visible = 0.0
					}
					tNear = ht
				}
			}
			result = result.Plus(m.Color.Scale(fd).Scale(visible))
		}
	}
	return result
}

// Color checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a background gradient
func (scene Scene) tracePhotons(r geom.Ray, depth int, power Color, rnd *rand.Rand) {
	// fmt.Println(depth)
	if depth <= 0 {
		return
	}

	hit, tNear, surf := scene.intersect(r)

	if !hit {
		return
	}

	incident := r.Dir.Unit()
	p := r.At(tNear)
	n, m := surf.Surface(p)
	// Properly oriented normal
	orientedN := n
	if n.Dot(incident) >= 0.0 {
		orientedN = orientedN.Inv()
	}
	// BRDF modulator
	f := m.Color
	// Maximum reflectivity for russian roulette
	// rrp := math.Max(math.Max(f.R(), f.G()), f.B())
	rrp := (f.R() + f.G() + f.B()) / 3

	// "Normal" material
	if m.Normal || m.Emittance > 0 {
		return
	}

	if m.Lambert { // Lambertian material

	} else if m.Reflectivity > 0 { // Metalic material
		reflected := incident.Reflect(n)
		// Add roughness/fuzzyness
		reflected = reflected.Plus(randSample(orientedN).Scale(m.Roughness))
		r2 := geom.NewRay(p, reflected)
		scene.tracePhotons(r2, depth-1, f.Times(power), rnd)
	} else if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		scene.tracePhotons(r2, depth-1, power, rnd)
	} else {

		if rnd.Float64() < rrp { // absorb photon
			// fmt.Println("absorb photon", depth)
			att := f.Times(power).Scale(1.0 / (1.0 - rrp))
			scene.pmap.Store(att.E, p.E, incident.E)
		} else { // trace another ray
			// Random ray
			// att := f.Times(power).Scale(1.0 / rrp)
			r2 := geom.NewRay(p, randSample(orientedN))
			scene.tracePhotons(r2, depth-1, f.Times(power), rnd)
		}
	}
}
