package main

import (
	"encoding/json"
	"fmt"
	"math"
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

type swarmConfig struct {
	maxX, minX, maxY, minY, maxZ, minZ int
	targetSelector                     func(*Swarm) vector.Vector
	actorPositionUpdater               func(*Swarm)
}

type Actor struct {
	Position     vector.Vector `json:"position"`
	IsLeader     bool          `json:"leader"`
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

func MoveLeaderToTargetUpdater(s *Swarm) {
	for i := range s.Actors {
		a := s.Actors[i]
		if a.IsLeader {
			d := s.Target.Subtract(s.Actors[i].Position)
			length := math.Sqrt(math.Pow(d.X, 2) + math.Pow(d.Y, 2) + math.Pow(d.Z, 2))
			if length < 15.0 { //hacky, but only one leader right now
				s.NextTarget()
			}
			s.Actors[i].velocity = a.velocity.Scale(a.deceleration).Add(d.Scale(1 / length).Scale(a.acceleration))
			s.Actors[i].Position = a.Position.Add(a.velocity)
		}
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
			{IsLeader: true, deceleration: 0.8, acceleration: 1.1},
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
