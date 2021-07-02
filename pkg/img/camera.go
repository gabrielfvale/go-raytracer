package img

import "github.com/gabrielfvale/go-raytracer/pkg/geom"

// Constant camera variables
const (
	aspectRatio    float64 = 16.0 / 9.0
	viewportHeight float64 = 2.0
	viewportWidth  float64 = aspectRatio * viewportHeight
	focalLength    float64 = 1.0
)

// Camera vectors
var (
	origin     = geom.NewVec3(0.0, 0.0, 0.0)
	horizontal = geom.NewVec3(viewportWidth, 0, 0)
	vertical   = geom.NewVec3(0, viewportHeight, 0)
	focalVec   = geom.NewVec3(0, 0, focalLength)
	lowerLeft  = origin.Minus(horizontal.Scale(0.5)).Minus(vertical.Scale(0.5)).Minus(focalVec)
)

// Type definition for Camera
type Camera struct {
}

// Ray returns a new Ray using the camera, given
// u, v coordinates.
func (c Camera) Ray(u, v float64) geom.Ray {
	return geom.NewRay(
		origin,
		lowerLeft.Plus(horizontal.Scale(u)).Plus(vertical.Scale(v)).Minus(origin).Unit(),
	)
}
