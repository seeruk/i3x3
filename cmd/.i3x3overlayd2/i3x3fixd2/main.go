package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/SeerUK/i3x3/internal/i3"
)

// i3x3fixd is used to redistribute i3's workspaces in a way that will allow i3x3 to function
// correctly. It's a particularly useful utility if you work on a laptop and add / remove displays,
// as i3 will automatically move workspaces around for you as you add / remove the displays.
//
// i3x3 automatically adjusts it's grid size, expanding to fit in new workspaces outside of the
// original grid, but because of the way it expects workspaces to be arranged, it can get into a bad
// state when displays are added or removed. For example, if you have 3 screens, and then one is
// removed, you may end up in a state where i3x3 will expect odd numbered workspaces to be on screen
// 1, when they could be on screen 2.
//
// When using this tool, your desktop may look a little bit crazy whilst it's re-arranging your
// workspaces - but it should be _very_ quick!

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Initialise our X connection.
	x, err := initialiseXEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	xevChan := make(chan struct{})

	go xevThread(x, xevChan)
	go fixThread(xevChan)

	sig := <-signals

	log.Printf("caught signal: %v\n", sig)
}

// fatal checks if the given err is nil, and if it is, logs the err and exits the application.
func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// initialiseXEnvironment sets up the connection to the X server, and prepare our randr
// configuration so we can act appropriately whenever an event happens.
func initialiseXEnvironment() (*xgb.Conn, error) {
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

// fixThread waits for signals from the given `in` channel, and when it receives one proceeds to
// update the workspace distribution so that i3x3 can continue to function correctly.
func fixThread(in chan struct{}) {
	log.Println("fix: waiting for notifications...")

	for {
		<-in

		log.Println("fix: got notification")

		outputs, err := i3.FindOutputs()
		fatal(err)

		workspaces, err := i3.FindWorkspaces()
		fatal(err)

		activeOutputs := i3.ActiveOutputs(outputs)
		activeOutputsNum := len(activeOutputs)

		currentWorkspace := i3.CurrentWorkspaceNum(workspaces)

		// Sort the active outputs so that the primary display is always first.
		sort.Slice(activeOutputs, func(i, j int) bool {
			return activeOutputs[i].Primary
		})

		// Loop over the existing workspaces, and ensure they're on the display we expect them to be on,
		// only moving them if they're not in the right place.
		for _, workspace := range workspaces {
			workspaces := float64(workspace.Num)
			outputs := float64(activeOutputsNum)

			expected := i3.CurrentOutputNum(workspaces, outputs)
			expectedOutput := activeOutputs[int(expected)-1]

			if expectedOutput.Name != workspace.Output {
				i3.MoveWorkspaceToOutput(workspaces, expectedOutput.Name)
			}
		}

		// Move focus back to original workspace.
		i3.SwitchToWorkspace(currentWorkspace)

		log.Println("fix: done fixing")
	}
}

// xevThread waits for x events to occur, and then notifies the given `out` thread.
func xevThread(x *xgb.Conn, out chan struct{}) {
	log.Println("xev: waiting for x events...")

	for {
		ev, xerr := x.WaitForEvent()
		if xerr != nil {
			log.Printf("error waiting for event: %v\n", xerr)
			continue
		}

		log.Printf("xev: got event: %+v\n", ev)

		out <- struct{}{}

		log.Println("xev: sent notification")
	}
}
