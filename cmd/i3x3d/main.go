package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/SeerUK/i3x3/pkg/daemon"
	"github.com/SeerUK/i3x3/pkg/proto"
)

// i3x3d is an optional daemon used to perform non-critical tasks for i3x3. It is daemonised to
// improve performance, and to gather information about the environment at startup.
//
// The functionality currently includes:
// * Message handling:
//   Processing messages from i3x3ctl, sending commands to i3 as necessary to switch workspaces, or
//   move containers too.
// * Automatic workspace redistribution:
//   When you change your display configuration (i.e. remove an output from X, or add one, etc.),
//   i3x3d will detect this change, and automatically redistribute i3's workspaces in a way that
//   will ensure that i3x3 still behaves as expected. It does however mean that your containers may
//   end up on another output when you add a new output.
// * GTK-based overlay:
//   After initially building this into the i3x3ctl command, performance became an issue. Having the
//   overlay in i3x3d means GTK can start up and be initialised, leaving as little work as possible
//   left to do when we want the overlay to be shown. The result is a much more responsive overlay.

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	commands := make(chan proto.DaemonCommand)

	rpcService := daemon.NewRPCService(commands)
	rpcThread := daemon.NewRPCThread(rpcService)
	rpcThreadDone := daemon.NewBackgroundThread(ctx, rpcThread)

	sig := <-signals
	log.Printf("caught signal: %v\n", sig)

	cfn()

	// Wait for our background threads to clean up.
	<-rpcThreadDone
}
