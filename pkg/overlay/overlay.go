package overlay

import (
	"fmt"
	"os"
	"time"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
)

// Spawn creates the grid overlay visualisation. The overlay uses GTK, and runs in a separate
// thread, which is initialised when it is requested.
//
// @todo: Make the colours configurable, probably via environment.
func Spawn(environment grid.Environment, size grid.Size, target float64) <-chan time.Time {
	// Try create the WS grid preview. This should be created in another thread so we can still send
	// the i3-msg commands as quickly as possible.
	go func() {
		gtk.Init(&os.Args)

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
		window.Connect("destroy", gtk.MainQuit)

		// Set main colours
		window.ModifyBG(gtk.STATE_NORMAL, gdk.NewColor("#000000"))
		window.ModifyFG(gtk.STATE_NORMAL, gdk.NewColor("#D3D3D3"))

		table := gtk.NewTable(uint(size.RealX), uint(size.RealY), false)
		table.SetBorderWidth(3)

		labelCount := size.RealX * size.RealY

		for i := 0; i < labelCount; i++ {
			iao := int(environment.ActiveOutputs)
			ico := int(environment.CurrentOutput)

			ws := ico + (iao * i)

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

			row := i / size.RealX
			col := i - (row * size.RealX)

			urow := uint(row)
			ucol := uint(col)

			// Attach it to the correct place in the table
			table.Attach(box, ucol, ucol+1, urow, urow+1, gtk.EXPAND, gtk.EXPAND, 2, 2)
		}

		window.Add(table)
		window.ShowAll()

		gtk.Main()
	}()

	return time.After(500 * time.Millisecond)
}
