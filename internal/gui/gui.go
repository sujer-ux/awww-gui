package gui

import (
	"awww-gui/internal/current"
	"awww-gui/internal/gtkutils"
	"awww-gui/internal/images"
	"awww-gui/internal/state"
	"awww-gui/internal/wallctl"
	"fmt"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

type App struct {
	state    *state.State
	window   *gtk.Window
	viewport *gtk.Box
	wrapper  *gtk.FlowBox
	wallctl  *wallctl.Control
	pos      int
	logger   hclog.Logger
}

const IMAGEMARGIN = 3
const PRESCROLLZONE = 30
const SPACING = 20
const SCROLLPADDING = 50

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
	err = a.wallctl.Rescan()
	if err != nil {
		a.logger.Error("Failed to rescan", "Error", err)
	}

	a.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}
	a.initLayerShell()
	a.styling()

	a.viewport, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	a.viewport.SetName("viewport")

	a.wrapper, err = gtk.FlowBoxNew()
	if err != nil {
		return err
	}

	a.wrapper.SetName("wrapper")
	a.wrapper.SetColumnSpacing(SPACING)
	a.wrapper.SetMarginTop(30)
	a.wrapper.SetMarginBottom(20)
	a.wrapper.SetMarginStart(SPACING)
	a.wrapper.SetMarginEnd(SPACING)
	a.wrapper.SetCanFocus(true)
	a.wrapper.SetSelectionMode(gtk.SELECTION_SINGLE)
	a.wrapper.SetHomogeneous(false)
	a.wrapper.SetHAlign(gtk.ALIGN_START)
	a.wrapper.SetVAlign(gtk.ALIGN_START)
	a.wrapper.SetProperty("orientation", uint(gtk.ORIENTATION_VERTICAL))

	a.wrapper.Connect("key-press-event", func(_ interface{}, event *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		if keyEvent.KeyVal() == gdk.KEY_Escape {
			a.Close()
			return true
		}
		return false
	})

	a.wrapper.Connect("selected-children-changed", func() {
		selected := a.wrapper.GetSelectedChildren()
		if len(selected) > 0 {
			child := selected[0]
			a.move(child)
		}
	})

	current, cErr := current.Get()
	list := a.state.GetWallctl().GetList()
	for _, image := range list {
		child := a.createChild(image)
		if cErr == nil && current == image.Name {
			a.wrapper.SelectChild(child)
			child.GrabFocus()
		}
		a.wrapper.Add(child)
	}

	a.wrapper.SetMaxChildrenPerLine(uint(len(list)))
	a.viewport.Add(a.wrapper)
	a.window.Add(a.viewport)
	a.window.ShowAll()
	return nil
}

func (a *App) Close() {
	a.pos = 0
	a.window.Destroy()
	a.window = nil
}

func (a *App) Active() bool {
	return a.window != nil
}

func (a *App) createChild(image *images.Image) *gtk.FlowBoxChild {
	container, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	container.SetHAlign(gtk.ALIGN_START)
	container.SetVAlign(gtk.ALIGN_START)
	container.SetCanFocus(true)

	preview, err := gtkutils.CreateImageFromFile(
		image.Thumbnail,
		a.state.GetConfig().Size,
		a.state.GetConfig().BorderRadius,
	)
	if err != nil {
		a.logger.Error("Failed to create gtk.Image", "Error", err)
		return nil
	}

	preview.SetHAlign(gtk.ALIGN_CENTER)
	preview.SetMarginBottom(IMAGEMARGIN)
	preview.SetMarginTop(IMAGEMARGIN)
	preview.SetMarginEnd(IMAGEMARGIN)
	preview.SetMarginStart(IMAGEMARGIN)

	var finalWidget gtk.IWidget
	if image.Format == "gif" {
		overlay, err := gtk.OverlayNew()
		if err != nil {
			return nil
		}

		overlay.Add(preview)

		label, _ := gtk.LabelNew("GIF")
		label.SetName("gif-label")

		label.SetHAlign(gtk.ALIGN_END)
		label.SetVAlign(gtk.ALIGN_START)
		label.SetMarginTop(10)
		label.SetMarginEnd(10)

		overlay.AddOverlay(label)

		finalWidget = overlay
	} else {
		finalWidget = preview
	}

	preWrap, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	preWrap.Add(finalWidget)
	preWrap.SetMarginBottom(10)
	container.Add(preWrap)

	title, _ := gtk.LabelNew(image.Name)
	title.SetHAlign(gtk.ALIGN_CENTER)
	container.Add(title)

	child, _ := gtk.FlowBoxChildNew()
	child.SetName("child")
	child.SetHAlign(gtk.ALIGN_START)
	child.SetVAlign(gtk.ALIGN_START)
	child.SetHExpand(false)
	child.SetVExpand(false)
	child.Connect("activate", func() {
		current, _ := current.Get()
		if current != image.Name {
			a.wallctl.Set(image.Name)
		}
		a.Close()
	})

	child.Add(container)
	return child
}

func (a *App) move(child *gtk.FlowBoxChild) {
	allocation := child.GetAllocation()
	windowWidth, err := getMonitorWidthForWindow(a.window)
	if err != nil {
		a.logger.Error("Failed to get monitor", "Error", err)
		windowWidth = 1360
	}

	start := allocation.GetX()
	end := allocation.GetX() + allocation.GetWidth()

	currentStart := start + a.pos
	currentEnd := end + a.pos

	left := currentStart < PRESCROLLZONE
	right := currentEnd > windowWidth-PRESCROLLZONE

	if left {
		a.pos = SCROLLPADDING - start

		if a.pos > 0 {
			a.pos = 0
		}

		gtkutils.AddStyle(a.viewport, fmt.Sprintf("#viewport {margin-left: %dpx}", a.pos))
	}

	if right {
		a.pos = windowWidth - SCROLLPADDING - end - SPACING
		gtkutils.AddStyle(a.viewport, fmt.Sprintf("#viewport {margin-left: %dpx}", a.pos))
	}
}

func (a *App) styling() error {
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}

	config := a.state.GetConfig()
	css := fmt.Sprintf(`
		window {
			background: %s;
		}

        #child {
            padding: 5px;
        }
        
        #child:selected {
            outline: none;
            box-shadow: none;
			background: transparent;
			color: #ffffff;
        }
        
        #child:selected > box > box {
            box-shadow: 0 0 0 2px %s; 
            border-radius: %dpx;
        }

		#gif-label {
			background: #ffffff;
			color: #242424;
			padding: 3px 5px;
			border-radius: 5px;
			font-weight: 800;
			font-size: 13px;
		}
    `, config.BGColor, config.AccentСolor, config.BorderRadius+IMAGEMARGIN)

	cssProvider.LoadFromData(css)
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return err
	}
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return nil
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
	layershell.SetKeyboardMode(a.window, layershell.LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND)
}

func getMonitorWidthForWindow(window *gtk.Window) (int, error) {
	display, err := gdk.DisplayGetDefault()
	if err != nil {
		return 0, err
	}

	gdkWin, err := window.GetWindow()
	if err != nil {
		return 0, err
	}

	monitor, err := display.GetMonitorAtWindow(gdkWin)
	if err != nil {
		return 0, err
	}

	geometry := monitor.GetGeometry()
	return geometry.GetWidth(), nil
}
