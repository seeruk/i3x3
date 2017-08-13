package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

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

		s.overlay.CreateTable(req.Direction)
		s.overlay.AttachTable()
		s.overlay.ShowTable()
		s.overlay.ShowWindow()

		timer := time.NewTimer(5 * time.Second)

		select {
		case result := <-timer.C:
			log.Printf("Hiding from timer: %v\n", result)
			s.overlay.HideWindow()
			s.overlay.DestroyTable()
		case <-s.overlay.interrupt:
			log.Println("Hiding from interrupt")
			timer.Stop()
			s.overlay.DestroyTable()
		}
	}()

	return &proto.SwitchWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) RedistributeWorkspaces(ctx context.Context, req *proto.RedistributeWorkspacesRequest) (*proto.RedistributeWorkspacesResponse, error) {
	log.Printf("Redistribute workspace request received")

	return &proto.RedistributeWorkspacesResponse{}, nil
}

func main() {
	start := time.Now()

	// Initialise GLib, GDK, and GTK
	glib.ThreadInit(nil)
	gdk.ThreadsInit()
	gdk.ThreadsEnter()
	gtk.Init(&os.Args)

	elapsed := time.Since(start)

	fmt.Println("Took %s", elapsed)

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

	overlay.CreateWindow()

	server := grpc.NewServer()
	proto.RegisterWorkspaceServer(server, &WorkspaceServer{
		overlay: overlay,
	})
	server.Serve(conn)
}

type GridOverlay struct {
	sync.Mutex

	interrupt chan bool

	table  *gtk.Table
	window *gtk.Window
}

func (g *GridOverlay) Interrupt() {
	if g.table == nil {
		return
	}

	g.interrupt <- true
}

func (g *GridOverlay) CreateTable(direction proto.Direction) *gtk.Table {
	ix := 3
	iy := 3

	co := 1
	no := 2

	target := 5

	switch direction {
	case proto.Direction_UP:
		target = 3
	case proto.Direction_DOWN:
		target = 15
	case proto.Direction_LEFT:
		target = 7
	case proto.Direction_RIGHT:
		target = 11
	}

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

	g.table = table

	return table
}

func (g *GridOverlay) ShowTable() {
	if g.table == nil {
		return
	}

	gdk.ThreadsEnter()
	g.table.ShowAll()
	gdk.ThreadsLeave()
}

func (g *GridOverlay) AttachTable() {
	if g.table == nil || g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.Add(g.table)
	gdk.ThreadsLeave()
}

func (g *GridOverlay) DetachTable() {
	if g.table == nil || g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.Remove(g.table)
	gdk.ThreadsLeave()
}

func (g *GridOverlay) DestroyTable() {
	log.Println("Destroying table")

	if g.table == nil {
		return
	}

	gdk.ThreadsEnter()
	g.table.Destroy()
	gdk.ThreadsLeave()

	g.table = nil
}

func (g *GridOverlay) CreateWindow() *gtk.Window {
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

	g.window = window

	return window
}

func (g *GridOverlay) ShowWindow() {
	if g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.ShowAll()
	gdk.ThreadsLeave()
}

func (g *GridOverlay) HideWindow() {
	if g.window == nil {
		return
	}

	gdk.ThreadsEnter()
	g.window.Hide()
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
