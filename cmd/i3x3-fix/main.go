package main

import (
	"sort"

	"github.com/SeerUK/i3x3/pkg/i3"
)

// i3x3-fix is used to redistribute i3's workspaces in a way that will allow i3x3 to function
// correctly. It's a particularly useful utility if you work on a laptop and add / remove displays,
// as i3 will automatically move workspaces around for you as you add / remove the displays.
//
// i3x3 automatically adjusts it's grid size, expanding to fit in new workspaces outside of the
// original grid, but because of the way it expects workspaces to be arranged, it can get into a bad
// state when displays are added or removed. For example, if you have 3 screens, and then one is
// removed, you may end up in a state where i3x3 will expect odd numbered workspaces to be on screen
// 1, when they could be on screen 2.
//
// When using this tool, your desktop may look a little bit crazy whilst it's re-arranging your
// workspaces - but it should be _very_ quick!

func main() {
	outputs, err := i3.FindOutputs()
	fatal(err)

	workspaces, err := i3.FindWorkspaces()
	fatal(err)

	activeOutputs := i3.ActiveOutputs(outputs)
	activeOutputsNum := len(activeOutputs)

	currentWorkspace := i3.CurrentWorkspaceNum(workspaces)

	// Sort the active outputs so that the primary display is always first (i.e. will always have
	// workspace 1), and then all others should be in alphabetical order.
	sort.Slice(activeOutputs, func(i, j int) bool {
		return activeOutputs[i].Primary || activeOutputs[i].Name < activeOutputs[j].Name
	})

	// Then loop over the existing workspaces, and ensure they're on the display we expect them to
	// be on, only moving them if they're not in the right place.
	for _, workspace := range workspaces {
		ws := float64(workspace.Num)
		os := float64(activeOutputsNum)

		expected := i3.CurrentOutputNum(ws, os)
		expectedOutput := activeOutputs[int(expected)-1]

		if expectedOutput.Name != workspace.Output {
			i3.MoveWorkspaceToOutput(ws, expectedOutput.Name)
		}
	}

	// Move focus back to original workspace.
	i3.SwitchToWorkspace(currentWorkspace)
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
