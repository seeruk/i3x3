package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Wait for signal
		<-signals

		// Cancel the context
		cfn()
	}()

	// @TODO: Prepare grid environment, and periodically update (?). See current overlay file.
	// @TODO: Initialise GTK thread.

	/*
	 * @TODO: Handle messages:
	 * Accept incoming messages to show overlay. This should contain the grid size, and the
	 * target workspace. These messages should be in the form of ProtoBuf types, that are accepted
	 * over a gRPC interface. Once a message comes in it should pass that information into a channel
	 * that will be tied through the application to the UI thread, prompting the creation / updating
	 * of the overlay.
	 *
	 * If the overlay is already visible, it needs to be updated. Can we handle updating the already
	 * visible overlay? What would it take to do that? This could help reduce flickering, but could
	 * cause some other issues, such as not showing the overlay on the correct output. We also run
	 * the risk of running into memory leaks with gotk3 more with the daemonised approach.
	 *
	 * Maybe at first we should just try continuing with the same approach we have now, overlapping
	 * new windows over each other, and see how having GTK already initialised and ready to go
	 * affects performance - as this may be enough on it's own to keep it simple, and improve the
	 * perceived performance.
	 */

	// Wait for context cancellation (from signal)
	<-ctx.Done()
}
