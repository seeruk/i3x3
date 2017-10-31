package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/SeerUK/i3x3/pkg/overlayd"
	"github.com/SeerUK/i3x3/pkg/proto"
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
	fmt.Println(message)
}
