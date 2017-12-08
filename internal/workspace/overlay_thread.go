package workspace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SeerUK/i3x3/internal/grid"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/inconshreveable/log15"
)

const OverlayDuration = 500 * time.Millisecond

// OverlayThread is a long-running process than handles showing the GTK-based overlay.
type OverlayThread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger
	window *gtk.Window

	msgCh <-chan SwitchMessage
}

// NewOverlayThread creates a new workspace overlay thread.
func NewOverlayThread(logger log15.Logger, msgCh <-chan SwitchMessage) *OverlayThread {
	logger = logger.New("module", "workspace/overlayThread")

	return &OverlayThread{
		logger: logger,
		msgCh:  msgCh,
	}
}

// Start attempts to start the overlay thread.
func (t *OverlayThread) Start() error {
	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.Unlock()

	defer func() {
		t.Lock()
		t.ctx = nil
		t.cfn = nil
		t.window = nil
		t.Unlock()
	}()

	// Set up GTK
	gtk.Init(nil)

	t.Lock()
	t.window = buildWindow()
	t.Unlock()

	// Use dark theme.
	settings, _ := gtk.SettingsGetDefault()
	settings.SetProperty("gtk-application-prefer-dark-theme", true)

	reaperChan := make(chan struct{})

	go t.enqueueMessages(reaperChan)
	go t.windowReaper(reaperChan)

	t.logger.Info("thread started")

	// This is a blocking call.
	gtk.Main()

	t.logger.Info("thread stopped")

	return nil
}

// Stop attempts to stop the overlay thread.
func (t *OverlayThread) Stop() error {
	gtk.MainQuit()
	return nil
}

// handleMessage takes a message and updates the window UI appropriately, finally showing the
// window (if it's not already visible) at the end.
func (t *OverlayThread) handleMessage(window *gtk.Window, msg SwitchMessage) bool {
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

	size := grid.NewSize(msg.Environment, 3, 3)

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
		iao := int(msg.Environment.ActiveOutputs)
		ico := int(msg.Environment.CurrentOutput)

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
		if int(msg.Target) == ws {
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

// enqueueMessages routes incoming and outgoing messages, handling updating the overlay window.
func (t *OverlayThread) enqueueMessages(reaperCh chan<- struct{}) {
	for {
		select {
		case message := <-t.msgCh:
			// Show the overlay
			glib.IdleAdd(t.handleMessage, t.window, message)

			// Notify other threads.
			reaperCh <- struct{}{}
			message.ResponseCh <- nil
		case <-t.ctx.Done():
			break
		}
	}
}

// windowReaper waits for a set amount of time before hiding the overlay window. If another message
// comes in whilst the window is open, it's life is extended.
func (t *OverlayThread) windowReaper(messages <-chan struct{}) {
	var timer *time.Timer

	for {
		select {
		case <-messages:
			if timer != nil {
				timer.Stop()
			}

			timer = time.AfterFunc(OverlayDuration, func() {
				glib.IdleAdd(t.window.Hide)
			})
		case <-t.ctx.Done():
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
