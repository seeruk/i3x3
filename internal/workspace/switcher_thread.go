package workspace

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/SeerUK/i3x3/internal/grid"
	"github.com/SeerUK/i3x3/internal/i3"
	"github.com/SeerUK/i3x3/internal/proto"
	"github.com/SeerUK/i3x3/internal/rpc"
	"github.com/inconshreveable/log15"
)

// SwitchTimeout is the amount of time the switcher will wait for outbound message acknowledgement.
const SwitchTimeout = time.Second

// SwitchResult is the result of an attempt to switch workspaces.
type SwitchMessage struct {
	// Context is a context used to cancel downstream events. It should be set with a timeout.
	Context context.Context
	// ResponseCh is a channel to send a response down. The response may simply be nil, indicating
	// success. If an error is sent, it may bubble up and be sent to the client.
	ResponseCh chan<- error
	// Environment is a grid environment (containing things about the current state of the grid).
	Environment grid.Environment
	// Target is the workspace we're going to switch to, if we're going to switch workspaces.
	Target float64
}

// NewSwitchMessage creates a new switch message, used to notify some consumer.
func NewSwitchMessage(ctx context.Context, env grid.Environment, target float64) (SwitchMessage, chan error) {
	responseCh := make(chan error, 1)

	message := SwitchMessage{
		Context:     ctx,
		ResponseCh:  responseCh,
		Environment: env,
		Target:      target,
	}

	return message, responseCh
}

// SwitchThread is the long-running workspace switcher. It will process one message at a time.
type SwitchThread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger

	msgCh <-chan rpc.Message
	outCh chan<- SwitchMessage
}

// NewSwitchThread creates a new workspace switcher thread.
func NewSwitchThread(logger log15.Logger, msgCh <-chan rpc.Message, outCh chan<- SwitchMessage) *SwitchThread {
	logger = logger.New("module", "workspace/switcherThread")

	return &SwitchThread{
		logger: logger,
		msgCh:  msgCh,
		outCh:  outCh,
	}
}

// Start a loop that will loop until canceled. It waits for messages to come in, handles them, and
// sends a response back to the message sender.
func (t *SwitchThread) Start() error {
	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.Unlock()

	defer func() {
		t.Lock()
		t.ctx = nil
		t.cfn = nil
		t.Unlock()
	}()

	t.logger.Info("thread started")

	defer func() {
		t.logger.Info("thread stopped")
	}()

	for {
		select {
		case msg := <-t.msgCh:
			// Similar to how an HTTP server might work, we accept new messages, and process them in
			// a goroutine. This allows messages to avoid blocking each other.
			go func() {
				ctx := msg.Context
				cmd := msg.Command

				// @TODO: Actually switch workspaces here.
				msg.ResponseCh <- t.handleCommand(ctx, cmd)

				t.logger.Debug("sent response",
					"direction", cmd.Direction,
					"move", cmd.Move,
					"overlay", cmd.Overlay,
				)
			}()
		case <-t.ctx.Done():
			return t.ctx.Err()
		}
	}
}

// Stop attempts to gracefully stop the switcher's loop.
func (t *SwitchThread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.ctx != nil && t.cfn != nil {
		t.cfn()
	}

	return nil
}

// handleCommand takes a daemon command, and actions it.
func (t *SwitchThread) handleCommand(ctx context.Context, cmd proto.DaemonCommand) error {
	// Perform the switch, returning information to react on in other threads.
	env, tar, err := t.switchWorkspace(cmd.Direction, cmd.Move, cmd.Overlay)

	ctx, _ = context.WithTimeout(ctx, SwitchTimeout)
	msg, responseCh := NewSwitchMessage(ctx, env, tar)

	if err == nil && cmd.Overlay {
		select {
		case t.outCh <- msg:
			t.logger.Debug("sent message",
				"target", fmt.Sprintf("%.0f", tar),
			)
		case <-t.ctx.Done():
			return fmt.Errorf("workspace/switcher: sending: %v", t.ctx.Err())
		case <-ctx.Done():
			return fmt.Errorf("workspace/switcher: sending: timed out")
		}

		select {
		case err := <-responseCh:
			if err != nil {
				return err
			}
		case <-t.ctx.Done():
			return fmt.Errorf("workspace/switcher: receiving: %v", t.ctx.Err())
		case <-ctx.Done():
			return fmt.Errorf("workspace/switcher: receiving: timed out")
		}
	}

	return err
}

// switchWorkspace actually performs the workspace switching, communicating with i3.
func (t *SwitchThread) switchWorkspace(direction string, move, overlay bool) (grid.Environment, float64, error) {
	dir := grid.Direction(direction)
	env := grid.Environment{}

	// Env-based config
	ix, err := envAsInt("I3X3_X_SIZE", 3)
	if err != nil {
		return env, 0, err
	}

	iy, err := envAsInt("I3X3_Y_SIZE", 3)
	if err != nil {
		return env, 0, err
	}

	outputs, err := i3.FindOutputs()
	if err != nil {
		return env, 0, err
	}

	workspaces, err := i3.FindWorkspaces()
	if err != nil {
		return env, 0, err
	}

	// Initialise the state of the grid.
	gridEnv := grid.NewEnvironment(outputs, workspaces)
	gridSize := grid.NewSize(gridEnv, ix, iy)

	edgeFuncs := grid.BuildEdgeFuncs(gridEnv, gridSize)
	targetFuncs := grid.BuildTargetFuncs(gridEnv, gridSize)

	targetFunc, ok := targetFuncs[dir]
	if !ok {
		return gridEnv, 0, fmt.Errorf("invalid direction: %q", direction)
	}

	edgeFunc, ok := edgeFuncs[dir]
	if !ok {
		return gridEnv, 0, fmt.Errorf("invalid direction: %q", direction)
	}

	// Check if we're at an edge...
	if edgeFunc(gridEnv.CurrentWorkspace) {
		// ... and if we are, just return.
		return gridEnv, 0, fmt.Errorf("hit edge of grid")
	}

	// Retrieve the target workspace that we should be moving to.
	target := targetFunc()

	if move {
		// If we need to move the currently focused container, we must do it before switching space,
		// because i3 will move whatever is focused when move is ran. In other words, this cannot be
		// handled concurrently.
		err = i3.MoveToWorkspace(target)
		if err != nil {
			return gridEnv, 0, err
		}
	}

	// Switch to the target workspace.
	err = i3.SwitchToWorkspace(target)
	if err != nil {
		return gridEnv, 0, err
	}

	return gridEnv, target, nil
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
