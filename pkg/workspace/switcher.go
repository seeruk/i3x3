package workspace

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/SeerUK/i3x3/pkg/i3"
	"github.com/SeerUK/i3x3/pkg/proto"
	"github.com/SeerUK/i3x3/pkg/rpc"
	"github.com/inconshreveable/log15"
)

// SwitchResult is the result of an attempt to switch workspaces.
type SwitchMessage struct {
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
	outCh chan<- SwitchMessage
}

// NewSwitcher creates a new workspace switcher.
func NewSwitcher(logger log15.Logger, msgCh <-chan rpc.Message, outCh chan<- SwitchMessage) *Switcher {
	logger = logger.New("module", "workspace/switcher")

	return &Switcher{
		logger: logger,
		msgCh:  msgCh,
		outCh:  outCh,
	}
}

// Loop starts a loop that will loop until canceled. It waits for messages to come in, handles them,
// and sends a response back to the message sender.
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
			msg.ResponseCh <- s.handleCommand(cmd)
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

// GracefulStop attempts to stop the switcher's loop.
func (s *Switcher) GracefulStop() {
	s.Lock()
	defer s.Unlock()

	if s.ctx != nil && s.cfn != nil {
		s.cfn()
	}
}

// handleCommand takes a daemon command, and actions it.
func (s *Switcher) handleCommand(cmd proto.DaemonCommand) error {
	dir := grid.Direction(cmd.Direction)

	// Env-based config
	ix, err := envAsInt("I3X3_X_SIZE", 3)
	if err != nil {
		return err
	}

	iy, err := envAsInt("I3X3_Y_SIZE", 3)
	if err != nil {
		return err
	}

	outputs, err := i3.FindOutputs()
	if err != nil {
		return err
	}

	workspaces, err := i3.FindWorkspaces()
	if err != nil {
		return err
	}

	// Initialise the state of the grid.
	gridEnv := grid.NewEnvironment(outputs, workspaces)
	gridSize := grid.NewSize(gridEnv, ix, iy)

	edgeFuncs := grid.BuildEdgeFuncs(gridEnv, gridSize)
	targetFuncs := grid.BuildTargetFuncs(gridEnv, gridSize)

	targetFunc, ok := targetFuncs[dir]
	if !ok {
		return fmt.Errorf("invalid direction: %q", cmd.Direction)
	}

	edgeFunc, ok := edgeFuncs[dir]
	if !ok {
		return fmt.Errorf("invalid direction: %q", cmd.Direction)
	}

	// Check if we're at an edge...
	if edgeFunc(gridEnv.CurrentWorkspace) {
		// ... and if we are, just return.
		return nil
	}

	// Retrieve the target workspace that we should be moving to.
	target := targetFunc()

	responseCh := make(chan error, 1)

	go func() {
		result := SwitchMessage{
			ResponseCh: responseCh,
			Target:     target,
		}

		select {
		case s.outCh <- result:
		default:
		}
	}()

	if cmd.Move {
		// If we need to move the currently focused container, we must do it before switching space,
		// because i3 will move whatever is focused when move is ran. In other words, this cannot be
		// handled concurrently.
		err = i3.MoveToWorkspace(target)
		if err != nil {
			return err
		}
	}

	// Switch to the target workspace.
	return i3.SwitchToWorkspace(target)
}

// envAsInt attempts to lookup the value of an environment variable by the given key. If it is not
// found then the given fallback value is used. If the value is found but can't be converted to a
// int, an error will be returned.
func envAsInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	return strconv.Atoi(val)
}
