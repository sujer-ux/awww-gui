package gui

import (
	"awww-gui/internal/state"
	"fmt"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gtk"
)

func Open(state *state.State) error {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}

	layershell.InitForWindow(window)
	layershell.SetNamespace(window, "awww-gui")
	layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_TOP)

	fmt.Println("asdasdasфы123ыффффффффффф")
	window.SetDefaultSize(300, 300)

	splash, err := gtk.LabelNew("Test")
	if err != nil {
		return err
	}

	window.Add(splash)
	window.ShowAll()
	return nil
}
