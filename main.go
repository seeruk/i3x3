package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
)

// A direction represents the course that will be taken for movement through workspaces.
type direction string

const (
	up    direction = "up"
	down  direction = "down"
	left  direction = "left"
	right direction = "right"
)

var (
	// Grid x size, from env.
	x float64 = 3
	// Grid y size, from env.
	y float64 = 3
	// Current workspace, from i3.
	cw float64
	// Current output, from i3.
	co float64
	// Number of outputs, from i3.
	no float64
	// direction (i.e. "up", "down", "left", "right"), from flags.
	dir direction
)

// An edgeFunc takes the a workspace number, and then based on the other settings (such as current
// workspace, current output, number of outputs, so on) it will calculate if we're at the edge of a
// side of the grid. Each side has it's own function to calculate if the given workspace is on the
// edge, defined below.
type edgeFunc func(tar float64) bool

// The edgeFuncs are the aforementioned edge detection functions.
var edgeFuncs = map[direction]edgeFunc{
	// up detects if we're on the top edge.
	up: func(tar float64) bool {
		return tar-(no*x) <= 0
	},
	// down detects if we're on the bottom edge.
	down: func(tar float64) bool {
		return tar+(no*x) > (no*((x*y)-1))+co
	},
	// down detects if we're on the left edge.
	left: func(tar float64) bool {
		return math.Mod((tar-co)/y, 2) == 0
	},
	// down detects if we're on the right edge.
	right: func(tar float64) bool {
		return math.Mod(((tar-(no*(x-1)))-co)/y, 2) == 0
	},
}

// A targetFunc calculates the next workspace that would be moved to, in a given direction, based on
// other settings (such as current workspace, current output, number of outputs, so on). The return
// value may not be a valid workspace number. The use of the above edgeFuncs helps to ensure that
// targetFuncs only get called when necessary. This is especially important to note when dealing
// with sideways-movement, as unlike moving up and down, they can return valid workspace numbers
// when they should not be used.
type targetFunc func() float64

// The targetFuncs are available targetFunc functions.
var targetFuncs = map[direction]targetFunc{
	// up returns the workspace above.
	up: func() float64 {
		return cw - (no * x)
	},
	// down returns the workspace below.
	down: func() float64 {
		return cw + (no * x)
	},
	// left returns the workspace to the left.
	left: func() float64 {
		return cw - no
	},
	// right returns the workspace to the right.
	right: func() float64 {
		return cw + no
	},
}

// A rect represents a rectangle shape.
type rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// An output represents an i3 output (i.e. a display).
type output struct {
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Primary          bool   `json:"primary"`
	Rect             rect   `json:"rect"`
	CurrentWorkspace string `json:"current_workspace"`
}

// A workspace represents an i3 workspace.
type workspace struct {
	Num     int    `json:"num"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Focused bool   `json:"focused"`
	Rect    rect   `json:"rect"`
	Output  string `json:"output"`
	Urgent  bool   `json:"urgent"`
}

func main() {
	var move bool
	var sdir string

	flag.BoolVar(&move, "move", false, "Whether or not to move the focused container too")
	flag.StringVar(&sdir, "direction", "down", "The direction to move in (up, down, left, right)")
	flag.Parse()

	dir = direction(sdir)

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

	targetFn, ok := targetFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	edgeFn, ok := edgeFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	if edgeFn(cw) {
		// Already at edge, we shouldn't do anything.
		os.Exit(0)
	}

	target := targetFn()

	if move {
		moveToWorkspace(target)
	}

	switchToWorkspace(target)
}

// moveToWorkspace tells i3 to move the current container to the given workspace. It does not also
// switch to the workspace. Any error running the i3-msg command will be returned.
func moveToWorkspace(workspace float64) error {
	return exec.Command("i3-msg", "move", "container", "to", "workspace", fmt.Sprintf("%v", workspace)).Run()
}

// switchToWorkspace tells i3 to switch to the given workspace. Any error running the i3-msg command
// will be returned.
func switchToWorkspace(workspace float64) error {
	return exec.Command("i3-msg", "workspace", fmt.Sprintf("%v", workspace)).Run()
}

// getOutputs fetches an array of outputs from i3. This will return all outputs, not just the active
// ones.
func getOutputs() ([]output, error) {
	var outputs []output

	out, err := exec.Command("i3-msg", "-t", "get_outputs").Output()
	if err != nil {
		return []output{}, err
	}

	err = json.Unmarshal(out, &outputs)
	if err != nil {
		return []output{}, err
	}

	return outputs, nil
}

// getWorkspaces fetches an array of workspaces from i3. This will return all workspaces, not just
// the visible ones, or focused ones.
func getWorkspaces() ([]workspace, error) {
	var workspaces []workspace

	out, err := exec.Command("i3-msg", "-t", "get_workspaces").Output()
	if err != nil {
		return []workspace{}, err
	}

	err = json.Unmarshal(out, &workspaces)
	if err != nil {
		return []workspace{}, err
	}

	return workspaces, nil
}

// getActiveOutputCount counts the number of active outputs in the given outputs slice. This could
// be, but is unlikely to be 0.
func getActiveOutputCount(os []output) float64 {
	var aos []output

	for _, o := range os {
		if o.Active {
			aos = append(aos, o)
		}
	}

	return float64(len(aos))
}

// getCurrentOutputNum calculates the current "display number" that i3x3 uses internally based on
// the workspace number, and the number of outputs. We avoid trying to figure out the physical
// layout of displays because that will be both complicated, and error prone. This method works best
// if you only use i3x3 to move between workspaces.
func getCurrentOutputNum(workspace float64, outputs float64) float64 {
	mod := math.Mod(workspace, outputs)

	if mod == 0 {
		return outputs
	}

	return mod
}

// getCurrentWorkspaceNum gets the currently focused workspace number from the given workspaces.
func getCurrentWorkspaceNum(workspaces []workspace) float64 {
	for _, workspace := range workspaces {
		if workspace.Focused {
			return float64(workspace.Num)
		}
	}

	return 1.0
}
