package obj

import "github.com/gabrielfvale/go-traytracer/pkg/geom"

type Hitable interface {
	Hit(r geom.Ray, tMin, tMax float64) (t float64, p geom.Vec3, n geom.Vec3)
}
