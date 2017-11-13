package daemon

import (
	"context"
	"fmt"
	"net"

	"github.com/SeerUK/i3x3/pkg/proto"
	"google.golang.org/grpc"
)

const (
	// RPCPort is the port that the RPC server will listen on.
	RPCPort uint16 = 44045
)

// RPCThread is a thread that when started will start an RPC server.
type RPCThread struct {
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

	// Create the gRPC server
	t.server = grpc.NewServer()

	// Register our service type with our server.
	proto.RegisterDaemonServiceServer(t.server, t.service)

	return t.server.Serve(listener)
}

// Stop gracefully stops this server.
func (t *RPCThread) Stop() error {
	if t.server != nil {
		t.server.GracefulStop()
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

	res := &proto.DaemonCommandResponse{
		// @TODO:
		Target: 1,
	}

	return res, nil
}
