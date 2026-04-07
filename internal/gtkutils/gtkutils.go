package gtkutils

import (
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

func CreateImageFromFile(source string, height int) (*gtk.Image, error) {
	image, err := CreateImage0(source, height)
	if err == nil {
		return image, nil
	}
	return CreateImage0("image-missing", height)
}

func CreateImage0(source string, height int) (*gtk.Image, error) {
	var err error
	var pixbuf *gdk.Pixbuf

	image, err := gtk.ImageNew()
	if err != nil {
		return nil, err
	}

	scaleFactor := image.GetScaleFactor()
	physicalHeight := height * scaleFactor

	pixbuf, err = gdk.PixbufNewFromFileAtScale(source, -1, physicalHeight, true)
	if err != nil {
		return nil, err
	}

	surface, err := gdk.CairoSurfaceCreateFromPixbuf(pixbuf, scaleFactor, nil)
	if err != nil {
		return nil, err
	}

	image.SetFromSurface(surface)

	pixbufWidth := pixbuf.GetWidth()
	pixbufHeight := pixbuf.GetHeight()

	widgetHeight := height
	widgetWidth := (pixbufWidth * widgetHeight) / pixbufHeight

	image.SetSizeRequest(widgetWidth, widgetHeight)

	return image, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
