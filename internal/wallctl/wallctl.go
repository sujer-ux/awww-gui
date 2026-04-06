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

func New(awww *awww.Awww, imgmngr *images.Manager, logger hclog.Logger) *Control {
	awww.Init()

	return &Control{
		awww:    awww,
		imgmngr: imgmngr,
		logger:  logger,
	}
}

func (c *Control) GetList() []*images.Image {
	return c.imgmngr.GetList()
}

func (c *Control) SetRandom() {
	image := c.imgmngr.Random()
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
	} else {
		c.logger.Trace("Wallpaper successfully set",
			"name", image.Name,
			"format", image.Format,
			"image", image.Original,
		)
		current.Set(name)
	}
}
