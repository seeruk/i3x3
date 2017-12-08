package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/SeerUK/i3x3/internal/daemon"
	"github.com/SeerUK/i3x3/internal/rpc"
	"github.com/SeerUK/i3x3/internal/workspace"
	"github.com/inconshreveable/log15"
)

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
	var debug bool

	flag.BoolVar(&debug, "debug", false, "Enabled debug logging")
	flag.Parse()

	logLevel := log15.LvlInfo
	if debug {
		logLevel = log15.LvlDebug
	}

	baseLogger := log15.New()
	baseLogger.SetHandler(log15.LvlFilterHandler(logLevel, log15.StderrHandler))

	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	rpcMessages := make(chan rpc.Message)
	switchMessages := make(chan workspace.SwitchMessage)

	logger := baseLogger.New("module", "main/main")
	logger.Info("starting background threads")

	rpcService := rpc.NewService(baseLogger, rpcMessages)
	rpcThread := rpc.NewThread(baseLogger, rpcService)
	rpcThreadDone := daemon.NewBackgroundThread(ctx, rpcThread)

	workspaceSwitchThread := workspace.NewSwitchThread(baseLogger, rpcMessages, switchMessages)
	workspaceSwitchDone := daemon.NewBackgroundThread(ctx, workspaceSwitchThread)

	workspaceOverlayThread := workspace.NewOverlayThread(baseLogger, switchMessages)
	workspaceOverlayDone := daemon.NewBackgroundThread(ctx, workspaceOverlayThread)

	var metricsThreadDone <-chan daemon.BackgroundThreadResult
	if debug {
		metricsThread := daemon.NewMetricsThread(baseLogger)
		metricsThreadDone = daemon.NewBackgroundThread(ctx, metricsThread)
	}

	//overlayThread := daemon.NewOverlayThread(...)
	//overlayThreadDone := daemon.NewBackgroundThread(ctx, overlayThread)

	select {
	case sig := <-signals:
		fmt.Println() // Skip the ^C
		logger.Info("stopping background threads", "signal", sig)
	case res := <-rpcThreadDone:
		logger.Crit("error starting RPC thread", "error", res.Error)
	case res := <-workspaceSwitchDone:
		logger.Crit("error starting workspace switch thread", "error", res.Error)
	case res := <-workspaceOverlayDone:
		logger.Crit("error starting workspace overlay thread", "error", res.Error)
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
	<-workspaceOverlayDone
	<-workspaceSwitchDone

	if debug {
		<-metricsThreadDone
	}

	logger.Info("threads stopped, exiting")
}
