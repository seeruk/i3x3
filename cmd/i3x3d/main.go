package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/SeerUK/i3x3/pkg/daemon"
	"github.com/SeerUK/i3x3/pkg/rpc"
	"github.com/SeerUK/i3x3/pkg/workspace"
	"github.com/inconshreveable/log15"
)

var logger = log15.New("module", "main/main")

// i3x3d is the daemon used actually execute the functionality provided by i3x3. It is a daemon
// mainly to ensure better performance vs. running commands on-demand. This way, the client can be
// and ultra-thin wrapper; and the server can already be initialised, have gathered information
// about the environment from i3, and started other non-essential background work like initialising
// GTK and setting up the overlay.
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
//   @TODO: ^ This should also happen periodically (every 30 seconds or so?)
// * GTK-based overlay:
//   After initially building this into the i3x3ctl command, performance became an issue. Having the
//   overlay in i3x3d means GTK can start up and be initialised, leaving as little work as possible
//   left to do when we want the overlay to be shown. The result is a much more responsive overlay.

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	commands := make(chan rpc.Message, 1)
	switchResults := make(chan workspace.SwitchResult, 1)

	logger.Info("starting background threads")

	rpcService := rpc.NewService(commands)
	rpcThread := daemon.NewRPCThread(rpcService)
	rpcThreadDone := daemon.NewBackgroundThread(ctx, rpcThread)

	workspaceSwitcher := workspace.NewSwitcher(commands, switchResults)
	workspaceSwitcherThread := daemon.NewWorkspaceSwitcherThread(workspaceSwitcher)
	workspaceSwitcherDone := daemon.NewBackgroundThread(ctx, workspaceSwitcherThread)

	//overlayThread := daemon.NewOverlayThread(...)
	//overlayThreadDone := daemon.NewBackgroundThread(ctx, overlayThread)

	select {
	case sig := <-signals:
		fmt.Println() // Skip the ^C
		logger.Info("stopping background threads", "signal", sig)
	case res := <-rpcThreadDone:
		fatal(fmt.Errorf("error starting RPC thread: %v", res.Error))
		//case res := <-overlayThreadDone:
		//	fatal(fmt.Errorf("error starting overlay thread: %v", res.Error))
	case res := <-workspaceSwitcherDone:
		fatal(fmt.Errorf("error starting workspace switcher thread: %v", res.Error))
	}

	cfn()

	go func() {
		time.AfterFunc(5*time.Second, func() {
			logger.Error("took too long stopping, quitting")
			os.Exit(1)
		})
	}()

	// Wait for our background threads to clean up.
	<-rpcThreadDone
	//<-overlayThreadDone
	<-workspaceSwitcherDone

	logger.Info("threads stopped, exiting")
}

func fatal(err error) {
	if err != nil {
		logger.Crit("a fatal error occurred", "error", err)
	}
}
