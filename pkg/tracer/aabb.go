package tracer

import (
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

// Type definition for AABB
type AABB struct {
	MinBound geom.Vec3
	MaxBound geom.Vec3
	Center   geom.Vec3
	Mat      Material
}

// NewAABB returns an Axis-Aligned Bounding Box given min and max bounds
func NewAABB(min, max geom.Vec3, mat Material) AABB {
	center := max.Plus(min).Scale(0.5)
	return AABB{MinBound: min, MaxBound: max, Center: center, Mat: mat}
}

// Hit checks if a Ray hit the sphere, returning
// t, p (point in ray) and n (surface normal)
func (aabb AABB) Hit(r geom.Ray, tMin, tMax float64) (t float64, surf Surface) {
	minB, maxB := aabb.MinBound, aabb.MaxBound
	n := minB.Minus(r.Orig).Div(r.Dir)
	f := maxB.Minus(r.Orig).Div(r.Dir)
	n, f = n.Min(f), n.Max(f)
	t0 := math.Max(math.Max(n.X(), n.Y()), n.Z())
	t1 := math.Min(math.Min(f.X(), f.Y()), f.Z())

	ltMin, ltMax := tMin, tMax
	if t0 > ltMin {
		ltMin = t0
	}
	if t1 < ltMax {
		ltMax = t1
	}

	if ltMax <= ltMin {
		return -1.0, aabb
	}
	return t0, aabb
}

func (a AABB) Material() (m Material) {
	return a.Mat
}

func (a AABB) Pos() (p geom.Vec3) {
	return a.Center
}

func (a AABB) Surface(p geom.Vec3) (n geom.Vec3, m Material) {
	eps := 1e-4
	switch {
	case p.X() < a.MinBound.X()+eps:
		n = geom.NewVec3(-1.0, 0, 0)
	case p.X() > a.MaxBound.X()-eps:
		n = geom.NewVec3(1.0, 0, 0)
	case p.Y() < a.MinBound.Y()+eps:
		n = geom.NewVec3(0, -1.0, 0)
	case p.Y() > a.MaxBound.Y()-eps:
		n = geom.NewVec3(0, 1.0, 0)
	case p.Z() < a.MinBound.Z()+eps:
		n = geom.NewVec3(0, 0, -1.0)
	case p.Z() > a.MaxBound.Z()-eps:
		n = geom.NewVec3(0, 0, 1.0)
	}
	return n, a.Mat
}
