package cubes

import (
	"fmt"

	"github.com/whistlinwilly/swarm-behavior/internal/vector"
)

const maxValues int = 7

// Subcubes are referenced by positive / negative x, y, z
const (
	PXPYPZ = iota
	PXPYNZ
	PXNYPZ
	PXNYNZ
	NXPYPZ
	NXPYNZ
	NXNYPZ
	NXNYNZ
)

// Adjacent cubes are referenced by position / negative x, y, z
const (
	NX = iota
	PX
	NY
	PY
	NZ
	PZ
)

type cubeValue interface {
	Position() vector.Vector
}

// Cube represents a volume capable of holding either point values or pointers
// to subcubes
type Cube struct {
	numValues                          int
	holdsValues                        bool
	values                             []cubeValue
	minX, minY, minZ, maxX, maxY, maxZ float64
	adjCubes                           []*Cube
	subcubes                           []*Cube
	parent                             *Cube
}

func cubeFromBoundaries(minX, minY, minZ, maxX, maxY, maxZ float64) *Cube {
	return &Cube{
		numValues:   0,
		holdsValues: true,
		values:      []cubeValue{},
		minX:        minX,
		minY:        minY,
		minZ:        minZ,
		maxX:        maxX,
		maxY:        maxY,
		maxZ:        maxZ,
		adjCubes:    make([]*Cube, 6),
	}
}

func (c *Cube) contains(v vector.Vector) bool {
	if v.X < c.maxX && v.X >= c.minX {
		if v.Y < c.maxY && v.Y >= c.minY {
			if v.Z < c.maxZ && v.Z >= c.minZ {
				return true
			}
		}
	}
	return false
}

func (c *Cube) add(v cubeValue) {
	if c.holdsValues {
		c.values = append(c.values, v)
		c.numValues++
	} else {
		for _, x := range c.subcubes {
			if x.contains(v.Position()) {
				x.add(v)
			}
		}
	}
}

func (c *Cube) remove(i int) {
	c.values = append(c.values[:i], c.values[i+1:]...)
	c.numValues--
}

func (c *Cube) size() int {
	if c.holdsValues {
		return c.numValues
	}
	numValues := 0
	for _, cube := range c.subcubes {
		numValues += cube.size()
	}
	return numValues
}

func (c *Cube) resize() {
	if c.numValues > maxValues {
		c.split()
	} else if c.numValues == 0 {
		c.rejoin()
	}
	for _, cube := range c.subcubes {
		cube.resize()
	}
}

func (c *Cube) rejoin() {
	//TODO
}

