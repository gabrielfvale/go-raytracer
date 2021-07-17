package tracer

import (
	"math"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"gonum.org/v1/gonum/spatial/kdtree"
)

// Photon type definition
type Photon struct {
	pos        [3]float64
	power      [3]float64
	plane      uint8
	theta, phi uint8
}

// Compare satisfies the axis comparisons method of the kdtree.Comparable interface.
func (p Photon) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(Photon)
	if d > 2 {
		panic("illegal dimension")
	}
	return p.pos[d] - q.pos[d]
}

// Dims returns the number of dimensions to be considered.
func (p Photon) Dims() int { return 3 }

// Distance returns the distance between the receiver and c.
func (p Photon) Distance(c kdtree.Comparable) float64 {
	q := c.(Photon)
	return dist2(p, q)
}

// Dust2 returns the distance between two photons
func dist2(p, q Photon) float64 {
	pPos := p.pos
	qPos := q.pos

	dist1 := pPos[0] - qPos[0]
	dist2 := dist1 * dist1

	dist1 = pPos[1] - qPos[1]
	dist2 += dist1 * dist1

	dist1 = pPos[2] - qPos[2]
	dist2 += dist1 * dist1

	return dist2
}

// photonList is a collection of the place type that satisfies kdtree.Interface.
type photonList []Photon

func (p photonList) Index(i int) kdtree.Comparable         { return p[i] }
func (p photonList) Len() int                              { return len(p) }
func (p photonList) Pivot(d kdtree.Dim) int                { return plane{photonList: p, Dim: d}.Pivot() }
func (p photonList) Slice(start, end int) kdtree.Interface { return p[start:end] }

// plane is required to help photonList.
type plane struct {
	kdtree.Dim
	photonList
}

func (p plane) Less(i, j int) bool {
	d := p.Dim
	if d > 2 {
		panic("illegal dimension")
	}
	return p.photonList[i].pos[d] < p.photonList[j].pos[d]
}
func (p plane) Pivot() int { return kdtree.Partition(p, kdtree.MedianOfMedians(p)) }
func (p plane) Slice(start, end int) kdtree.SortSlicer {
	p.photonList = p.photonList[start:end]
	return p
}
func (p plane) Swap(i, j int) {
	p.photonList[i], p.photonList[j] = p.photonList[j], p.photonList[i]
}

// PhotonMap type definition
type PhotonMap struct {
	photons       *kdtree.Tree
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

	pmap.photons = kdtree.New(photonList{}, true)

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
	useAllPhotons := radius == 0.0

	var keep kdtree.Keeper
	keep = kdtree.NewNKeeper(nphotons)
	pmap.photons.NearestSet(keep, point)

	found := 0
	// r2 is the squared distance to the nth nearest photon
	r2 := 0.0
	for _, c := range keep.(*kdtree.NKeeper).Heap {
		p := c.Comparable.(Photon)
		dist2 := p.Distance(point)
		if useAllPhotons || dist2 < radius2 {
			pdirf := pmap.PhotonDir(&p)
			pdir := geom.NewVec3(pdirf[0], pdirf[1], pdirf[2])
			if pdir.Dot(normal) < 0.0 {
				flux := geom.NewVec3(p.power[0], p.power[1], p.power[2])
				irrad = irrad.Plus(flux)
				r2 = dist2
				found++
			}
		}
	}

	// if less than 8 photons, return
	if found < 8 {
		return
	}

	// estimate of density
	tmp := 1.0 / (math.Pi * r2)
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
	pmap.photons.Insert(node, true)
}

// ScalePhotonPower is used to scale the power of all
// photons once they have been emitted from the light source.
func (pmap *PhotonMap) ScalePhotonPower(scale float64) {
	var newPhotons photonList
	pmap.photons.Do(func(c kdtree.Comparable, b *kdtree.Bounding, i int) (done bool) {
		k := c.(Photon)
		k.power[0] *= scale
		k.power[1] *= scale
		k.power[2] *= scale
		newPhotons = append(newPhotons, k)
		return
	})
	pmap.photons = kdtree.New(newPhotons, true)
	pmap.prevScale = pmap.storedPhotons + 1
}
