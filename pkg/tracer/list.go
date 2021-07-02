package tracer

import "github.com/gabrielfvale/go-raytracer/pkg/geom"

// Type definition for List
type List struct {
	HL []Hitable
}

func NewList(hl ...Hitable) List {
	return List{HL: hl}
}

func (l List) Hit(r geom.Ray, tMin, tMax float64) (t float64, surf Surface) {
	closest := tMax
	for _, s := range l.HL {
		if ht, hs := s.Hit(r, tMin, closest); ht > 0.0 {
			closest, t = ht, ht
			surf = hs
		}
	}
	return
}
