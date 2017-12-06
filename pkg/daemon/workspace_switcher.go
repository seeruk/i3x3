package daemon

import (
	"fmt"

	"github.com/SeerUK/i3x3/pkg/workspace"
	"github.com/inconshreveable/log15"
)

// WorkspaceSwitcherThread is a pipe thread, both consuming and producing data on input and output
// channels. The workspace switcher thread manages actually telling i3 to switch workspaces based on
// incoming commands.
type WorkspaceSwitcherThread struct {
	logger   log15.Logger
	switcher *workspace.Switcher
}

// NewWorkspaceSwitcherThread creates a new workspace switcher thread.
func NewWorkspaceSwitcherThread(switcher *workspace.Switcher) *WorkspaceSwitcherThread {
	logger := log15.New("module", "daemon/workspaceSwitcher")

	return &WorkspaceSwitcherThread{
		logger:   logger,
		switcher: switcher,
	}
}

// Start attempts to start a switcher's loop.
func (t *WorkspaceSwitcherThread) Start() error {
	defer func() {
		t.logger.Info("thread stopped")
	}()

	t.logger.Info("thread started")

	return t.switcher.Loop()
}

// Stop gracefully stops the switcher.
func (t *WorkspaceSwitcherThread) Stop() error {
	if t.switcher == nil {
		return fmt.Errorf("daemon/workspaceSwitcher: switcher should not be nil")
	}

	t.switcher.GracefulStop()

	return nil
}
