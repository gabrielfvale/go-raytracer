package tracer

import (
	"fmt"
	"math"

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

// NearestPhotons type definition
type NearestPhotons struct {
	max      int
	found    int
	got_heap int
	pos      [3]float64
	dist2    [2]float64
	index    []*Photon
}

// PhotonMap type definition
type PhotonMap struct {
	photons           *kdtree.KDTree
	storedPhotons     int
	halfStoredPhotons int
	maxPhotons        int
	prevScale         int

	costheta [256]float64
	sintheta [256]float64
	cosphi   [256]float64
	sinphi   [256]float64

	bboxMin [3]float64
	bboxMax [3]float64
}

// NewPhotonMap returns a PhotonMap with maxPhotons
func NewPhotonMap(maxPhotons int) (pmap PhotonMap) {
	pmap.storedPhotons = 0
	pmap.prevScale = 0
	pmap.maxPhotons = maxPhotons

	// points := []kdtree.Point
	pmap.photons = kdtree.New([]kdtree.Point{})
	pmap.bboxMin[0], pmap.bboxMin[1], pmap.bboxMin[2] = 1e8, 1e8, 1e8
	pmap.bboxMax[0], pmap.bboxMax[1], pmap.bboxMax[2] = -1e8, -1e8, -1e8

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

func dist2(p1, p2 [3]float64) (d float64) {
	d = 0.0
	for i := 0; i < 3; i++ {
		k := p2[i] - p1[i]
		d += k * k
	}
	return
}

// IrradianceEst returns the computed irradiance estimate at a given surface pos
func (pmap *PhotonMap) IrradianceEst(pos, normal [3]float64, maxDist float64, nphotons int) (irrad [3]float64) {
	irrad[0], irrad[1], irrad[2] = 0.0, 0.0, 0.0
	point := Photon{pos: pos}

	// locate the nearest photons
	nearest := pmap.photons.KNN(&point, nphotons)
	if len(nearest) == 0 {
		return
	}

	// Get only nearest that distance <= maxDist
	found := 0
	for ; found < len(nearest); found++ {
		maxDist2 := maxDist * maxDist
		nearPos := nearest[found].(*Photon).pos
		if dist2(pos, nearPos) > maxDist2 {
			break
		}
	}
	found--

	// if less than 8 photons, return
	if found < 8 {
		return
	}

	// sum radiance for all photons
	for i := 1; i < len(nearest); i++ {
		p := nearest[i].(*Photon)
		pdir := pmap.PhotonDir(p)
		if (pdir[0]*normal[0] + pdir[1]*normal[1] + pdir[2]*normal[2]) < 0.0 {
			irrad[0] += p.power[0]
			irrad[1] += p.power[1]
			irrad[2] += p.power[2]
		}
	}

	closest := nearest[found].(*Photon)
	dist2 := dist2(pos, closest.pos)

	// estimate of density
	tmp := (1.0 / math.Pi) / math.Sqrt(dist2)

	irrad[0] *= tmp
	irrad[1] *= tmp
	irrad[2] *= tmp
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
	// fmt.Println("stored photon", node)
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
