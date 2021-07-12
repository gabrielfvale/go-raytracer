package tracer

import (
	"fmt"
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"github.com/kyroy/kdtree"
)

// Photon type definition
type Photon struct {
	pos        [3]float64
	power      [3]float64
	plane      uint8
	theta, phi uint8
}

func (p *Photon) Dimensions() int {
	return 3
}

func (p *Photon) Dimension(i int) float64 {
	return p.pos[i]
}

func (p *Photon) String() string {
	return fmt.Sprintf("{%.2f %.2f %.2f}", p.pos[1], p.pos[1], p.pos[2])
}

// PhotonMap type definition
type PhotonMap struct {
	photons       *kdtree.KDTree
	storedPhotons int
	maxPhotons    int
	prevScale     int

	costheta [256]float64
	sintheta [256]float64
	cosphi   [256]float64
	sinphi   [256]float64
}

// NewPhotonMap returns a PhotonMap with maxPhotons
func NewPhotonMap(maxPhotons int) (pmap PhotonMap) {
	pmap.storedPhotons = 0
	pmap.prevScale = 0
	pmap.maxPhotons = maxPhotons

	// points := []kdtree.Point
	pmap.photons = kdtree.New([]kdtree.Point{})

	// initialize direction conversion tables
	for i := 0; i < 256; i++ {
		angle := float64(i) * (1.0 / 256.0) * math.Pi
		pmap.costheta[i] = math.Cos(angle)
		pmap.sintheta[i] = math.Sin(angle)
		pmap.cosphi[i] = math.Cos(2.0 * angle)
		pmap.sinphi[i] = math.Sin(2.0 * angle)
	}

	return
}

// PhotonDir returns the direction of a photon
func (pmap *PhotonMap) PhotonDir(p *Photon) (dir [3]float64) {
	dir[0] = pmap.sintheta[p.theta] * pmap.cosphi[p.phi]
	dir[1] = pmap.sintheta[p.theta] * pmap.sinphi[p.phi]
	dir[2] = pmap.costheta[p.theta]
	return
}

// IrradianceEst returns the computed irradiance estimate at a given surface pos
func (pmap *PhotonMap) IrradianceEst(pos, normal geom.Vec3, radius float64, nphotons int) (irrad geom.Vec3) {
	irrad = geom.NewVec3(0.0, 0.0, 0.0)
	point := Photon{pos: pos.E}
	radius2 := radius * radius

	// locate the nearest photons
	nearest := pmap.photons.KNN(&point, nphotons)
	if len(nearest) == 0 {
		return
	}

	// Get only nearest that distance <= radius
	found := 0
	for ; found < len(nearest); found++ {
		p := nearest[found].(*Photon)

		pdirf := pmap.PhotonDir(p)
		ppos := geom.NewVec3(p.pos[0], p.pos[1], p.pos[2])
		pdir := geom.NewVec3(pdirf[0], pdirf[1], pdirf[2])
		t := ppos.Minus(pos)

		if t.Dot(t) < radius2 {
			if pdir.Dot(normal) < 0.0 {
				flux := geom.NewVec3(p.power[0], p.power[1], p.power[2])
				irrad = irrad.Plus(flux)
			}
		}
	}
	found--

	// if less than 8 photons, return
	if found < 8 {
		return
	}

	// estimate of density
	tmp := (1.0 / math.Pi) / radius2
	irrad = irrad.Scale(tmp)
	return
}

// Store puts a photon into the kd-tree
func (pmap *PhotonMap) Store(power, pos, dir [3]float64) {
	if pmap.storedPhotons >= pmap.maxPhotons {
		return
	}
	pmap.storedPhotons++
	var node Photon

	for i := 0; i < 3; i++ {
		node.pos[i] = pos[i]
		node.power[i] = power[i]
	}

	theta := int(math.Acos(dir[2]) * (256.0 / math.Pi))
	if theta > 255 {
		node.theta = 255
	} else {
		node.theta = uint8(theta)
	}

	phi := int(math.Atan2(dir[1], dir[0]) * (256.0 / (2.0 * math.Pi)))
	if phi > 255 {
		node.phi = 255
	} else if phi < 0 {
		node.phi = uint8(phi + 256)
	} else {
		node.phi = uint8(phi)
	}
	pmap.photons.Insert(&node)
}

// ScalePhotonPower is used to scale the power of all
// photons once they have been emitted from the light source.
func (pmap *PhotonMap) ScalePhotonPower(scale float64) {
	photons := pmap.photons.Points()
	newPhotons := make([]*Photon, pmap.storedPhotons)

	for i := 0; i < pmap.storedPhotons; i++ {
		conv := photons[i].(*Photon)
		newPhotons[i] = conv
	}

	for i := pmap.prevScale; i < pmap.storedPhotons; i++ {
		newPhotons[i].power[0] *= scale
		newPhotons[i].power[1] *= scale
		newPhotons[i].power[2] *= scale
	}

	var newPoints []kdtree.Point
	for i := 0; i < pmap.storedPhotons; i++ {
		newPoints = append(newPoints, newPhotons[i])
	}

	pmap.photons = kdtree.New(newPoints)
	pmap.prevScale = pmap.storedPhotons + 1
}

func (pmap *PhotonMap) Balance() {
	// fmt.Println("started balancing")
	pmap.photons.Balance()
}
