package daemon

import "github.com/SeerUK/i3x3/pkg/rpc"

// OverlayThread is a thread that handles triggering the overlay.
type OverlayThread struct {
	commands <-chan rpc.Message
}

// NewOverlayThread creates a new overlay thread instance.
func NewOverlayThread(commands <-chan rpc.Message) *OverlayThread {
	return &OverlayThread{
		commands: commands,
	}
}

// Start attempts to start the overlay event loop.
func (t *OverlayThread) Start() error {
	return nil
}

// Stop gracefully stops the overlay thread.
func (t *OverlayThread) Stop() error {
	return nil
}
