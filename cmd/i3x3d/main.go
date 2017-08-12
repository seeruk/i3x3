package main

import (
	"log"
	"net"

	"github.com/SeerUK/i3x3/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const port = ":7890"

type WorkspaceServer struct{}

func (s *WorkspaceServer) MoveWorkspace(ctx context.Context, req *proto.MoveWorkspaceRequest) (*proto.MoveWorkspaceResponse, error) {
	log.Printf("Move workspace request received: %v\n", req)

	return &proto.MoveWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) SwitchWorkspace(ctx context.Context, req *proto.SwitchWorkspaceRequest) (*proto.SwitchWorkspaceResponse, error) {
	log.Printf("Switch workspace request received: %v\n", req)

	return &proto.SwitchWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) RedistributeWorkspaces(ctx context.Context, req *proto.RedistributeWorkspacesRequest) (*proto.RedistributeWorkspacesResponse, error) {
	log.Printf("Redistribute workspace request received")

	return &proto.RedistributeWorkspacesResponse{}, nil
}

func main() {
	conn, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterWorkspaceServer(server, &WorkspaceServer{})
	server.Serve(conn)
}
