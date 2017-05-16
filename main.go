package main

import (
	"fmt"
	"math"
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
	cw float64 = 18
	// Current output
	co float64 = 2
	// Number of outputs
	no float64 = 2
	// Direction (i.e. "up", "down", "left", "right")
	dir = Up
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

func main() {
	fmt.Println(cw)

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

	fmt.Printf("Moving %v from %v to %v\n", dir, cw, target)
}
