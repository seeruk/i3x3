package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"sync"

	"github.com/SeerUK/i3x3/proto"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const port = ":7890"

type WorkspaceServer struct {
	overlay *GridOverlay
}

func (s *WorkspaceServer) MoveWorkspace(ctx context.Context, req *proto.MoveWorkspaceRequest) (*proto.MoveWorkspaceResponse, error) {
	log.Printf("Move workspace request received: %v\n", req)

	return &proto.MoveWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) SwitchWorkspace(ctx context.Context, req *proto.SwitchWorkspaceRequest) (*proto.SwitchWorkspaceResponse, error) {
	log.Printf("Switch workspace request received: %v\n", req)

	go func() {
		s.overlay.Interrupt()

		s.overlay.Lock()
		defer s.overlay.Unlock()

		s.overlay.CreateWindow()
		s.overlay.ShowWindow()

		timer := time.NewTimer(1 * time.Second)

		select {
		case result := <-timer.C:
			log.Printf("Hiding from timer: %v\n", result)
			s.overlay.DestroyWindow()
		case <-s.overlay.interrupt:
			log.Println("Hiding from interrupt")
			timer.Stop()
			s.overlay.DestroyWindow()
		}
	}()

	return &proto.SwitchWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) RedistributeWorkspaces(ctx context.Context, req *proto.RedistributeWorkspacesRequest) (*proto.RedistributeWorkspacesResponse, error) {
	log.Printf("Redistribute workspace request received")

	return &proto.RedistributeWorkspacesResponse{}, nil
}

func main() {
	// Initialise GLib, GDK, and GTK
	glib.ThreadInit(nil)
	gdk.ThreadsInit()
	gdk.ThreadsEnter()
	gtk.Init(&os.Args)

	defer gdk.ThreadsLeave()

	conn, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		// Don't block the main thread with GTK
		gtk.Main()
	}()

	overlay := &GridOverlay{
		interrupt: make(chan bool, 1),
	}

	server := grpc.NewServer()
	proto.RegisterWorkspaceServer(server, &WorkspaceServer{
		overlay: overlay,
	})
	server.Serve(conn)
}

type GridOverlay struct {
	sync.Mutex

	interrupt chan bool
	window    *gtk.Window
}

func (g *GridOverlay) Interrupt() {
	if g.window == nil {
		return
	}

	g.interrupt <- true
}

func (g *GridOverlay) ShowWindow() {
	if g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.ShowAll()
	gdk.ThreadsLeave()
}

func (g *GridOverlay) DestroyWindow() {
	if g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.Destroy()
	gdk.ThreadsLeave()

	g.window = nil
}

func (g *GridOverlay) CreateWindow() *gtk.Window {
	ix := 3
	iy := 3

	co := 1
	no := 2

	target := 5

	// Create the window
	window := gtk.NewWindow(gtk.WINDOW_POPUP)
	window.SetAcceptFocus(false)
	window.SetDecorated(false)
	window.SetKeepAbove(true)
	window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	window.SetResizable(false)
	window.SetSkipTaskbarHint(true)
	window.SetTitle("i3x3 GTK WSS")
	window.SetTypeHint(gdk.WINDOW_TYPE_HINT_NOTIFICATION)
	window.Stick()

	// Set main colours
	window.ModifyBG(gtk.STATE_NORMAL, gdk.NewColor("#000000"))
	window.ModifyFG(gtk.STATE_NORMAL, gdk.NewColor("#D3D3D3"))

	table := gtk.NewTable(uint(ix), uint(iy), false)
	table.SetBorderWidth(3)

	labelCount := ix * iy

	for i := 0; i < labelCount; i++ {
		ico := int(co)
		ino := int(no)

		ws := ico + (ino * i)

		label := gtk.NewLabel("")
		label.SetMarkup(fmt.Sprintf("%d", int(ws)))

		box := gtk.NewEventBox()
		box.SetSizeRequest(100, 100)
		box.ModifyBG(gtk.STATE_NORMAL, gdk.NewColor("#1A1A1A"))

		// Highlight the active workspace
		if int(target) == ws {
			box.ModifyBG(gtk.STATE_NORMAL, gdk.NewColor("#2A2A2A"))
			label.ModifyFG(gtk.STATE_NORMAL, gdk.NewColor("white"))
			label.SetMarkup(fmt.Sprintf("<b>%d</b>", ws))
		}

		box.Add(label)

		row := i / ix
		col := i - (row * ix)

		urow := uint(row)
		ucol := uint(col)

		// Attach it to the correct place in the table
		table.Attach(box, ucol, ucol+1, urow, urow+1, gtk.EXPAND, gtk.EXPAND, 2, 2)
	}

	window.Add(table)

	g.window = window

	return window
}
