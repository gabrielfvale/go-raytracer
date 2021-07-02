package geom

import (
	"fmt"
	"io"
	"math"
	"math/rand"
)

// Type definition for Vec3
type Vec3 struct {
	E [3]float64
}

// Creates a new Vec3, given x, y and z values
func NewVec3(x, y, z float64) (v Vec3) {
	v.E[0] = x
	v.E[1] = y
	v.E[2] = z
	return
}

// X returns the X coordinate of a given Vec3 v
func (v Vec3) X() float64 {
	return v.E[0]
}

// Y returns the Y coordinate of a given Vec3 v
func (v Vec3) Y() float64 {
	return v.E[1]
}

// Z returns the Z coordinate of a given Vec3 v
func (v Vec3) Z() float64 {
	return v.E[2]
}

// Inv returns an inverted Vec3 v
func (v Vec3) Inv() Vec3 {
	return NewVec3(-v.X(), -v.Y(), -v.Z())
}

// Plus returns the sum of two Vec3
func (v Vec3) Plus(v2 Vec3) Vec3 {
	return NewVec3(v.X()+v2.X(), v.Y()+v2.Y(), v.Z()+v2.Z())
}

// Minus returns the difference of two Vec3
func (v Vec3) Minus(v2 Vec3) Vec3 {
	return NewVec3(v.X()-v2.X(), v.Y()-v2.Y(), v.Z()-v2.Z())
}

// Times returns the multiplication of two Vec3
func (v Vec3) Times(v2 Vec3) Vec3 {
	return NewVec3(v.X()*v2.X(), v.Y()*v2.Y(), v.Z()*v2.Z())
}

// Div returns the division of two Vec3
func (v Vec3) Div(v2 Vec3) Vec3 {
	return NewVec3(v.X()/v2.X(), v.Y()/v2.Y(), v.Z()/v2.Z())
}

// LenSq returns the squared length of a Vec3
func (v Vec3) LenSq() float64 {
	return v.X()*v.X() + v.Y()*v.Y() + v.Z()*v.Z()
}

// Len returns the length of a Vec3
func (v Vec3) Len() float64 {
	return math.Sqrt(v.LenSq())
}

// Scale returns a Vec3 scaled by a float64 value
func (v Vec3) Scale(n float64) Vec3 {
	return NewVec3(v.X()*n, v.Y()*n, v.Z()*n)
}

// Dot returns the dot product of two Vec3
func (v Vec3) Dot(v2 Vec3) float64 {
	return v.X()*v2.X() + v.Y()*v2.Y() + v.Z()*v2.Z()
}

// Cross returns the cross product of two Vec3
func (v Vec3) Cross(v2 Vec3) Vec3 {
	newX := v.Y()*v2.Z() - v.Z()*v2.Y()
	newY := v.Z()*v2.X() - v.X()*v2.Z()
	newZ := v.X()*v2.Y() - v.Y()*v2.X()
	return NewVec3(newX, newY, newZ)
}

// Unit returns the unit Vec3 of v
func (v Vec3) Unit() Vec3 {
	k := 1.0 / v.Len()
	return v.Scale(k)
}

// NearZero returns if a Vec3 is close to 0
func (v Vec3) NearZero() bool {
	s := 1e-8
	return math.Abs(v.X()) < s && math.Abs(v.Y()) < s && math.Abs(v.Z()) < s
}

// Reflect returns a reflected Vec3 in relation to a normal n
func (v Vec3) Reflect(n Vec3) Vec3 {
	return v.Minus(n.Scale(2 * v.Dot(n)))
}

// SampleSphere returns a random unit vector in a sphere
func SampleSphere() Vec3 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	x := math.Cos(2*math.Pi*u2) * 2 * math.Sqrt(u1*(1.0-u1))
	y := math.Sin(2*math.Pi*u2) * 2 * math.Sqrt(u1*(1.0-u1))
	z := 1.0 - 2.0*u1
	return NewVec3(x, y, z).Unit()
}

// SampleHemisphere returns a random unit vector in a hemisphere
func SampleHemisphere() Vec3 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	x := math.Cos(2*math.Pi*u2) * 2 * math.Sqrt(1.0-u1*u1)
	y := math.Sin(2*math.Pi*u2) * 2 * math.Sqrt(1.0-u1*u1)
	z := u1
	return NewVec3(x, y, z).Unit()
}

// SampleHemisphereCos returns a random unit vector (weighted) in a hemisphere
func SampleHemisphereCos() Vec3 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	th := 2 * math.Pi * u2
	r := math.Sqrt(u1)

	x := math.Cos(th) * r
	y := math.Sin(th) * r
	z := 1.0 - x*x - y*y
	if z <= 0.0 {
		z = 0.0
	} else {
		z = math.Sqrt(z)
	}
	return NewVec3(x, y, z).Unit()
}

// IStream streams in space-separated Vec3 values from a Reader
func (v Vec3) IStream(r io.Reader) error {
	_, err := fmt.Fscan(r, v.X(), v.Y(), v.Z())
	return err
}

// OStream writes space-separated Vec3 values to a Writer
func (v Vec3) OStream(w io.Writer) error {
	_, err := fmt.Fprint(w, v.X(), v.Y(), v.Z())
	return err
}
