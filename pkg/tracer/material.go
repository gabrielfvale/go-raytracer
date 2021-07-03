package tracer

import (
	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

// Material represents a material that produces a scattered ray
type Material interface {
	Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color)
}

// Lambertian describes a material attenuated by an albedo
type Lambertian struct {
	Albedo Color
}

// NewLambert returns a lambertian (diffuse) material
func NewLambert(albedo Color) Lambertian {
	return Lambertian{Albedo: albedo}
}

// Scatter scatters an incoming ray in a lambertian pattern
func (lm Lambertian) Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color) {
	scatterDir := n.Unit().Plus(geom.SampleHemisphereCos())

	if scatterDir.NearZero() {
		scatterDir = n
	}
	out = geom.NewRay(p, scatterDir)
	return true, out, lm.Albedo
}

// Metal describes a material that reflects light
type Metal struct {
	Albedo Color
	Rough  float64
}

// NewMetal returns a metalic (specular) material
func NewMetal(albedo Color, roughness float64) Metal {
	clampedRoughness := roughness
	if clampedRoughness > 1 {
		clampedRoughness = 1
	}
	return Metal{Albedo: albedo, Rough: roughness}
}

// Scatter scatters an incoming ray in a metal pattern
func (m Metal) Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color) {
	reflected := in.Dir.Reflect(n)
	// Add roughness/fuzzyness
	reflected = reflected.Plus(geom.SampleHemisphereCos().Scale(m.Rough))
	out = geom.NewRay(p, reflected)
	return reflected.Dot(n) > 0, out, m.Albedo
}

// Dielectric describes a material that refracts light
type Dielectric struct {
	RefrIndex float64
}

// NewDielectric returns a dielectric material
func NewDielectric(index float64) Dielectric {
	return Dielectric{RefrIndex: index}
}

// Scatter scatters an incoming ray in a dielectric pattern
func (m Dielectric) Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color) {
	attenuation = NewColor(1.0, 1.0, 1.0)
	etai, etat := 1.0, m.RefrIndex
	refrRatio := etai / etat
	rayDir := geom.NewVec3(0.0, 0.0, 0.0)

	if refracts, refracted := in.Dir.Refract(n, refrRatio); refracts {
		rayDir = refracted
	} else {
		rayDir = in.Dir.Reflect(n)
	}

	out = geom.NewRay(p, rayDir)
	return true, out, attenuation
}
