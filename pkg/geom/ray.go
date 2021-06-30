package geom

// Type definition for Rat
type Ray struct {
	Orig Vec3
	Dir  Vec3
}

// NewRay returns a Ray given origin and direction Vec3
func NewRay(origin Vec3, direction Vec3) Ray {
	return Ray{Orig: origin, Dir: direction}
}

// At returns the point in the Ray given t
func (r Ray) At(t float64) Vec3 {
	return r.Orig.Plus(r.Dir.Scale(t))
}
