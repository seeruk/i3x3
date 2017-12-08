package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/SeerUK/i3x3/internal/daemon"
	"github.com/SeerUK/i3x3/internal/proto"
	"google.golang.org/grpc"
)

func main() {
	var direction string
	var disableOverlay bool
	var move bool

	flag.BoolVar(&move, "move", false, "Whether or not to move the focused container too")
	flag.StringVar(&direction, "direction", "down", "The direction to move in (up, down, left, right)")
	flag.BoolVar(&disableOverlay, "no-overlay", false, "Used to disable the GTK-based overlay")
	flag.Parse()

	ctx, cfn := context.WithTimeout(context.Background(), time.Second)
	defer cfn()

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("127.0.0.1:%v", daemon.RPCPort), grpc.WithInsecure())
	fatal(err)

	defer conn.Close()

	client := proto.NewDaemonServiceClient(conn)

	resp, err := client.HandleCommand(ctx, &proto.DaemonCommand{
		Direction: direction,
		Overlay:   !disableOverlay,
		Move:      move,
	})

	fatal(err)

	fmt.Println(resp.Target)
}

// fatal panics if the given error is not nil.
func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
