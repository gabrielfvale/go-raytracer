package obj

import "github.com/gabrielfvale/go-raytracer/pkg/geom"

// Type definition for List
type List struct {
	HL []Hitable
}

func NewList(hl ...Hitable) List {
	return List{HL: hl}
}

func (l List) Hit(r geom.Ray, tMin, tMax float64) (t float64, p geom.Vec3, n geom.Vec3) {
	closest := tMax
	for _, s := range l.HL {
		if st, sp, sn := s.Hit(r, tMin, closest); st > 0.0 {
			closest, t = st, st
			p = sp
			n = sn
		}
	}
	return
}
