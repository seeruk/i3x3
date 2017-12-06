package daemon

import "github.com/SeerUK/i3x3/pkg/rpc"

type OverlayThread struct {
	commands <-chan rpc.Message
}

func NewOverlayThread(commands <-chan rpc.Message) *OverlayThread {
	return &OverlayThread{
		commands: commands,
	}
}

func (t *OverlayThread) Start() error {
	return nil
}

func (t *OverlayThread) Stop() error {
	return nil
}
