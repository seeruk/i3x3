package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/SeerUK/i3x3/pkg/overlayd"
)

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	environment, err := overlayd.FindEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	environmentState := &overlayd.EnvironmentState{}
	environmentState.SetEnvironment(environment)

	log.Printf("Outputs: %v\n", environment.ActiveOutputs)
	log.Printf("Output: %v\n", environment.CurrentOutput)
	log.Printf("Workspace: %v\n", environment.CurrentWorkspace)

	go func() {
		// Wait for signal
		<-signals

		// Cancel the context
		cfn()
	}()

	// @TODO: Grid environment.
	// Prepare the environment, and allow this information to be updated (protected by a RWMutex).
	// But what updates it? Is that going to be i3x3fix? Is it going to be a simple command? It's
	// likely to be exposed via a gRPC endpoint I guess, so maybe i3x3fix should be the one to run
	// it, but then it's polluting what it normally does.
	//
	// How about instead, we just build in the ability to kill old versions of the daemon, similar
	// to the way that imwheel works (e.g. a --kill flag?). It means we can use the flags package
	// at least, because it is very simple. We just need to investigate killing a process in code.
	// This approach will also help keep the gRPC messages focused solely on movement within the
	// grid too.

	// @TODO: GTK thread.
	// Look at how we did this in cnotifyd. That handles multi-threaded GTK pretty well, and it was
	// also pretty straightforward too - and is also using gotk3.

	// @TODO: Message handling thread.
	// Accept incoming messages to show overlay. This should contain the grid size, and the
	// target workspace. These messages should be in the form of ProtoBuf types, that are accepted
	// over a gRPC interface. Once a message comes in it should pass that information into a channel
	// that will be tied through the application to the UI thread, prompting the creation / updating
	// of the overlay.
	//
	// If the overlay is already visible, it needs to be updated. Can we handle updating the already
	// visible overlay? What would it take to do that? This could help reduce flickering, but could
	// cause some other issues, such as not showing the overlay on the correct output. We also run
	// the risk of running into memory leaks with gotk3 more with the daemonised approach.
	//
	// Maybe at first we should just try continuing with the same approach we have now, overlapping
	// new windows over each other, and see how having GTK already initialised and ready to go
	// affects performance - as this may be enough on it's own to keep it simple, and improve the
	// perceived performance.

	// Wait for context cancellation (from signal)
	<-ctx.Done()
}
