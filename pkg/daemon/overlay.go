package daemon

import (
	"github.com/SeerUK/i3x3/pkg/rpc/rpctypes"
)

type OverlayThread struct {
	commands <-chan rpctypes.Message
}

func NewOverlayThread(commands <-chan rpctypes.Message) *OverlayThread {
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
