package gui

import (
	"awww-gui/internal/state"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gtk"
)

type App struct {
	state  *state.State
	window *gtk.Window
}

func New(state *state.State) *App {
	return &App{
		state:  state,
		window: nil,
	}
}

func (a *App) Open() error {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}

	layershell.InitForWindow(window)
	layershell.SetNamespace(window, "awww-gui")
	layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetExclusiveZone(window, -1)

	layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_TOP, false)
	layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_BOTTOM, false)

	layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_RIGHT, true)

	window.SetSizeRequest(50, 100)

	splash, err := gtk.LabelNew("Test")
	if err != nil {
		return err
	}

	splash.SetJustify(gtk.JUSTIFY_CENTER)

	window.Add(splash)
	window.ShowAll()
	a.window = window
	return nil
}

func (a *App) Close() {
	a.window.Destroy()
	a.window = nil
}

func (a *App) Active() bool {
	return a.window != nil
}
