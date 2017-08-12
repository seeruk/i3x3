package main

import (
	"context"
	"log"

	"github.com/SeerUK/i3x3/proto"
	"google.golang.org/grpc"
)

const address = "localhost:7890"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to i3x3d: %v", err)
	}

	defer conn.Close()

	client := proto.NewWorkspaceClient(conn)

	sreq := proto.SwitchWorkspaceRequest{
		Direction: proto.Direction_DOWN,
	}

	mreq := proto.MoveWorkspaceRequest{
		Direction: proto.Direction_LEFT,
	}

	rreq := proto.RedistributeWorkspacesRequest{}

	client.SwitchWorkspace(context.Background(), &sreq)
	client.MoveWorkspace(context.Background(), &mreq)
	client.RedistributeWorkspaces(context.Background(), &rreq)
}
