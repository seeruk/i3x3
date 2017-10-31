package overlayd

import (
	"sync"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/SeerUK/i3x3/pkg/i3"
)

// EnvironmentState is a thread-safe container for the current grid environment, protected with an
// embedded RWMutex.
type EnvironmentState struct {
	sync.RWMutex

	environment grid.Environment
}

// Environment returns the current environment that we have stored.
func (es *EnvironmentState) Environment() grid.Environment {
	es.RLock()
	defer es.RUnlock()

	return es.environment
}

// SetEnvironment sets a new environment, updating the stored state.
func (es *EnvironmentState) SetEnvironment(environment grid.Environment) {
	es.Lock()
	defer es.Unlock()

	es.environment = environment
}

// FindEnvironment fetches the current environment from i3.
func FindEnvironment() (grid.Environment, error) {
	outputs, err := i3.FindOutputs()
	if err != nil {
		return grid.Environment{}, err
	}

	workspaces, err := i3.FindWorkspaces()
	if err != nil {
		return grid.Environment{}, err
	}

	return grid.NewEnvironment(outputs, workspaces), nil
}
