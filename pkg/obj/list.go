package obj

import "github.com/gabrielfvale/go-traytracer/pkg/geom"

// Type definition for List
type List struct {
	SS []Surface
}

func NewList(s ...Surface) List {
	return List{SS: s}
}

func (l List) Hit(r geom.Ray, tMin, tMax float64) (t float64, p geom.Vec3, n geom.Vec3) {
	closest := tMax
	for _, s := range l.SS {
		if st, sp, sn := s.Hit(r, tMin, closest); st > 0.0 {
			closest, t = st, st
			p = sp
			n = sn
		}
	}
	return
}