func (c *Cube) split() {
	c.holdsValues = false
	c.numValues = 0

	// create subcubes
	c.subcubes = make([]*Cube, 8)
	middleX := (c.maxX + c.minX) / 2
	middleY := (c.maxY + c.minY) / 2
	middleZ := (c.maxZ + c.minZ) / 2
	fmt.Printf("Splitting cube (%p) at point (%v, %v, %v)\n", c, middleX, middleY, middleZ)
	c.subcubes[NXNYNZ] = cubeFromBoundaries(c.minX, c.minY, c.minZ, middleX, middleY, middleZ)
	c.subcubes[NXNYPZ] = cubeFromBoundaries(c.minX, c.minY, middleZ, middleX, middleY, c.maxZ)
	c.subcubes[NXPYNZ] = cubeFromBoundaries(c.minX, middleY, c.minZ, middleX, c.maxY, middleZ)
	c.subcubes[NXPYPZ] = cubeFromBoundaries(c.minX, middleY, middleZ, middleX, c.maxY, c.maxZ)
	c.subcubes[PXNYNZ] = cubeFromBoundaries(middleX, c.minY, c.minZ, c.maxX, middleY, middleZ)
	c.subcubes[PXNYPZ] = cubeFromBoundaries(middleX, c.minY, middleZ, c.maxX, middleY, c.maxZ)
	c.subcubes[PXPYNZ] = cubeFromBoundaries(middleX, middleY, c.minZ, c.maxX, c.maxY, middleZ)
	c.subcubes[PXPYPZ] = cubeFromBoundaries(middleX, middleY, middleZ, c.maxX, c.maxY, c.maxZ)

	// set parent
	for _, subcube := range c.subcubes {
		subcube.parent = c
	}

	fmt.Println("Subcubes:", c.subcubes)
	fmt.Println("Adjcubes:", c.adjCubes)

	// set adjacent subcubes
	c.subcubes[NXNYNZ].setAdjacentCubes(c.adjCubes[NX], c.subcubes[PXNYNZ], c.adjCubes[NY], c.subcubes[NXPYNZ], c.adjCubes[NZ], c.subcubes[NXNYPZ])
	c.subcubes[NXNYPZ].setAdjacentCubes(c.adjCubes[NX], c.subcubes[PXNYNZ], c.adjCubes[NY], c.subcubes[NXPYNZ], c.subcubes[NXNYNZ], c.adjCubes[PZ])
	c.subcubes[NXPYNZ].setAdjacentCubes(c.adjCubes[NX], c.subcubes[PXNYNZ], c.subcubes[NXNYNZ], c.adjCubes[PY], c.adjCubes[NZ], c.subcubes[NXNYPZ])
	c.subcubes[NXPYPZ].setAdjacentCubes(c.adjCubes[NX], c.subcubes[PXNYNZ], c.subcubes[NXNYPZ], c.adjCubes[PY], c.subcubes[NXNYNZ], c.adjCubes[PZ])
	c.subcubes[PXNYNZ].setAdjacentCubes(c.subcubes[NXNYNZ], c.adjCubes[PX], c.adjCubes[NY], c.subcubes[NXPYNZ], c.adjCubes[NZ], c.subcubes[NXNYPZ])
	c.subcubes[PXNYPZ].setAdjacentCubes(c.subcubes[NXNYPZ], c.adjCubes[PX], c.adjCubes[NY], c.subcubes[NXPYNZ], c.subcubes[NXNYNZ], c.adjCubes[PZ])
	c.subcubes[PXPYNZ].setAdjacentCubes(c.subcubes[NXPYNZ], c.adjCubes[PX], c.subcubes[NXNYNZ], c.adjCubes[PY], c.adjCubes[NZ], c.subcubes[NXNYPZ])
	c.subcubes[PXPYPZ].setAdjacentCubes(c.subcubes[NXPYPZ], c.adjCubes[PX], c.subcubes[NXNYPZ], c.adjCubes[PY], c.subcubes[NXNYNZ], c.adjCubes[PZ])

	vs := make([]cubeValue, len(c.values))
	for i, v := range c.values {
		vs[i] = v
	}

	for _, v := range vs {
		c.remove(0)
		c.add(v)
	}

	c.values = nil
	c.resize()
}

func (c *Cube) setAdjacentCubes(nx, px, ny, py, nz, pz *Cube) {
	c.adjCubes[NX] = nx
	c.adjCubes[PX] = px
	c.adjCubes[NY] = ny
	c.adjCubes[PY] = py
	c.adjCubes[NZ] = nz
	c.adjCubes[PZ] = pz
}

// CubeSpace represents a rectilinear grid (though at this point just a
// cartesean grid) capable of efficiently answering nearest neighbor
// queries (also, at this point, a lie)
type CubeSpace struct {
	valueToCube map[cubeValue]*Cube
	root        *Cube
}

// NewCubeSpace intializes a CubeSpace for use
func NewCubeSpace(boundaries []float64, values []cubeValue) *CubeSpace {
	adj := []*Cube{nil, nil, nil, nil, nil, nil}
	cube := &Cube{
		numValues:   len(values),
		holdsValues: true,
		values:      values,
		minX:        boundaries[0],
		minY:        boundaries[2],
		minZ:        boundaries[4],
		maxX:        boundaries[1],
		maxY:        boundaries[3],
		maxZ:        boundaries[5],
		adjCubes:    adj,
	}
	cubespace := &CubeSpace{valueToCube: make(map[cubeValue]*Cube), root: cube}
	for _, v := range values {
		cubespace.valueToCube[v] = cube
	}
	cube.resize()
	return cubespace
}
