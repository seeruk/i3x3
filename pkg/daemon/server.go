package daemon

import (
	"context"
	"fmt"
	"net"

	"github.com/SeerUK/i3x3/pkg/proto"
	"google.golang.org/grpc"
)

// Server is the GRPC server used to listen to commands to control i3x3. At it's core, it is what
// propagates messages throughout the application.
type Server struct {
	port uint16

	messages chan<- proto.DaemonCommand
	server   *grpc.Server
}

// NewServer creates a new i3x3 daemon server.
func NewServer(port uint16, messages chan<- proto.DaemonCommand) *Server {
	return &Server{
		port: port,
	}
}

// NewServerBackgroundThread starts the given server, waiting for the given context to signal for a
// shutdown event, allowing the server to end gracefully. The returned channel will receive a
// message when the server has finished shutting down.
func NewServerBackgroundThread(ctx context.Context, srv *Server) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		<-ctx.Done()

		// If our context says we're done, shut down the server.
		srv.Shutdown()
		close(done)
	}()

	// Start listening.
	srv.ListenAndServe()

	return done
}

// HandleCommand routes a command through the application so that it may be handled appropriately by
// other
func (s *Server) HandleCommand(ctx context.Context, cmd *proto.DaemonCommand) (*proto.DaemonCommandResponse, error) {
	res := &proto.DaemonCommandResponse{
		Target: 1,
	}

	return res, nil
}

// ListenAndServe attempts to start listening on the configured port.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.port))
	if err != nil {
		return fmt.Errorf("error launching listener: %v", err)
	}

	// Create the gRPC server
	s.server = grpc.NewServer()

	// Register our service type with our server.
	proto.RegisterDaemonServiceServer(s.server, s)

	return s.server.Serve(listener)
}

// Shutdown gracefully stops this server.
func (s *Server) Shutdown() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
