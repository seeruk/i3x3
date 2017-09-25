package overlay

import (
	"fmt"
	"os"
	"time"

	"github.com/SeerUK/i3x3/pkg/grid"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// Spawn creates the grid overlay visualisation. The overlay uses GTK, and runs in a separate
// thread, which is initialised when it is requested.
func Spawn(environment grid.Environment, size grid.Size, target float64) <-chan time.Time {
	// Try create the WS grid preview. This should be created in another thread so we can still send
	// the i3-msg commands as quickly as possible.
	go func() {
		gtk.Init(&os.Args)

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
		window.Connect("destroy", gtk.MainQuit)

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

		gtk.Main()
	}()

	return time.After(500 * time.Millisecond)
}
