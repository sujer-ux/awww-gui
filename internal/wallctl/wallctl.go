package wallctl

import (
	"awww-gui/internal/awww"
	"awww-gui/internal/current"
	"awww-gui/internal/images"

	"github.com/hashicorp/go-hclog"
)

type Control struct {
	awww    *awww.Awww
	imgmngr *images.Manager
	logger  hclog.Logger
}

func New(awww *awww.Awww, imgmngr *images.Manager, logger hclog.Logger) (*Control, error) {
	err := awww.Init()
	if err != nil {
		return nil, err
	}

	return &Control{
		awww:    awww,
		imgmngr: imgmngr,
		logger:  logger,
	}, nil
}

func (c *Control) GetList() []*images.Image {
	return c.imgmngr.GetList()
}

func (c *Control) Rescan() error {
	return c.imgmngr.Rescan()
}

func (c *Control) SetRandom() {
	image := c.imgmngr.Random()
	if image == nil {
		c.logger.Error("No wallpapers available")
		return
	}
	c.Set(image.Name)
}

func (c *Control) SetCurrent() {
	var name string

	name, err := current.Get()
	if err != nil {
		c.SetRandom()
		return
	}

	c.Set(name)
}

func (c *Control) Set(name string) {
	image := c.imgmngr.Get(name)
	if image == nil {
		c.logger.Error("Image not found")
		return
	}

	err := c.awww.Set(image.Original)
	if err != nil {
		c.logger.Error("Failed to set wallpaper",
			"name", image.Name,
			"format", image.Format,
			"image", image.Original,
			"err", err,
		)
		return
	}

	c.logger.Trace("Wallpaper successfully set",
		"name", image.Name,
		"format", image.Format,
		"image", image.Original,
	)

	err = current.Set(name)
	if err != nil {
		c.logger.Error("Failed to persist wallpaper name", "Error", err)
	}
}
