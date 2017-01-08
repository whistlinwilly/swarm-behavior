package cubes

import (
	"testing"

	"github.com/whistlinwilly/swarm-behavior/internal/vector"
)

type testActor struct {
	position vector.Vector
}

func (ta *testActor) Position() vector.Vector {
	return ta.position
}

func TestCubeSplit(t *testing.T) {
	actors := []cubeValue{}
	i := -3.5
	for i <= 3.5 {
		actors = append(actors, &testActor{position: vector.Vector{X: i, Y: i, Z: i}})
		i = i + 0.5
	}
	cs := NewCubeSpace([]float64{-10.0, 10.0, -10.0, 10.0, -10.0, 10.0}, actors)
	if cs.root.holdsValues != false {
		t.Fatal("Expected cube to split")
	}
	// Subcubes are allocated on split
	if cs.root.subcubes == nil {
		t.Fatal("Expected subcubes to be allocated on split")
	}
	if cs.root.size() != len(actors) {
		t.Fatal("Expected number of values held to be constant after split")
	}
}
