package main

import (
	"flag"
	"os"
	"strconv"

	"time"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/SeerUK/i3x3/pkg/i3"
	"github.com/SeerUK/i3x3/pkg/overlay"
)

// i3x3 is a workspace grid management utility for i3. It allows you to navigate workspaces by
// moving around a grid, instead of going to a specific numbered workspace. A major benefit of this
// approach is that you don't need as many keys to be bound to get around workspaces, and it scales
// quite well. If you want a very large grid, it's very quick to navigate.
//
// i3x3 relies on workspaces being positioned on the correct output. This is based on the number of
// outputs you currently have. If that number changes, it can affect the functionality of i3x3, and
// you may end up in a state where you cannot navigate to a workspace using i3x3. i3x3-fix can be
// used to automatically distribute your workspaces in a way that will allow i3x3 to function
// correctly. If you're on a desktop, you may never need to use this - it's aimed more at laptop
// users who are more likely to have to deal with adding and removing displays.

func main() {
	var move bool
	var noOverlay bool
	var sdir string

	flag.BoolVar(&move, "move", false, "Whether or not to move the focused container too")
	flag.StringVar(&sdir, "direction", "down", "The direction to move in (up, down, left, right)")
	flag.BoolVar(&noOverlay, "no-overlay", false, "Used to disable the GTK-based overlay")
	flag.Parse()

	dir := grid.Direction(sdir)

	// Env-based config
	ix, err := envAsInt("I3X3_X_SIZE", 3)
	fatal(err)
	iy, err := envAsInt("I3X3_Y_SIZE", 3)
	fatal(err)

	outputs, err := i3.FindOutputs()
	fatal(err)
	workspaces, err := i3.FindWorkspaces()
	fatal(err)

	// Initialise the state of the grid.
	gridEnv := grid.NewEnvironment(outputs, workspaces)
	gridSize := grid.NewSize(gridEnv, ix, iy)

	edgeFuncs := grid.BuildEdgeFuncs(gridEnv, gridSize)
	targetFuncs := grid.BuildTargetFuncs(gridEnv, gridSize)

	targetFunc, ok := targetFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	edgeFunc, ok := edgeFuncs[dir]
	if !ok {
		panic("Invalid direction")
	}

	// Check if we're at an edge...
	if edgeFunc(gridEnv.CurrentWorkspace) {
		// ... and if we are, just quit.
		os.Exit(0)
	}

	// Retrieve the target workspace that we should be moving to.
	target := targetFunc()

	var overlayDone <-chan time.Time

	if !noOverlay {
		// Create the UI overlay, based on the current grid, and target.
		overlayDone = overlay.Spawn(gridEnv, gridSize, target)
	}

	if move {
		// If we need to move the currently focused container, we must do it before switching space,
		// because i3 will move whatever is focused when move is ran. In other words, this cannot be
		// handled concurrently, using goroutines, sadly.
		err := i3.MoveToWorkspace(target)
		fatal(err)
	}

	// Switch to the target workspace.
	err = i3.SwitchToWorkspace(target)
	fatal(err)

	// Wait until the overlay timer ends before exiting.
	if !noOverlay {
		<-overlayDone
	}
}

// fatal panics if the given error is not nil.
func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

// envAsInt attempts to lookup the value of an environment variable by the given key. If it is not
// found then the given fallback value is used. If the value is found but can't be converted to a
// int, an error will be returned.
func envAsInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	return strconv.Atoi(val)
}
