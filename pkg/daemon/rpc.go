package daemon

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/SeerUK/i3x3/pkg/proto"
	"google.golang.org/grpc"
)

const (
	// RPCPort is the port that the RPC server will listen on.
	RPCPort uint16 = 44045
)

// RPCThread is a thread that when started will start an RPC server.
type RPCThread struct {
	sync.Mutex

	port uint16

	service *RPCService
	server  *grpc.Server
}

// NewRPCThread creates a new RPC thread.
func NewRPCThread(service *RPCService) *RPCThread {
	return &RPCThread{
		service: service,
	}
}

// Start attempts to start listening on the configured port.
func (t *RPCThread) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", RPCPort))
	if err != nil {
		return fmt.Errorf("error launching listener: %v", err)
	}

	t.Lock()
	t.server = grpc.NewServer()
	t.Unlock()

	// Register our service type with our server.
	proto.RegisterDaemonServiceServer(t.server, t.service)

	return t.server.Serve(listener)
}

// Stop gracefully stops this server.
func (t *RPCThread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.server != nil {
		t.server.GracefulStop()
		t.server = nil
	}

	return nil
}

// RPCService is the GRPC server used to listen to commands to control i3x3. At it's core, it is
// what propagates messages throughout the application.
type RPCService struct {
	messages chan<- proto.DaemonCommand
}

// NewRPCService creates a new i3x3 daemon server.
func NewRPCService(messages chan<- proto.DaemonCommand) *RPCService {
	return &RPCService{
		messages: messages,
	}
}

// HandleCommand routes a command through the application so that it may be handled appropriately by
// other
func (s *RPCService) HandleCommand(ctx context.Context, cmd *proto.DaemonCommand) (*proto.DaemonCommandResponse, error) {
	s.messages <- *cmd

	// @TODO: Rethink this. It should be for errors that need to be sent to the client. We should
	// probably make another channel that we receive from that holds the reply. But how on earth
	// would we know which value coming through that channel was related to this call of this
	// service? We couldn't link the two, and it'd be waiting for _any_ value to come through...
	// Maybe the client just needs to acknowledge the fact that it received the message, and then
	// any errors are just logged in the other daemon threads?
	//
	// To achieve the above, you could make a new type that contains the command, and a channel to
	// respond with the result in (would it need to be a pointer to that channel?) then when we're
	// done it'd send the result back. The client might not want to wait for that though. It should
	// be as fast as possible. Fire and forget.
	res := &proto.DaemonCommandResponse{
		// @TODO:
		Target: 1,
	}

	return res, nil
}
