package rpc

import (
	"context"
	"errors"
	"time"

	"github.com/SeerUK/i3x3/pkg/proto"
	"github.com/inconshreveable/log15"
)

const (
	// DefaultPort is the port that the RPC server will listen on.
	DefaultPort uint16 = 44045
	// DefaultTimeout is the time the server will spend waiting for a response from other threads.
	DefaultTimeout = time.Second
)

var (
	// ErrTimeout is an error that is given if the RPC server isn't able to respond in time.
	ErrTimeout = errors.New("timed out waiting for internal response")
)

// Message is a message container that provides the structure to have more of a request / response
// cycle. Once the message has been processed, a response can be issued by passing the result (an
// error, or nil) to the response channel.
type Message struct {
	// Command is a command that comes from an RPC client.
	Command proto.DaemonCommand
	// ResponseCh is a channel to send a response down. The response may simply be nil, indicating
	// success. An error sent down this channel will likely be sent to the client (i3x3ctl).
	ResponseCh chan<- error
}

// NewMessage creates a new RPC message, and returns a channel that a response should be passed to.
func NewMessage(command *proto.DaemonCommand) (Message, chan error) {
	responseCh := make(chan error)
	message := Message{
		ResponseCh: responseCh,
		Command:    *command,
	}

	return message, responseCh
}

// Service is the GRPC server used to listen to commands to control i3x3. At it's core, it is what
// propagates messages throughout the application.
type Service struct {
	logger log15.Logger
	msgCh  chan<- Message
}

// NewService creates a new i3x3 RPC server.
func NewService(logger log15.Logger, msgCh chan<- Message) *Service {
	logger = logger.New("module", "rpc/rpc")

	return &Service{
		logger: logger,
		msgCh:  msgCh,
	}
}

// HandleCommand routes a command through the application so that it may be handled appropriately by
// other
func (s *Service) HandleCommand(ctx context.Context, cmd *proto.DaemonCommand) (*proto.DaemonCommandResponse, error) {
	msg, responseCh := NewMessage(cmd)
	timeoutCh := make(chan struct{})

	timer := time.AfterFunc(DefaultTimeout, func() {
		// "Broadcast" the timeout signal.
		close(timeoutCh)
	})

	defer timer.Stop()

	var err error

	go func() {
		// Don't block while trying to send the message. Try as long as we can, otherwise we just
		// time out, stopping this goroutine from leaking.
		select {
		case s.msgCh <- msg:
			s.logger.Debug("sent message",
				"direction", cmd.Direction,
				"move", cmd.Move,
				"overlay", cmd.Overlay,
			)
		case <-ctx.Done():
		case <-timeoutCh:
		}
	}()

	select {
	case err = <-responseCh:
	case <-ctx.Done():
		err = ctx.Err()
	case <-timeoutCh:
		err = ErrTimeout
	}

	var message string
	if err != nil {
		message = err.Error()
	}

	res := &proto.DaemonCommandResponse{
		Message: message,
	}

	return res, nil
}
