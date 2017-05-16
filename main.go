package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
)

type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

// These come from env
const (
	// Grid x size
	x float64 = 3
	// Grid y size
	y float64 = 3
)

// These come from i3
var (
	// Current workspace
	cw float64
	// Current output
	co float64
	// Number of outputs
	no float64
	// Direction (i.e. "up", "down", "left", "right")
	dir = Down
)

type EdgeFunc func(tar float64) bool

var edgeFuncs = map[Direction]EdgeFunc{
	Up: func(tar float64) bool {
		return tar-(no*x) <= 0
	},
	Down: func(tar float64) bool {
		return tar+(no*x) > (no*((x*y)-1))+co
	},
	Left: func(tar float64) bool {
		return math.Mod((tar-co)/y, 2) == 0
	},
	Right: func(tar float64) bool {
		return math.Mod(((tar-(no*(x-1)))-co)/y, 2) == 0
	},
}

type TargetFunc func() float64

var targetFuncs = map[Direction]TargetFunc{
	Up: func() float64 {
		return cw - (no * x)
	},
	Down: func() float64 {
		return cw + (no * x)
	},
	Left: func() float64 {
		return cw - no
	},
	Right: func() float64 {
		return cw + no
	},
}

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Output struct {
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Primary          bool   `json:"primary"`
	Rect             Rect   `json:"rect"`
	CurrentWorkspace string `json:"current_workspace"`
}

type Workspace struct {
	Num     int    `json:"num"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Focused bool   `json:"focused"`
	Rect    Rect   `json:"rect"`
	Output  string `json:"output"`
	Urgent  bool   `json:"urgent"`
}

func main() {
	var move bool
	var sdir string

	flag.BoolVar(&move, "move", false, "Whether or not to move the focused container too")
	flag.StringVar(&sdir, "direction", "down", "The direction to move in (up, down, left, right)")
	flag.Parse()

	dir = Direction(sdir)

	outputs, err := getOutputs()
	if err != nil {
		panic(err)
	}

	workspaces, err := getWorkspaces()
	if err != nil {
		panic(err)
	}

	// Setup i3 values...
	no = getActiveOutputCount(outputs)
	cw = getCurrentWorkspaceNum(workspaces)
	co = getCurrentOutputNum(cw, no)

	fmt.Println(no)
	fmt.Println(cw)
	fmt.Println(co)

	targetFn, ok := targetFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	edgeFn, ok := edgeFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	if edgeFn(cw) {
		fmt.Println("Already at edge...")
		os.Exit(0)
	}

	target := targetFn()

	if move {
		moveToWorkspace(target)
	}

	switchToWorkspace(target)
}

func moveToWorkspace(workspace float64) error {
	return exec.Command("i3-msg", "move", "container", "to", "workspace", fmt.Sprintf("%v", workspace)).Run()
}

func switchToWorkspace(workspace float64) error {
	return exec.Command("i3-msg", "workspace", fmt.Sprintf("%v", workspace)).Run()
}

func getOutputs() ([]Output, error) {
	var outputs []Output

	out, err := exec.Command("i3-msg", "-t", "get_outputs").Output()
	if err != nil {
		return []Output{}, err
	}

	err = json.Unmarshal(out, &outputs)
	if err != nil {
		return []Output{}, err
	}

	return outputs, nil
}

func getWorkspaces() ([]Workspace, error) {
	var workspaces []Workspace

	out, err := exec.Command("i3-msg", "-t", "get_workspaces").Output()
	if err != nil {
		return []Workspace{}, err
	}

	err = json.Unmarshal(out, &workspaces)
	if err != nil {
		return []Workspace{}, err
	}

	return workspaces, nil
}

func getActiveOutputCount(os []Output) float64 {
	var aos []Output

	for _, o := range os {
		if o.Active {
			aos = append(aos, o)
		}
	}

	return float64(len(aos))
}

func getCurrentOutputNum(workspace float64, outputs float64) float64 {
	mod := math.Mod(workspace, outputs)

	if mod == 0 {
		return outputs
	}

	return mod
}

func getCurrentWorkspaceNum(workspaces []Workspace) float64 {
	for _, workspace := range workspaces {
		if workspace.Focused {
			return float64(workspace.Num)
		}
	}

	return 1.0
}
