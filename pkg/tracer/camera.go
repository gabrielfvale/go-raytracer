package tracer

import (
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
)

// Type definition for Camera
type Camera struct {
	origin, horizontal,
	vertical, lowerLeft geom.Vec3
}

// NewCamera creates a Camera given 3 vectors, fov and aspect ratio
func NewCamera(eye, lookat, vup geom.Vec3, vfov, aspect float64) (c Camera) {
	theta := vfov * math.Pi / 180
	halfH := math.Tan(theta / 2)
	halfW := aspect * halfH

	w := eye.Minus(lookat).Unit()
	u := vup.Cross(w).Unit()
	v := u.Cross(w).Unit()

	c.origin = eye
	c.lowerLeft = c.origin.Minus(u.Scale(halfW)).Minus(v.Scale(halfH)).Minus(w)
	c.horizontal = u.Scale(2 * halfW)
	c.vertical = v.Scale(2 * halfH)
	return
}

// Ray returns a new Ray using the camera, given
// u, v coordinates.
func (c Camera) Ray(u, v float64) geom.Ray {
	return geom.NewRay(
		c.origin,
		c.lowerLeft.Plus(c.horizontal.Scale(u)).Plus(c.vertical.Scale(v)).Minus(c.origin),
	)
}
