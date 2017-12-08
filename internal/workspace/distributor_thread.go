package workspace

import (
	"context"
	"sync"
	"time"

	"sort"

	"fmt"

	"github.com/SeerUK/i3x3/internal/i3"
	"github.com/inconshreveable/log15"
)

// DistributionInterval is the amount of time between each automatic redistribution. These will
// happen periodically, regardless of any messages being sent down the message channel.
const DistributionInterval = 30 * time.Second

// DistributorThread is a long-running process that handles re-distributing workspaces so that i3x3
// remains functional. Whenever an X event occurs, or also periodically, workspaces will be placed
// on the output that i3x3 expects them to be on.
type DistributorThread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger

	msgCh <-chan struct{}
}

// NewDistributorThread creates a new workspace distributor thread.
func NewDistributorThread(logger log15.Logger, msgCh chan struct{}) *DistributorThread {
	logger = logger.New("module", "workspace/distributorThread")

	return &DistributorThread{
		logger: logger,
		msgCh:  msgCh,
	}
}

// Start attempts to start the distributor thread.
func (t *DistributorThread) Start() error {
	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.Unlock()

	ticker := time.NewTicker(DistributionInterval)

	t.logger.Info("thread started")

	defer func() {
		t.logger.Info("thread stopped")
	}()

	for {
		select {
		case <-ticker.C:
			err := redistributeWorkspaces()
			if err != nil {
				return err
			}
		case <-t.msgCh:
			err := redistributeWorkspaces()
			if err != nil {
				return err
			}
		case <-t.ctx.Done():
			return t.ctx.Err()
		}
	}
}

// Stop attempts to stop the distributor thread.
func (t *DistributorThread) Stop() error {
	if t.ctx != nil && t.cfn != nil {
		t.cfn()
	}

	return nil
}

func redistributeWorkspaces() error {
	outputs, err := i3.FindOutputs()
	if err != nil {
		return fmt.Errorf("couldn't find outputs: %v", err)
	}

	workspaces, err := i3.FindWorkspaces()
	if err != nil {
		return fmt.Errorf("couldn't find workspaces: %v", err)
	}

	activeOutputs := i3.ActiveOutputs(outputs)
	activeOutputsNum := len(activeOutputs)
	currentWorkspace := i3.CurrentWorkspaceNum(workspaces)

	// Sort the active outputs so that the primary display is always first.
	sort.Slice(activeOutputs, func(i, j int) bool {
		return activeOutputs[i].Primary
	})

	// Loop over the existing workspaces, and ensure they're on the display we expect them to be on,
	// only moving them if they're not in the right place.
	for _, workspace := range workspaces {
		workspaces := float64(workspace.Num)
		outputs := float64(activeOutputsNum)

		expected := i3.CurrentOutputNum(workspaces, outputs)
		expectedOutput := activeOutputs[int(expected)-1]

		if expectedOutput.Name != workspace.Output {
			err := i3.MoveWorkspaceToOutput(workspaces, expectedOutput.Name)
			if err != nil {
				return err
			}
		}
	}

	// Move focus back to original workspace.
	return i3.SwitchToWorkspace(currentWorkspace)
}
