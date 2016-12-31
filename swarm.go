package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/whistlinwilly/swarm-behavior/internal/vector"
)

type Swarm struct {
	Actors []Actor       `json:"Actors"`
	Target vector.Vector `json:"Target"`
	config swarmConfig
}

// For now, just return all other actors
func (s *Swarm) FindClosestN(a Actor, n int) []Actor {
	var actors []Actor
	for i := range s.Actors {
		if s.Actors[i] != a {
			actors = append(actors, s.Actors[i])
		}
	}
	return actors
}

type swarmConfig struct {
	maxX, minX, maxY, minY, maxZ, minZ int
	targetSelector                     func(*Swarm) vector.Vector
	actorPositionUpdater               func(*Swarm)
	separationSetSize                  int
	separationWeight                   float64
	alignmentWeight                    float64
	cohesionWeight                     float64
}

type Actor struct {
	Position     vector.Vector `json:"position"`
	IsLeader     bool          `json:"leader"`
	lastVelocity vector.Vector
	velocity     vector.Vector
	acceleration float64
	deceleration float64
}

type swarmPositionHandler struct {
	s *Swarm
}

func (sph swarmPositionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sph.s.config.actorPositionUpdater(sph.s)
	json.NewEncoder(w).Encode(sph.s)
}

func DumbSwarmUpdater(s *Swarm) {
	for i := range s.Actors {
		randomVector := vector.Vector{
			X: float64(rand.Intn(20) - 10),
			Y: float64(rand.Intn(20) - 10),
			Z: float64(rand.Intn(20) - 10),
		}
		s.Actors[i].Position = s.Actors[i].Position.Add(randomVector)
	}
}

// TODO: refactor magic constants into config
func MoveLeaderToTargetUpdater(s *Swarm) {
	for i := range s.Actors {
		delta := vector.Zero()
		if s.Actors[i].IsLeader {
			delta = s.Target.Subtract(s.Actors[i].Position)
			if delta.Length() < 15.0 { //hacky, but only one leader right now
				s.NextTarget()
			}
		} else {
			// Followers currently follow a linear combination of
			// a) Separation (short range repulsion)
			// b) Alignment (average heading of neighbors)
			// c) Cohesion (average position of neighbors)
			separationDelta := vector.Zero()
			alignmentDelta := vector.Zero()
			cohesionDelta := vector.Zero()

			neighbors := s.FindClosestN(s.Actors[i], s.config.separationSetSize)

			for j := range neighbors {
				fmt.Println("Neighbor position:", neighbors[j].Position)
				fmt.Println("My position:", s.Actors[i].Position)
				actorToNeighbor := neighbors[j].Position.Subtract(s.Actors[i].Position)
				fmt.Println("Actor to neighbor:", actorToNeighbor)
				actorToNeighborLength := actorToNeighbor.Length()
				// Separation
				if actorToNeighborLength < 5.0 {
					separationDelta = separationDelta.Add(actorToNeighbor.Scale(-1.0))
				}
				// Alignment
				alignmentDelta = alignmentDelta.Add(neighbors[j].lastVelocity)
				// Cohesion
				cohesionDelta = cohesionDelta.Add(actorToNeighbor)
			}
			separationDelta = separationDelta.Normalize()
			alignmentDelta = alignmentDelta.Normalize()
			cohesionDelta = cohesionDelta.Normalize()

			delta = separationDelta.Scale(s.config.separationWeight).Add(alignmentDelta.Scale(s.config.alignmentWeight)).Add(cohesionDelta.Scale(s.config.cohesionWeight))
		}
		s.Actors[i].velocity = s.Actors[i].velocity.Scale(s.Actors[i].deceleration).Add(delta.Normalize().Scale(s.Actors[i].acceleration))
	}
	for i := range s.Actors {
		s.Actors[i].Position = s.Actors[i].Position.Add(s.Actors[i].velocity)
		s.Actors[i].lastVelocity = s.Actors[i].velocity
	}
}

func (s *Swarm) NextTarget() {
	s.Target = s.config.targetSelector(s)
}

func RandomTargetSelector(s *Swarm) vector.Vector {
	x := float64(rand.Intn(s.config.maxX-s.config.minX) - ((s.config.maxX - s.config.minX) / 2))
	y := float64(rand.Intn(s.config.maxY-s.config.minY) - ((s.config.maxY - s.config.minY) / 2))
	z := float64(rand.Intn(s.config.maxZ-s.config.minZ) - ((s.config.maxZ - s.config.minZ) / 2))
	fmt.Printf("Picked new target (%v, %v, %v)\n", x, y, z)
	return vector.Vector{X: x, Y: y, Z: z}
}

func main() {
	testSwarm := &Swarm{
		Actors: []Actor{
			{IsLeader: true, deceleration: 0.8, acceleration: 1.1, Position: vector.Zero()},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: 10.0, Y: 10.0, Z: 10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: -10.0, Y: 10.0, Z: 10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: 10.0, Y: -10.0, Z: 10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: 10.0, Y: 10.0, Z: -10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: -10.0, Y: 10.0, Z: -10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: 10.0, Y: -10.0, Z: -10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: -10.0, Y: -10.0, Z: 10.0}},
			{IsLeader: false, deceleration: 0.8, acceleration: 1.1, Position: vector.Vector{X: -10.0, Y: -10.0, Z: -10.0}},
		},
		config: swarmConfig{
			minX:                 -100,
			maxX:                 100,
			minY:                 -100,
			maxY:                 100,
			minZ:                 -100,
			maxZ:                 100,
			targetSelector:       RandomTargetSelector,
			actorPositionUpdater: MoveLeaderToTargetUpdater,
			separationSetSize:    10,
			separationWeight:     5.0,
			alignmentWeight:      1.0,
			cohesionWeight:       1.5,
		},
	}
	testSwarm.NextTarget()

	mux := http.NewServeMux()
	mux.Handle("/swarm", swarmPositionHandler{s: testSwarm})
	mux.Handle("/", http.FileServer(http.Dir("web/")))
	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go openBrowser("http://localhost:8080/")
	fmt.Println("Running...")
	s.ListenAndServe()
}

// OS independent launcher
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin": //mac
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
