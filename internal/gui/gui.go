package gui

import (
	"awww-gui/internal/gtkutils"
	"awww-gui/internal/images"
	"awww-gui/internal/state"
	"awww-gui/internal/wallctl"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

type App struct {
	state   *state.State
	window  *gtk.Window
	wrapper *gtk.FlowBox
	wallctl *wallctl.Control
	logger  hclog.Logger
}

func New(state *state.State) *App {
	return &App{
		state:   state,
		window:  nil,
		wallctl: state.GetWallctl(),
		logger:  state.GetLogger().Named("gui"),
	}
}

func (a *App) Open() error {
	var err error

	a.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}
	a.initLayerShell()

	a.wrapper, err = gtk.FlowBoxNew()
	if err != nil {
		return err
	}

	a.wrapper.SetCanFocus(true)
	a.wrapper.SetSelectionMode(gtk.SELECTION_SINGLE)

	a.wrapper.Connect("key-press-event", func(_ interface{}, event *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		if keyEvent.KeyVal() == gdk.KEY_Escape {
			a.Close()
			return true
		}
		return false
	})

	list := a.state.GetWallctl().GetList()
	for _, image := range list {
		a.createChild(image)
	}

	a.window.Add(a.wrapper)
	a.window.ShowAll()
	return nil
}

func (a *App) Close() {
	a.window.Destroy()
	a.window = nil
}

func (a *App) Active() bool {
	return a.window != nil
}

func (a *App) createChild(image *images.Image) {
	container, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	container.SetCanFocus(true)

	preveiw, err := gtkutils.CreateImageFromFile(image.Thumbnail, a.state.GetConfig().Size)
	if err != nil {
		a.logger.Error("Не удалось создать изображение", "Error", err)
		return
	}
	container.Add(preveiw)

	title, _ := gtk.LabelNew(image.Name)
	container.Add(title)

	child, _ := gtk.FlowBoxChildNew()
	child.Connect("activate", func() {
		a.wallctl.Set(image.Name)
		a.Close()
	})

	child.Add(container)
	a.wrapper.Add(child)
}

func (a *App) initLayerShell() {
	layershell.InitForWindow(a.window)
	layershell.SetNamespace(a.window, "awww-gui")
	layershell.SetLayer(a.window, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetExclusiveZone(a.window, -1)
	layershell.SetAnchor(a.window, layershell.LAYER_SHELL_EDGE_TOP, false)
	layershell.SetAnchor(a.window, layershell.LAYER_SHELL_EDGE_BOTTOM, false)
	layershell.SetAnchor(a.window, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetAnchor(a.window, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	layershell.SetKeyboardMode(a.window, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
}
