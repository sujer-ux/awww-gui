package main

import (
	"awww-gui/internal/awww"
	"awww-gui/internal/config"
	"awww-gui/internal/flags"
	"awww-gui/internal/gui"
	"awww-gui/internal/images"
	"awww-gui/internal/signals"
	"awww-gui/internal/state"
	"awww-gui/internal/wallctl"

	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

const CONFIGPATH = ".config/awww-gui"
const CONFIGFILE = "awww.conf"

func main() {
	var err error

	flags := flags.Get()
	logger := createLogger(flags.LogLevel)

	// client mode
	if !flags.Daemon {
		err = sendToDaemon(flags)
		if err != nil {
			logger.Error("Failed to send signal", "Error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// daemon mode
	// config
	path := searchConfig()
	config, err := config.New(path, logger)
	if err != nil {
		logger.Error("Failed to create config", "Error", err)
		os.Exit(2)
	}

	// gtk start
	gtk.Init(nil)

	// image manager
	temp, _ := gtk.LabelNew("")
	awww := awww.New(config.AwwwFlags)
	imgmanager, err := images.New(config.Folder, config.Size*temp.GetScaleFactor(), logger)
	if err != nil {
		logger.Error("Failed to create image manager", "Error", err)
		os.Exit(3)
	}
	temp.Destroy()

	// wallpaper manager
	wallctl, err := wallctl.New(awww, imgmanager, logger)
	if err != nil {
		logger.Error("Failed to start awww-daemon", "Error", err)
		os.Exit(4)
	}

	// create state
	state := state.New()
	state.SetLogger(logger)
	state.SetFlags(&flags)
	state.SetConfig(config)
	state.SetWallctl(wallctl)

	// create gui
	app := gui.New(state)

	// wath signals
	err = signals.StartDaemon(
		// SIGUSR1
		func() {
			if !app.Active() {
				logger.Debug("Open Gui")
				app.Open()
			} else {
				logger.Debug("Close Gui")
				app.Close()
			}
		},
		// SIGUSR2
		func() {
			logger.Debug("Random wallpapers")
			wallctl.SetRandom()
		})
	if err != nil {
		logger.Error("Starting daemon failed", "Error", err)
		os.Exit(5)
	}

	// set last wallpapers
	wallctl.SetCurrent()
	gtk.Main()
}

func sendToDaemon(flags flags.Flags) error {
	sm := signals.NewSignalManager()

	running, err := sm.DStat()
	if err != nil {
		return fmt.Errorf("failed to get process: %v", err)
	}

	if !running {
		fmt.Println("Daemon not running")
		fmt.Println("run: setsid awww-gui -d")
		return nil
	}

	if flags.Kill {
		return sm.Send(syscall.SIGTERM)
	}

	if flags.Random {
		return sm.Send(syscall.SIGUSR2)
	}

	return sm.Send(syscall.SIGUSR1)
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

func searchConfig() string {
	prod := filepath.Join(os.Getenv("HOME"), CONFIGPATH, CONFIGFILE)
	_, err := os.Stat(prod)
	if err == nil {
		return prod
	}

	exe, _ := os.Executable()
	dev := filepath.Join(filepath.Dir(exe), CONFIGFILE)
	return dev
}
