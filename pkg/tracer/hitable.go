package tracer

import "github.com/gabrielfvale/go-raytracer/pkg/geom"

type Surface interface {
	Surface(p geom.Vec3) (n geom.Vec3, m Material)
}

// Hitable represents an object that can be hit by a Ray
type Hitable interface {
	Hit(r geom.Ray, tMin, tMax float64) (t float64, s Surface)
	Material() (m Material)
	Pos() (p geom.Vec3)
}
