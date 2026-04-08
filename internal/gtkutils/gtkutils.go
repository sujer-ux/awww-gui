package gtkutils

import (
	"math"
	"os"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

func CreateImageFromFile(source string, height int, radius int) (*gtk.Image, error) {
	image, err := CreateImage0(source, height, radius)
	if err == nil {
		return image, nil
	}
	return CreateImage0("image-missing", height, radius)
}

func CreateImage0(source string, height int, radius int) (*gtk.Image, error) {
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

	if radius > 0 {
		width := surface.GetWidth()
		height := surface.GetHeight()

		roundedSurface := surface.CreateSimilar(cairo.CONTENT_COLOR_ALPHA, width, height)

		cr := cairo.Create(roundedSurface)

		cr.NewPath()
		roundedRect(cr, float64(width), float64(height), float64(radius))
		cr.Clip()

		cr.SetSourceSurface(surface, 0, 0)
		cr.Paint()

		surface = roundedSurface
	}

	image.SetFromSurface(surface)

	pixbufWidth := pixbuf.GetWidth()
	pixbufHeight := pixbuf.GetHeight()

	widgetHeight := height
	widgetWidth := (pixbufWidth * widgetHeight) / pixbufHeight

	image.SetSizeRequest(widgetWidth, widgetHeight)

	return image, nil
}

func roundedRect(cr *cairo.Context, width, height, radius float64) {
	if radius > width/2 {
		radius = width / 2
	}
	if radius > height/2 {
		radius = height / 2
	}

	cr.MoveTo(radius, 0)
	cr.LineTo(width-radius, 0)
	cr.Arc(width-radius, radius, radius, -math.Pi/2, 0)
	cr.LineTo(width, height-radius)
	cr.Arc(width-radius, height-radius, radius, 0, math.Pi/2)
	cr.LineTo(radius, height)
	cr.Arc(radius, height-radius, radius, math.Pi/2, math.Pi)
	cr.LineTo(0, radius)
	cr.Arc(radius, radius, radius, math.Pi, 3*math.Pi/2)
	cr.ClosePath()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func AddStyle(widget gtk.IWidget, style string) (*gtk.CssProvider, error) {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return nil, err
	}

	err = provider.LoadFromData(style)
	if err != nil {
		return nil, err
	}

	context, err := widget.ToWidget().GetStyleContext()
	if err != nil {
		return nil, err
	}

	context.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return provider, nil
}
