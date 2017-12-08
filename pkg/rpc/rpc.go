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
	// Context is a context used to cancel downstream events. It should be set with a timeout.
	Context context.Context
	// ResponseCh is a channel to send a response down. The response may simply be nil, indicating
	// success. An error sent down this channel will likely be sent to the client (i3x3ctl).
	ResponseCh chan<- error
}

// NewMessage creates a new RPC message, and returns a channel that a response should be passed to.
func NewMessage(ctx context.Context, command *proto.DaemonCommand) (Message, chan error) {
	responseCh := make(chan error, 1)

	message := Message{
		Command:    *command,
		Context:    ctx,
		ResponseCh: responseCh,
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
	// For every new command that comes in, we make a new context. Sort of like a HTTP server.
	msgCtx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	msg, responseCh := NewMessage(msgCtx, cmd)

	var err error
	var res proto.DaemonCommandResponse

	select {
	case s.msgCh <- msg:
		s.logger.Debug("sent message",
			"direction", cmd.Direction,
			"move", cmd.Move,
			"overlay", cmd.Overlay,
		)
	case <-ctx.Done():
		err = ctx.Err()
	case <-msgCtx.Done():
		err = ErrTimeout
	}

	if err != nil {
		res.Message = err.Error()
		return &res, err
	}

	select {
	case err = <-responseCh:
	case <-ctx.Done():
		err = ctx.Err()
	case <-msgCtx.Done():
		err = ErrTimeout
	}

	if err != nil {
		res.Message = err.Error()
	}

	defer func() {
		s.logger.Debug("sent response",
			"response", res,
		)
	}()

	return &res, err
}
