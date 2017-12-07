package daemon

import (
	"fmt"
	"net"
	"sync"

	"github.com/SeerUK/i3x3/pkg/proto"
	"github.com/SeerUK/i3x3/pkg/rpc"
	"github.com/inconshreveable/log15"
	"google.golang.org/grpc"
)

// RPCThread is a thread that when started will start an RPC server.
type RPCThread struct {
	sync.Mutex

	logger  log15.Logger
	service *rpc.Service
	server  *grpc.Server
}

// NewRPCThread creates a new RPC thread.
func NewRPCThread(service *rpc.Service) *RPCThread {
	logger := log15.New("module", "daemon/rpc")

	return &RPCThread{
		logger:  logger,
		service: service,
	}
}

// Start attempts to start listening on the configured port.
func (t *RPCThread) Start() error {
	defer func() {
		t.logger.Info("thread stopped")
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", rpc.DefaultPort))
	if err != nil {
		return fmt.Errorf("daemon/rpc: error launching listener: %v", err)
	}

	t.Lock()
	t.server = grpc.NewServer()
	t.Unlock()

	// Register our service type with our server.
	proto.RegisterDaemonServiceServer(t.server, t.service)

	t.logger.Info("thread started, listening",
		"port", rpc.DefaultPort,
	)

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
