package obj

import (
	"math"

	"github.com/gabrielfvale/go-traytracer/pkg/geom"
)

// Type definition for Sphere
type Sphere struct {
	Center geom.Vec3
	Radius float64
}

// NewSphere returns a Sphere given center and radius
func NewSphere(center geom.Vec3, radius float64) Sphere {
	return Sphere{Center: center, Radius: radius}
}

// Hit checks if a Ray hit the sphere, returning
// t, p (point in ray) and n (surface normal)
func (s Sphere) Hit(r geom.Ray, tMin, tMax float64) (t float64, p geom.Vec3, n geom.Vec3) {
	oc := r.Orig.Minus(s.Center)
	a := r.Dir.LenSq()
	halfB := oc.Dot(r.Dir)
	c := oc.LenSq() - s.Radius*s.Radius
	disc := halfB*halfB - a*c

	if disc < 0.0 {
		return -1.0, p, n
	}

	sqrtd := math.Sqrt(disc)
	// Test both roots
	t = (-halfB - sqrtd) / a
	if t > tMin && t < tMax {
		p = r.At(t)
		return t, p, p.Minus(s.Center).Unit()
	}
	t = (-halfB + sqrtd) / a
	if t > tMin && t < tMax {
		p = r.At(t)
		return t, p, p.Minus(s.Center).Unit()
	}

	return -1.0, p, n
}
