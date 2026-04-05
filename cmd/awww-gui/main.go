package main

import (
	"awww-gui/internal/awww"
	"awww-gui/internal/config"
	"awww-gui/internal/current"
	"awww-gui/internal/flags"
	"awww-gui/internal/gui"
	"awww-gui/internal/images"
	"awww-gui/internal/signals"
	"awww-gui/internal/state"
	"log"
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

const CONFIG = ".config/awww-gui/main.conf"

func main() {
	flags := flags.Get()
	logger := createLogger(flags.LogLevel)

	if !flags.Deamon {
		err := signals.SendSignal()
		if err != nil {
			logger.Error("Errror send signal", "err", err)
			os.Exit(3)
		}
		os.Exit(0)
	}

	config, err := config.New(filepath.Join(os.Getenv("HOME"), CONFIG), logger)
	if err != nil {
		logger.Error("creating config", "err", err)
		os.Exit(1)
	}

	awww := awww.New(config.AwwwFlags)

	gtk.Init(nil)

	temp, _ := gtk.LabelNew("")
	imgmanager, err := images.NewManager(config.Folder, config.Size*temp.GetScaleFactor())
	if err != nil {
		logger.Error("images", "err", err)
		os.Exit(2)
	}
	temp.Destroy()

	state := state.New()
	state.SetFlags(&flags)
	state.SetConfig(config)
	state.SetAwww(awww)
	state.SetImages(imgmanager)

	app := gui.New(state)

	err = signals.StartDaemon(func() {
		if !app.Active() {
			logger.Debug("Open Gui (SIGUSR1)")
			app.Open()
		} else {
			logger.Debug("Close Gui (SIGUSR1)")
			app.Close()
		}
	})
	if err != nil {
		logger.Error("ErrorStarting", "err", err)
	}

	initAwww(awww, imgmanager)

	gtk.Main()
}

func initAwww(awww *awww.Awww, imgmanager *images.Manager) {
	awww.Init()
	var name string

	name, err := current.Get()
	if err != nil {
		name = imgmanager.Random().Name
		current.Set(name)
	}

	image := imgmanager.Get(name)
	awww.Set(image.Original)
	log.Println(current.Get())
}

func createLogger(logLevel string) hclog.Logger {
	level := hclog.LevelFromString(logLevel)

	if level == hclog.NoLevel {
		level = hclog.Info
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:  "awww-gui",
		Level: level,
		Color: hclog.AutoColor,
	})
}
