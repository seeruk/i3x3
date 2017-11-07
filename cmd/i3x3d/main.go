package main

import (
	"log"
	"os"
	"os/signal"
)

// i3x3d is an optional daemon used to perform non-critical tasks for i3x3. It is daemonised to
// improve performance, and to gather information about the environment at startup.
//
// The functionality currently includes:
// * Automatic workspace redistribution:
//   When you change your display configuration (i.e. remove an output from X, or add one, etc.),
//   i3x3d will detect this change, and automatically redistribute i3's workspaces in a way that
//   will ensure that i3x3 still behaves as expected. It does however mean that your containers may
//   end up on another output when you add a new output.
// * GTK-based overlay:
//   After initially building this into the i3x3ctl command, performance became an issue. Having the
//   overlay in i3x3d means GTK can start up and be initialised, leaving as little work as possible
//   left to do when we want the overlay to be shown. The result is a much more responsive overlay.
//
// No other functionality is planned at present.

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// @TODO: Port i3x3fixd and i3x3overlayd code into i3x3d.

	sig := <-signals

	log.Printf("caught signal: %v\n", sig)
}
