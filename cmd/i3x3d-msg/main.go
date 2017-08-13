package main

import (
	"context"
	"flag"
	"log"

	"github.com/SeerUK/i3x3/proto"
	"google.golang.org/grpc"
)

const address = "localhost:7890"

func main() {
	var sdir string

	flag.StringVar(&sdir, "direction", "down", "The direction to move in (up, down, left, right)")
	flag.Parse()

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to i3x3d: %v", err)
	}

	defer conn.Close()

	client := proto.NewWorkspaceClient(conn)

	var direction proto.Direction

	switch sdir {
	case "up":
		direction = proto.Direction_UP
	case "down":
		direction = proto.Direction_DOWN
	case "left":
		direction = proto.Direction_LEFT
	case "right":
		direction = proto.Direction_RIGHT
	default:
		direction = proto.Direction_DOWN
	}

	sreq := proto.SwitchWorkspaceRequest{
		Direction: direction,
	}

	//mreq := proto.MoveWorkspaceRequest{
	//	Direction: proto.Direction_LEFT,
	//}
	//
	//rreq := proto.RedistributeWorkspacesRequest{}

	client.SwitchWorkspace(context.Background(), &sreq)
	//client.MoveWorkspace(context.Background(), &mreq)
	//client.RedistributeWorkspaces(context.Background(), &rreq)
}
