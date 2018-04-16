package rpc

import (
	"fmt"
	"net"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/seeruk/i3x3/internal/proto"
	"google.golang.org/grpc"
)

// Thread is a thread that when started will start an RPC server.
type Thread struct {
	sync.Mutex

	logger  log15.Logger
	service *Service
	server  *grpc.Server
}

// NewThread creates a new RPC thread.
func NewThread(logger log15.Logger, service *Service) *Thread {
	logger = logger.New("module", "rpc/thread")

	return &Thread{
		logger:  logger,
		service: service,
	}
}

// Start attempts to start listening on the configured port.
func (t *Thread) Start() error {
	defer func() {
		t.logger.Info("thread stopped")
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", DefaultPort))
	if err != nil {
		return fmt.Errorf("daemon/rpc: error launching listener: %v", err)
	}

	t.Lock()
	t.server = grpc.NewServer()
	t.Unlock()

	// Register our service type with our server.
	proto.RegisterDaemonServiceServer(t.server, t.service)

	t.logger.Info("thread started, listening",
		"port", DefaultPort,
	)

	return t.server.Serve(listener)
}

// Stop gracefully stops this server.
func (t *Thread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.server != nil {
		t.server.GracefulStop()
		t.server = nil
	}

	return nil
}
