package rpc

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/SeerUK/i3x3/pkg/proto"
	"github.com/SeerUK/i3x3/pkg/rpc/rpctypes"
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
	// rpclog is the logger for the rpc thread.
	rpclog = log.New(os.Stderr, "daemon/rpc: ", 0)
)

// Service is the GRPC server used to listen to commands to control i3x3. At it's core, it is what
// propagates messages throughout the application.
type Service struct {
	messages chan<- rpctypes.Message
}

// NewService creates a new i3x3 RPC server.
func NewService(messages chan<- rpctypes.Message) *Service {
	return &Service{
		messages: messages,
	}
}

// HandleCommand routes a command through the application so that it may be handled appropriately by
// other
func (s *Service) HandleCommand(ctx context.Context, cmd *proto.DaemonCommand) (*proto.DaemonCommandResponse, error) {
	rpclog.Printf("received command: %+v\n", cmd)

	message, responseCh := rpctypes.NewMessage(cmd)
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
		case s.messages <- message:
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

	if err != nil {
		rpclog.Println(err)
	}

	res := &proto.DaemonCommandResponse{
		Message: err.Error(),
	}

	return res, nil
}
