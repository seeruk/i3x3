package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/SeerUK/i3x3/internal/grid"
	"github.com/SeerUK/i3x3/internal/overlayd"
	"github.com/SeerUK/i3x3/internal/proto"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"google.golang.org/grpc"
)

func main() {
	ctx, cfn := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	messages := make(chan proto.OverlaydCommand)

	rpcDone := rpcThread(ctx, messages)
	guiDone := guiThread(ctx, messages)

	// Catch a signal
	sig := <-signals
	log.Printf("caught signal: %v\n", sig)

	// Propagate our cancellation.
	cfn()

	// Wait for everything to close, then exit.
	<-rpcDone
	<-guiDone
}

// rpcThread handles accepting new messages via gRPC from i3x3ctl.
func rpcThread(ctx context.Context, messages chan<- proto.OverlaydCommand) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%v", overlayd.Port))
		if err != nil {
			log.Fatal(err)
		}

		server := grpc.NewServer()
		proto.RegisterOverlaydServerServer(server, overlayd.NewServer(messages))

		go func() {
			select {
			case <-ctx.Done():
				log.Println("stopping gRPC thread")

				// Stop the gRPC server, signal we're done.
				server.GracefulStop()
				done <- struct{}{}
			}
		}()

		err = server.Serve(listen)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return done
}

// guiThread handles all GUI operations. Messages end up in here to be processed.
func guiThread(ctx context.Context, messages <-chan proto.OverlaydCommand) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		gtk.Init(nil)

		// Use dark theme.
		settings, _ := gtk.SettingsGetDefault()
		settings.SetProperty("gtk-application-prefer-dark-theme", true)

		window := buildWindow()

		// To notify multiple channels when a message has been received, we make some more channels
		enqueueChan := make(chan proto.OverlaydCommand)
		windowChan := make(chan struct{})

		// Then, we fan-out, emitting messages to each of those channels when a new message comes in
		go func() {
			for {
				select {
				case message := <-messages:
					// This could be a little more elegant...
					enqueueChan <- message
					windowChan <- struct{}{}
				case <-ctx.Done():
					break
				}
			}
		}()

		go enqueueMessages(ctx, window, enqueueChan)
		go windowReaper(ctx, window, windowChan)

		go func() {
			select {
			case <-ctx.Done():
				log.Println("stopping GUI thread")

				// Quit GTK, signal we're done, and then quit.
				gtk.MainQuit()
				done <- struct{}{}
			}
		}()

		gtk.Main()
	}()

	return done
}

// enqueueMessages processes incoming messages, handling updating the overlay window.
func enqueueMessages(ctx context.Context, window *gtk.Window, messages <-chan proto.OverlaydCommand) {
	for {
		select {
		case message := <-messages:
			glib.IdleAdd(processMessage, window, message)
		case <-ctx.Done():
			break
		}
	}
}

// windowReaper waits for a set amount of time before hiding the overlay window. If another message
// comes in whilst the window is open, it's life is extended.
func windowReaper(ctx context.Context, window *gtk.Window, messages <-chan struct{}) {
	var timer *time.Timer

	for {
		select {
		case <-messages:
			if timer != nil {
				timer.Stop()
			}

			timer = time.AfterFunc(500*time.Millisecond, func() {
				glib.IdleAdd(window.Hide)
			})
		case <-ctx.Done():
			break
		}
	}
}

// buildWindow creates the basic window that our overlay grid goes into.
func buildWindow() *gtk.Window {
	cssProvider, _ := gtk.CssProviderNew()
	cssProvider.LoadFromData(`
			.i3x3-window {
				background: #000000;
				color: #D3D3D3;
			}
		`)

	window, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
	window.SetAcceptFocus(false)
	window.SetDecorated(false)
	window.SetKeepAbove(true)
	window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	window.SetResizable(false)
	window.SetSkipTaskbarHint(true)
	window.SetTitle("i3x3 GTK WSS")
	window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NOTIFICATION)
	window.Stick()

	windowStyleContext, _ := window.GetStyleContext()
	windowStyleContext.AddClass("i3x3-window")
	windowStyleContext.AddProvider(cssProvider, 1)

	return window
}

// processMessage takes a message and updates the window UI appropriately, finally showing the
// window (if it's not already visible) at the end.
func processMessage(window *gtk.Window, message proto.OverlaydCommand) bool {
	environment, err := overlayd.FindEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	// Set up custom styles
	cssProvider, _ := gtk.CssProviderNew()
	cssProvider.LoadFromData(`
			.i3x3-grid {
				background: #2A2A2A;
				padding: 3px;
			}

			.i3x3-grid__box {
				background: #1A1A1A;
			}

			.i3x3-grid__box--active {
				background: #2A2A2A;
				color: #FFFFFF;
				font-weight: bold;
			}
		`)

	size := grid.NewSize(environment, 3, 3)
	target := float64(message.Target)

	// Remove all children...
	window.GetChildren().Foreach(func(item interface{}) {
		window.Remove(item.(*gtk.Widget))
	})

	ogrid, _ := gtk.GridNew()

	ogridStyleContext, _ := ogrid.GetStyleContext()
	ogridStyleContext.AddClass("i3x3-grid")
	ogridStyleContext.AddProvider(cssProvider, 1)

	labelCount := size.RealX * size.RealY

	for i := 0; i < labelCount; i++ {
		iao := int(environment.ActiveOutputs)
		ico := int(environment.CurrentOutput)

		ws := ico + (iao * i)

		label, _ := gtk.LabelNew("")
		label.SetMarkup(fmt.Sprintf("%d", int(ws)))

		box, _ := gtk.EventBoxNew()
		box.SetSizeRequest(50, 50)

		styles, _ := box.GetStyleContext()
		styles.AddClass("i3x3-grid__box")

		boxSC, _ := box.GetStyleContext()
		boxSC.AddProvider(cssProvider, 1)

		// Highlight the active workspace
		if int(target) == ws {
			styles.AddClass("i3x3-grid__box--active")
		}

		box.Add(label)

		row := i / size.RealX
		col := i - (row * size.RealX)

		// Attach it to the correct place in the table
		ogrid.Attach(box, col, row, 1, 1)
	}

	window.Add(ogrid)
	window.ShowAll()

	return false
}
