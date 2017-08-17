package main

import (
	"sort"

	"github.com/SeerUK/i3x3/pkg/i3"
)

// i3x3-fix is used to redistribute i3's workspaces in a way that will allow i3x3 to function
// correctly. It's a particularly useful utility if you work on a laptop and add / remove displays,
// as i3 will automatically move workspaces around for you.
//
// i3x3 automatically adjusts it's grid size, expanding to fit in new workspaces, but because of the
// way it expects workspaces to be arranged, it can get into a bad state when displays are added or
// removed.
//
// When using this tool, your desktop may look a little bit crazy whilst it's re-arranging your
// workspaces - but it should be quite quick!

func main() {
	outputs, err := i3.FindOutputs()
	fatal(err)

	workspaces, err := i3.FindWorkspaces()
	fatal(err)

	activeOutputs := i3.ActiveOutputs(outputs)
	activeOutputsNum := len(activeOutputs)

	// Sort the active outputs so that the primary display is always first (i.e. will always have
	// workspace 1), and then all others should be in alphabetical order.
	sort.Slice(activeOutputs, func(i, j int) bool {
		return activeOutputs[i].Primary || activeOutputs[i].Name < activeOutputs[j].Name
	})

	for _, workspace := range workspaces {
		ws := float64(workspace.Num)
		os := float64(activeOutputsNum)

		expected := int(i3.CurrentOutputNum(ws, os))
		expectedOutput := activeOutputs[expected-1]

		if expectedOutput.Name != workspace.Output {
			// We need to move it...
			i3.MoveWorkspaceToOutput(workspace, expectedOutput)
		}
	}
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
