package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"time"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/SeerUK/i3x3/pkg/overlayd"
	"github.com/SeerUK/i3x3/pkg/proto"
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

func guiThread(ctx context.Context, messages <-chan proto.OverlaydCommand) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		gtk.Init(nil)

		settings, _ := gtk.SettingsGetDefault()
		settings.SetProperty("gtk-application-prefer-dark-theme", true)

		go enqueueMessages(ctx, messages)

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

func enqueueMessages(ctx context.Context, messages <-chan proto.OverlaydCommand) {
	for {
		select {
		case message := <-messages:
			glib.IdleAdd(processMessage, message)
		case <-ctx.Done():
			break
		}
	}
}

func processMessage(message proto.OverlaydCommand) {
	log.Println(message)

	environment, err := overlayd.FindEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	size := grid.NewSize(environment, 3, 3)
	target := float64(message.Target)

	// Use dark theme.
	settings, _ := gtk.SettingsGetDefault()
	settings.SetProperty("gtk-application-prefer-dark-theme", true)

	// Set up custom styles
	cssProvider, _ := gtk.CssProviderNew()
	cssProvider.LoadFromData(`
			.i3x3-window {
				background: #000000;
				color: #D3D3D3;
			}

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

	// Create the window
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

	go func() {
		time.Sleep(500 * time.Millisecond)

		window.Hide()
		window.Close()
	}()
}
