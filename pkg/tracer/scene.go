package tracer

import (
	"log"
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
	W, H        int
	Cam         Camera
	Objects     []Hitable
	tObjects    []Hitable
	Lights      []Hitable
	lightArea   float64
	globalPmap  *PhotonMap
	causticPmap *PhotonMap
	maxDepth    int
}

type result struct {
	row    int
	pixels []byte
}

// NewScene returns a Scene, given width, height and object slice
func NewScene(width, height int, cam Camera, objects []Hitable, globalPmap *PhotonMap, causticPmap *PhotonMap) Scene {
	var lights []Hitable
	var tobjects []Hitable
	var lightArea float64 = 0.0
	// pre compute lights and dieletric objects
	for _, o := range objects {
		m := o.Material()
		if m.Emittance > 0 {
			lights = append(lights, o)
			e := o.Material().Color
			lightArea += e.R() + e.G() + e.B()
		}
		// The dielectric objects slice is used for the caustics photon map.
		if m.Transparent {
			tobjects = append(tobjects, o)
		}
	}
	return Scene{
		W:           width,
		H:           height,
		Cam:         cam,
		Objects:     objects,
		tObjects:    tobjects,
		Lights:      lights,
		lightArea:   lightArea,
		globalPmap:  globalPmap,
		causticPmap: causticPmap,
		maxDepth:    4,
	}
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
	log.Printf("Started rendering (%d samples)", samples)
	start := time.Now()

	scene.mapPhotons()

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
					c = c.Plus(scene.trace(r, 1, rnd))
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

	log.Printf("Rendering scene")
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

	elapsed := time.Since(start)
	log.Printf("Rendering took %s", elapsed)
}

// mapPhotons takes the two photons maps from the scene
// and runs the photon mapping routines for each one
func (scene Scene) mapPhotons() {
	rnd1 := rand.New(rand.NewSource(time.Now().Unix()))
	global := scene.globalPmap
	caustics := scene.causticPmap

	log.Printf("Tracing photons")
	for _, l := range scene.Lights {
		e := l.Material().Color
		area := e.R() + e.G() + e.B()
		pos := l.Pos()
		nl := geom.NewVec3(0, -1, 0)
		log.Printf("Global photon mapping")
		for global.storedPhotons < global.maxPhotons*int(scene.lightArea/area) {
			rp := geom.NewRay(pos, geom.SampleHemisphereNormal(nl, rnd1))
			scene.tracePhotons(rp, 1, NewColor(15.0, 15.0, 15.0), global, false, rnd1)
		}
		log.Printf("Caustics photon mapping")
		for caustics.storedPhotons < caustics.maxPhotons*int(scene.lightArea/area) {
			rp := geom.NewRay(pos, geom.SampleHemisphereNormal(nl, rnd1))
			scene.tracePhotons(rp, 1, NewColor(1.0, 1.0, 1.0), caustics, true, rnd1)
		}
	}
	// Scale photon power
	global.ScalePhotonPower(1000.0 / float64(global.maxPhotons))
	caustics.ScalePhotonPower(1000.0 / float64(caustics.maxPhotons))
}

// Intersect loops over a list of Hitable, returning if there was a hit,
// the nearest t and the surface hit s
func (scene Scene) intersect(r geom.Ray, objs []Hitable) (hit bool, t float64, s Surface) {
	tMin, tMax := bias, math.MaxFloat64
	t = tMax
	hit = false
	for _, o := range objs {
		if ht, hs := o.Hit(r, tMin, t); ht > 0.0 {
			hit = true
			t = ht
			s = hs
		}
	}
	return
}

// Irradiance traces a ray, and estimates a color given a photon map.
func (scene Scene) irradiance(pmap *PhotonMap, r geom.Ray, depth int, rnd *rand.Rand) Color {

	black := NewColor(0.0, 0.0, 0.0)
	if depth >= scene.maxDepth {
		return black
	}

	hit, tNear, surf := scene.intersect(r, scene.Objects)

	if !hit {
		return black
	}

	incident := r.Dir.Unit()
	p := r.At(tNear)
	n, m := surf.Surface(p)
	orientedN := n
	if n.Dot(incident) >= 0.0 {
		orientedN = orientedN.Inv()
	}
	// BRDF modulator
	f := m.Color

	if m.Reflectivity > 0 { // Metalic material
		reflected := incident.Reflect(n)
		// Add roughness/fuzzyness
		reflected = reflected.Plus(geom.SampleHemisphereNormal(n, rnd).Scale(m.Roughness))
		if reflected.Dot(n) > 0 {
			r2 := geom.NewRay(p, reflected)
			return f.Times(scene.irradiance(pmap, r2, depth+1, rnd))
		}
	} else if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		return scene.irradiance(pmap, r2, depth+1, rnd)
	} else {
		// Material is diffuse
		// Direct visualization of photon map
		irradVec := pmap.IrradianceEst(p, n, 0, 100)
		return NewColor(irradVec.X(), irradVec.Y(), irradVec.Z())
	}
	return black
}

