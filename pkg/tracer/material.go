package tracer

import "github.com/gabrielfvale/go-raytracer/pkg/geom"

// Material represents a material that produces a scattered ray
type Material interface {
	Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color)
}

// Lambertian describes a material attenuated by an albedo
type Lambertian struct {
	Albedo Color
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
}

// Scatter scatters an incoming ray in a metal pattern
func (m Metal) Scatter(in geom.Ray, p geom.Vec3, n geom.Vec3) (scattered bool, out geom.Ray, attenuation Color) {
	reflected := in.Dir.Reflect(n)
	out = geom.NewRay(p, reflected)
	return out.Dir.Dot(n) > 0, out, m.Albedo
}
