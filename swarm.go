package main

import (
  "os/exec"
  "runtime"
  "net/http"
  "encoding/json"
  "math/rand"
  "fmt"
)

type Swarm struct {
  Actors []Actor `json:"Actors"`
}

type Actor struct {
  XPos int `json:"x"`
  YPos int `json:"y"`
  ZPos int `json:"z"`
}

type swarmPositionHandler struct {
  s *Swarm
}

func (sph swarmPositionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  sph.s.dumbSwarmUpdate()
  json.NewEncoder(w).Encode(sph.s)
}

func (s *Swarm) dumbSwarmUpdate() {
  for i, _ := range s.Actors {
    s.Actors[i].XPos += (rand.Intn(20) - 10)
    s.Actors[i].YPos += (rand.Intn(20) - 10)
    s.Actors[i].ZPos += (rand.Intn(20) - 10)
  }
}

func main() {
  testSwarm := Swarm{
    Actors: []Actor{
      {XPos: 0, YPos: 0},
    },
  }

  mux := http.NewServeMux()
  mux.Handle("/swarm", swarmPositionHandler{s: &testSwarm})
  mux.Handle("/", http.FileServer(http.Dir("web/")))
  s := http.Server{
    Addr: ":8080",
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
