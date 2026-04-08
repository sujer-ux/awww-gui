//go:generate go run github.com/lotos-linux/ini/cmd/ini-generator -i $GOFILE -o ../../build
package config

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lotos-linux/ini"
)

// ini:awww.conf
type Conf struct {
	General `section:"General"`
}

type General struct {
	// Background color
	BGColor string `def:"rgba(0,0,0,0)"`
	// Accent color
	AccentСolor string `def:"rgba(255, 255, 255, 1)"`
	Padding     int    `def:"40"`
	// Preview border radius
	BorderRadius int `def:"10"`
	// Preview height
	Size int `def:"200"`
	// Wallpapers folder
	Folder string `def:"~/.config/wallpapers"`
	// Awww flags
	AwwwFlags string `def:"--transition-type center --transition-duration 1 --transition-step 90 --transition-fps 60"`
}

func New(path string, logger hclog.Logger) (*Conf, error) {
	cm := ini.New(path, logger)

	var conf Conf
	err := cm.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
