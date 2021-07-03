package tracer

import (
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

// Type definition for Sphere
type Sphere struct {
	Center geom.Vec3
	Radius float64
	Mat    Material
}

// NewSphere returns a Sphere given center and radius
func NewSphere(center geom.Vec3, radius float64, mat Material) Sphere {
	return Sphere{Center: center, Radius: radius, Mat: mat}
}

// Hit checks if a Ray hit the sphere, returning
// t, p (point in ray) and n (surface normal)
func (s Sphere) Hit(r geom.Ray, tMin, tMax float64) (t float64, surf Surface) {
	oc := r.Orig.Minus(s.Center)
	a := r.Dir.LenSq()
	halfB := oc.Dot(r.Dir)
	c := oc.LenSq() - s.Radius*s.Radius
	disc := halfB*halfB - a*c

	if disc < 0.0 {
		return -1.0, s
	}

	sqrtd := math.Sqrt(disc)
	// Test both roots
	t = (-halfB - sqrtd) / a
	if t > tMin && t < tMax {
		return t, s
	}
	t = (-halfB + sqrtd) / a
	if t > tMin && t < tMax {
		return t, s
	}

	return -1.0, s
}

func (s Sphere) Material() (m Material) {
	return s.Mat
}

func (s Sphere) Surface(p geom.Vec3) (n geom.Vec3, m Material) {
	return p.Minus(s.Center).Scale(s.Radius).Unit(), s.Mat
}
