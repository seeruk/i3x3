package grid_test

import (
	"testing"

	"github.com/SeerUK/i3x3/pkg/grid"
)

func TestWorkspaceGridPosition(t *testing.T) {
	var tests = []struct {
		workspace float64
		outputs   float64
		expected  float64
	}{
		{1, 1, 1},
		{1, 2, 1},
		{2, 2, 1},
		{15, 2, 8},
		{16, 2, 8},
		{25, 2, 13},
		{25, 1, 25},
		{34, 2, 17},
		{4, 3, 2},
		{19, 3, 7},
		{25, 3, 9},
	}

	for _, test := range tests {
		actual := grid.WorkspaceGridPosition(test.workspace, test.outputs)
		if actual != test.expected {
			t.Errorf(
				"Expected %v to equal %v for workspace %v, with %v outputs",
				actual,
				test.expected,
				test.workspace,
				test.outputs,
			)
		}
	}
}
