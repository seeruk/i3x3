package main

import "fmt"

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
	x = 4
	// Grid y size
	y = 6
)

// These come from i3
var (
	// Current workspace
	cw = 69
	// Current output
	co = 3
	// Number of outputs
	no = 3
	// Direction (i.e. "up", "down", "left", "right")
	dir = Right
)

type EdgeFunc func(tar int) bool

var edgeFuncs = map[Direction]EdgeFunc{
	Up: func(tar int) bool {
		return tar-(no*x) <= 0
	},
	Down: func(tar int) bool {
		return tar+(no*x) > (no*((x*y)-1))+co
	},
	Left: func(tar int) bool {
		return ((tar-co)/y)%2 == 0
	},
	Right: func(tar int) bool {
		return (((cw-(no*(x-1)))-co)/y)%2 == 0
	},
}

type TargetFunc func() int

var targetFuncs = map[Direction]TargetFunc{
	Up: func() int {
		return cw - (no * x)
	},
	Down: func() int {
		return cw + (no * x)
	},
	Left: func() int {
		return cw - no
	},
	Right: func() int {
		return cw + no
	},
}

func main() {
	targetFn, ok := targetFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	edgeFn, ok := edgeFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	if edgeFn(cw) {
		panic("Already at edge")
	}

	target := targetFn()

	fmt.Printf("Moving from %d to %d\n", cw, target)
}
