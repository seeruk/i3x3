package daemon

import (
	"fmt"
	"net"
	"sync"

	"github.com/SeerUK/i3x3/pkg/proto"
	"github.com/SeerUK/i3x3/pkg/rpc"
	"google.golang.org/grpc"
)

// RPCThread is a thread that when started will start an RPC server.
type RPCThread struct {
	sync.Mutex

	service *rpc.Service
	server  *grpc.Server
}

// NewRPCThread creates a new RPC thread.
func NewRPCThread(service *rpc.Service) *RPCThread {
	return &RPCThread{
		service: service,
	}
}

// Start attempts to start listening on the configured port.
func (t *RPCThread) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", rpc.DefaultPort))
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
