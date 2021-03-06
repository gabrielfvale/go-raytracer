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

// Min returns the minimum between two Vec3
func (v Vec3) Min(v2 Vec3) Vec3 {
	return NewVec3(math.Min(v.X(), v2.X()), math.Min(v.Y(), v2.Y()), math.Min(v.Z(), v2.Z()))
}

// Max returns the maximum between two Vec3
func (v Vec3) Max(v2 Vec3) Vec3 {
	return NewVec3(math.Max(v.X(), v2.X()), math.Max(v.Y(), v2.Y()), math.Max(v.Z(), v2.Z()))
}

// Sign returns a sign vector
func (v Vec3) Sign() Vec3 {
	signX, signY, signZ := 1.0, 1.0, 1.0
	if v.X() < 0.0 {
		signX = -1.0
	}
	if v.Y() < 0.0 {
		signY = -1.0
	}
	if v.Z() < 0.0 {
		signZ = -1.0
	}
	return NewVec3(signX, signY, signZ)
}

// Reflect returns a reflected Vec3 in relation to a normal n
func (v Vec3) Reflect(n Vec3) Vec3 {
	return v.Minus(n.Scale(2 * v.Dot(n))).Unit()
}

// Refract returns a refracted Vec3
func (v Vec3) Refract(n Vec3, etaRatio float64) (refracts bool, r Vec3) {
	refrN := n
	ratio := etaRatio

	if v.Dot(n) >= 0 { // ray inside
		refrN = n.Inv()     // invert normal
		ratio = 1.0 / ratio // effectively swap indexes
	}

	cosi := math.Min(v.Inv().Dot(refrN), 1.0)
	sini := math.Sqrt(1.0 - cosi*cosi)

	// Check for total internal reflection
	totalIntRefl := ratio*sini > 1.0
	// Use Schlick's approximation for reflectance
	r0 := (1 - ratio) / (1 + ratio)
	r0 = r0 * r0
	r0 = r0 + (1-r0)*math.Pow((1-cosi), 5)
	if totalIntRefl || r0 > rand.Float64() {
		return false, r
	}

	r1 := v.Plus(refrN.Scale(cosi)).Scale(ratio)
	r2 := refrN.Scale(-1 * math.Sqrt(math.Abs(1.0-r1.LenSq())))
	return true, r1.Plus(r2).Unit()
}

// SampleSphere returns a random unit vector in a sphere
func SampleSphere(rnd *rand.Rand) Vec3 {
	u1 := rnd.Float64()
	u2 := rnd.Float64()

	x := math.Cos(2*math.Pi*u2) * 2 * math.Sqrt(u1*(1.0-u1))
	y := math.Sin(2*math.Pi*u2) * 2 * math.Sqrt(u1*(1.0-u1))
	z := 1.0 - 2.0*u1
	return NewVec3(x, y, z).Unit()
}

// SampleHemisphere returns a random unit vector in a hemisphere
func SampleHemisphere(rnd *rand.Rand) Vec3 {
	u1 := rnd.Float64()
	u2 := rnd.Float64()

	x := math.Cos(2*math.Pi*u2) * 2 * math.Sqrt(1.0-u1*u1)
	y := math.Sin(2*math.Pi*u2) * 2 * math.Sqrt(1.0-u1*u1)
	z := u1
	return NewVec3(x, y, z).Unit()
}

// SampleHemisphereCos returns a random unit vector (weighted) in a hemisphere
func SampleHemisphereCos(rnd *rand.Rand) Vec3 {
	u1 := rnd.Float64()
	u2 := rnd.Float64()

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

func SampleHemisphereNormal(n Vec3, rnd *rand.Rand) Vec3 {
	r1 := 2 * math.Pi * rnd.Float64()
	r2 := rnd.Float64()
	r2s := math.Sqrt(r2)
	w := n
	u := NewVec3(1, 0, 0)
	if math.Abs(w.X()) > 0.1 {
		u = NewVec3(0, 1, 0)
	}
	u = u.Cross(w).Unit()
	v := w.Cross(u)

	uc := u.Scale(math.Cos(r1) * r2s)
	vc := v.Scale(math.Sin(r1) * r2s)
	wc := w.Scale(math.Sqrt(1 - r2))
	return uc.Plus(vc).Plus(wc).Unit()
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