// trace checks if a ray intersects a list of objects,
// returning their color. If there is no hit,
// returns a black background
func (scene Scene) trace(r geom.Ray, depth int, rnd *rand.Rand) Color {
	if depth >= scene.maxDepth {
		return NewColor(0.0, 0.0, 0.0)
	}

	hit, tNear, surf := scene.intersect(r, scene.Objects)

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

	/* Debugging: Global map
	irrad := scene.globalPmap.IrradianceEst(p, n, 0, 50)
	result = result.Plus(NewColor(irrad.X(), irrad.Y(), irrad.Z()))
	return result
	*/

	/* Debugging: Caustics map
	irrad := scene.causticPmap.IrradianceEst(p, n, 0, 50)
	result = result.Plus(NewColor(irrad.X(), irrad.Y(), irrad.Z()))
	return result
	*/

	// "Normal" material
	if m.Normal {
		return NewColor(n.X()+0.5, n.Y()+0.5, n.Z()+0.5).Scale(0.5)
	}

	if m.Emittance > 0 {
		result = m.Color.Scale(m.Emittance)
	} else if m.Lambert { // Lambertian material
		scattered := geom.SampleHemisphereNormal(n, rnd)
		if scattered.NearZero() {
			scattered = n
		}
		r2 := geom.NewRay(p, scattered)
		result = result.Plus(scene.trace(r2, depth+1, rnd).Times(m.Color))
	} else if m.Reflectivity > 0 { // Metalic material
		reflected := incident.Reflect(n)
		// Add roughness/fuzzyness
		reflected = reflected.Plus(geom.SampleHemisphereNormal(n, rnd).Scale(m.Roughness))
		if reflected.Dot(n) > 0 {
			r2 := geom.NewRay(p, reflected)
			result = result.Plus(scene.trace(r2, depth+1, rnd).Times(m.Color).Scale(m.Reflectivity))
		}
	} else if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		result = result.Plus(scene.trace(r2, depth+1, rnd))
	} else {
		// Material is diffuse

		const BRDF float64 = 1 / math.Pi
		irrad := geom.NewVec3(0.0, 0.0, 0.0)

		/* Caustics */
		irrad = irrad.Plus(scene.causticPmap.IrradianceEst(p, n, 1, 100).Scale(BRDF))

		/* Global illumination
		irrad = irrad.Plus(scene.globalPmap.IrradianceEst(p, n, 0, 100).Scale(BRDF))
		*/

		irradColor := Color{Vec3: irrad}
		result = result.Plus(irradColor.Times(m.Color))

		/* Direct illumination */
		for _, l := range scene.Lights {
			pos := l.Pos()
			dir := pos.Minus(p).Unit()
			power := l.Material().Color
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
			result = result.Plus(m.Color.Scale(fd).Times(power).Scale(visible))
		}
	}
	return result
}

// tracePhotons traces photons emitted from a light source,
// storing them if the surface hit is diffuse, and bouncing
// them otherwise.
func (scene Scene) tracePhotons(r geom.Ray, depth int, power Color, pmap *PhotonMap, caustics bool, rnd *rand.Rand) {
	if depth >= scene.maxDepth {
		return
	}

	if caustics && depth == 1 {
		if hit, _, _ := scene.intersect(r, scene.tObjects); !hit {
			return
		}
	}

	hit, tNear, surf := scene.intersect(r, scene.Objects)

	if !hit {
		return
	}

	incident := r.Dir.Unit()
	p := r.At(tNear)
	n, m := surf.Surface(p)

	if caustics && depth == 1 && !m.Transparent {
		return
	}

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
		reflected = reflected.Plus(geom.SampleHemisphereNormal(orientedN, rnd).Scale(m.Roughness))
		r2 := geom.NewRay(p, reflected)
		scene.tracePhotons(r2, depth+1, f.Times(power), pmap, caustics, rnd)
	} else if m.Transparent { // Dielectric material
		etai, etat := 1.0, m.RefrIndex
		refrRatio := etai / etat

		refracts, rayDir := incident.Refract(n, refrRatio)
		if !refracts {
			rayDir = incident.Reflect(n)
		}
		r2 := geom.NewRay(p, rayDir)
		scene.tracePhotons(r2, depth+1, power, pmap, caustics, rnd)
	} else {
		if rnd.Float64() < rrp { // absorb photon
			// fmt.Println("absorb photon", depth)
			att := f.Times(power).Scale(1.0 / (1.0 - rrp))
			pmap.Store(att.E, p.E, incident.E)
		} else { // trace another ray
			// Random ray
			r2 := geom.NewRay(p, geom.SampleHemisphereNormal(n, rnd))
			scene.tracePhotons(r2, depth+1, f.Times(power).Scale(1.0/rrp), pmap, caustics, rnd)
		}
	}
}
