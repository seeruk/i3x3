package workspace

import (
	"context"
	"sync"

	"github.com/SeerUK/i3x3/pkg/rpc"
	"github.com/inconshreveable/log15"
)

// SwitchResult is the result of an attempt to switch workspaces.
type SwitchResult struct {
	// If an error occurred when switching, it will be sent in this field.
	Err error
	// ResponseCh is a channel to send a response down. The response may simply be nil, indicating
	// success. If an error is sent, it may bubble up and be sent to the client.
	ResponseCh chan<- error
	// Target is the workspace we're going to switch to, if we're going to switch workspaces.
	Target float64
}

// Switcher is the long-running workspace switcher. It will process one message at a time.
type Switcher struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger

	msgCh <-chan rpc.Message
	outCh chan<- SwitchResult
}

// NewSwitcher creates a new workspace switcher.
func NewSwitcher(msgCh <-chan rpc.Message, outCh chan<- SwitchResult) *Switcher {
	logger := log15.New("module", "workspace/switcher")

	return &Switcher{
		logger: logger,
		msgCh:  msgCh,
		outCh:  outCh,
	}
}

func (s *Switcher) Loop() error {
	s.ctx, s.cfn = context.WithCancel(context.Background())

	defer func() {
		s.ctx = nil
		s.cfn = nil
	}()

	for {
		select {
		case msg := <-s.msgCh:
			cmd := msg.Command

			s.logger.Debug("received message",
				"direction", cmd.Direction,
				"move", cmd.Move,
				"overlay", cmd.Overlay,
			)

			// @TODO: Actually switch workspaces here.
			msg.ResponseCh <- nil
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}

	return nil
}

func (s *Switcher) GracefulStop() {
	s.Lock()
	defer s.Unlock()

	if s.ctx != nil && s.cfn != nil {
		s.cfn()
	}
}
