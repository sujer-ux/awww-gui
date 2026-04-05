//go:generate go run github.com/lotos-linux/ini/cmd/ini-generator -i $GOFILE -o ../../build
package config

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lotos-linux/ini"
)

// ini:main.conf
type Conf struct {
	General `section:"General"`
}

type General struct {
	BGColor      string `def:"rgba(0,0,0,0)"`
	BorderRadius int    `def:"10"`
	Size         int    `def:"200"`
	Folder       string `def:"~/.config/wallpapers"`
	AwwwFlags    string `def:"--transition-type center --transition-duration 1 --transition-step 90 --transition-fps 60"`
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
