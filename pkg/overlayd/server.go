package overlayd

import (
	"context"

	"github.com/SeerUK/i3x3/pkg/proto"
)

const (
	Port = 44044
)

// Server implements the OverlaydServer protocol.
type Server struct {
	messages chan<- proto.OverlaydCommand
}

// NewServer creates a new server instance.
func NewServer(messages chan<- proto.OverlaydCommand) *Server {
	return &Server{
		messages: messages,
	}
}

// SendCommand sends a command to the overlayd server.
func (s *Server) SendCommand(ctx context.Context, message *proto.OverlaydCommand) (*proto.OverlaydCommandResponse, error) {
	s.messages <- *message

	return &proto.OverlaydCommandResponse{
		Success: true,
	}, nil
}
