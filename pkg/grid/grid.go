package grid

import (
	"math"

	"github.com/SeerUK/i3x3/pkg/i3"
)

// Direction represents the course that will be taken for movement through workspaces.
type Direction string

// The direction constants are the available directions that workspaces can be traversed via.
const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

// Size represents the current size of the i3x3 grid.
type Size struct {
	// Integer "real" grid size, used for display.
	RealX int
	RealY int

	// Original "requested" grid size.
	OriginalX int
	OriginalY int
}

// NewSize initialises a new Size, to keep track of the grid size based on the current environment,
// and the requested grid size.
func NewSize(environment Environment, x int, y int) Size {
	// This next bit decides the real grid size. The real grid can be larger than the requested grid
	// if a workspace has ended up on a screen that shouldn't normally be there (e.g. after running
	// i3x3-fix, or if a user just creates a new workspace themselves that is outside of the bounds
	// of the grid).
	maxGridPos := WorkspaceGridPosition(environment.MaxWorkspace, environment.ActiveOutputs)
	maxRealRows := int(math.Ceil(maxGridPos / float64(x)))

	ry := y

	if maxRealRows > y {
		ry = maxRealRows
	}

	return Size{
		RealX:     x,
		RealY:     ry,
		OriginalX: x,
		OriginalY: y,
	}
}

// Environment represents the current state of the grid, and it's environment from i3.
type Environment struct {
	ActiveOutputs    float64
	CurrentOutput    float64
	CurrentWorkspace float64
	MaxWorkspace     float64
}

// NewEnvironment initialises a new environment, based on the given outputs and workspaces.
func NewEnvironment(outputs []i3.Output, workspaces []i3.Workspace) Environment {
	ao := i3.ActiveOutputs(outputs)
	cw := i3.CurrentWorkspace(workspaces)
	co := i3.CurrentOutput(cw, ao)
	mw := i3.MaxWorkspace(workspaces)

	return Environment{
		ActiveOutputs:    ao,
		CurrentOutput:    co,
		CurrentWorkspace: cw,
		MaxWorkspace:     mw,
	}
}

// An EdgeFunc takes the a workspace number, and then based on the other settings (such as current
// workspace, current output, number of outputs, so on) it will calculate if we're at the edge of a
// side of the grid. Each side has it's own function to calculate if the given workspace is on the
// edge, defined below.
type EdgeFunc func(tar float64) bool

// BuildEdgeFuncs creates the aforementioned edge detection functions.
func BuildEdgeFuncs(environment Environment, size Size) map[Direction]EdgeFunc {
	x := float64(size.RealX)
	y := float64(size.RealY)

	ao := environment.ActiveOutputs
	co := environment.CurrentOutput

	return map[Direction]EdgeFunc{
		// Up detects if we're on the top edge.
		Up: func(tar float64) bool {
			return tar-(ao*x) <= 0
		},
		// Down detects if we're on the bottom edge.
		Down: func(tar float64) bool {
			return tar+(ao*x) > (ao*((x*y)-1))+co
		},
		// Left detects if we're on the left edge.
		Left: func(tar float64) bool {
			return math.Mod((tar-co)/x, 2) == 0
		},
		// Right detects if we're on the right edge.
		Right: func(tar float64) bool {
			return math.Mod(((tar-(ao*(x-1)))-co)/x, 2) == 0
		},
	}
}

// A TargetFunc calculates the next workspace that would be moved to in a given direction, based on
// other info (such as current workspace, current output, number of outputs, so on). The return
// value may not be a valid workspace number. The use of the above EdgeFuncs helps to ensure that
// TargetFuncs only get called when necessary.
type TargetFunc func() float64

// BuildTargetFuncs creates the the aforementioned target workspace functions.
func BuildTargetFuncs(environment Environment, size Size) map[Direction]TargetFunc {
	x := float64(size.RealX)

	ao := environment.ActiveOutputs
	cw := environment.CurrentWorkspace

	return map[Direction]TargetFunc{
		// Up returns the workspace above.
		Up: func() float64 {
			return cw - (ao * x)
		},
		// Down returns the workspace below.
		Down: func() float64 {
			return cw + (ao * x)
		},
		// Left returns the workspace to the left.
		Left: func() float64 {
			return cw - ao
		},
		// Right returns the workspace to the right.
		Right: func() float64 {
			return cw + ao
		},
	}
}

// WorkspaceGridPosition calculates the position of a given workspace in a grid where each number
// increments by one, from the top left, to the bottom right. For example:
//
// Given 2 displays, the workspaces on the first screen may look like this:
//
//  1  3  5         1  2  3
//  7  9  11   ->   4  5  6
//  13 15 17        7  8  9
//
// The number returned for a given workspace from the left will return the position from the grid on
// the right. From that same example, this is an example result:
//
//  fmt.Println(WorkspaceGridPosition(15, 2))
//  // Output: 8
func WorkspaceGridPosition(workspace float64, outputs float64) float64 {
	if outputs == 1 {
		return workspace
	}

	output := i3.CurrentOutput(workspace, outputs)

	return (workspace + (outputs - output)) / outputs
}
