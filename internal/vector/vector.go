package vector

import (
	"math"
)

type Vector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (a Vector) Add(b Vector) Vector {
	return Vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a Vector) Subtract(b Vector) Vector {
	return Vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a Vector) Scale(b float64) Vector {
	return Vector{a.X * b, a.Y * b, a.Z * b}
}

func (a Vector) Length() float64 {
	return math.Sqrt(math.Pow(a.X, 2) + math.Pow(a.Y, 2) + math.Pow(a.Z, 2))
}

func (a Vector) Normalize() Vector {
	l := a.Length()
	if l == 0 {
		return a
	} else {
		return a.Scale(1.0 / l)
	}
}

func Zero() Vector { return Vector{X: 0.0, Y: 0.0, Z: 0.0} }
