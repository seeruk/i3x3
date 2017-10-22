package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/SeerUK/i3x3/pkg/overlayd"
)

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	environmentState := &overlayd.EnvironmentState{}

	_, err := updateEnvironmentState(environmentState)
	if err != nil {
		log.Fatal(err)
	}

	x, err := prepareXEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	envChan := make(chan struct{})
	xevChan := make(chan struct{})

	go xevThread(x, xevChan)
	go envThread(environmentState, xevChan, envChan)
	go sigThread(signals, cfn)

	go func() {
		for {
			// @todo: Temporary.
			<-envChan
		}
	}()

	// @TODO: GTK thread.
	// Look at how we did this in cnotifyd. That handles multi-threaded GTK pretty well, and it was
	// also pretty straightforward too - and is also using gotk3. Should accept messages to control
	// the state of the GUI, i.e. show / hide a grid.

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

func envThread(state *overlayd.EnvironmentState, in chan struct{}, out chan struct{}) {
	for {
		// Wait for an event, then update the grid environment.
		<-in

		_, err := updateEnvironmentState(state)
		if err != nil {
			log.Println(err)
		} else {
			out <- struct{}{}
		}
	}
}

func sigThread(signals chan os.Signal, cfn context.CancelFunc) {
	// Wait for signal
	<-signals

	// Cancel the context
	cfn()
}

func xevThread(x *xgb.Conn, out chan struct{}) {
	for {
		_, xerr := x.WaitForEvent()
		if xerr != nil {
			log.Printf("error waiting for event: %v\n", xerr)
			continue
		}

		out <- struct{}{}
	}
}

// prepareXEnvironment sets up the connection to the X server, and prepare our randr configuration
// so we can act appropriately whenever an event happens.
func prepareXEnvironment() (*xgb.Conn, error) {
	x, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("error establishing X connectiong: %v", err)
	}

	err = randr.Init(x)
	if err != nil {
		return nil, fmt.Errorf("error initialising randr: %v", err)
	}

	// Get the root window on the default screen.
	root := xproto.Setup(x).DefaultScreen(x).Root

	// Choose which events we watch for.
	events := randr.NotifyMaskScreenChange

	// Subscribe to some events.
	err = randr.SelectInputChecked(x, root, uint16(events)).Check()
	if err != nil {
		return nil, fmt.Errorf("error subscribing to events: %v", err)
	}

	return x, nil
}

// updateEnvironmentState updates the current grid environment state.
func updateEnvironmentState(state *overlayd.EnvironmentState) (grid.Environment, error) {
	environment, err := overlayd.FindEnvironment()
	if err != nil {
		return environment, fmt.Errorf("error finding grid environment: %v", err)
	}

	state.SetEnvironment(environment)

	log.Printf("Outputs: %v\n", environment.ActiveOutputs)
	log.Printf("Output: %v\n", environment.CurrentOutput)
	log.Printf("Workspace: %v\n", environment.CurrentWorkspace)

	return environment, nil
}
