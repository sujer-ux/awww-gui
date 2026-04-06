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

const CONFIG = ".config/awww-gui/main.conf"

func main() {
	flags := flags.Get()
	logger := createLogger(flags.LogLevel)

	if !flags.Daemon {
		sendToDaemon(&flags, logger)
	}

	path := searchConfig()
	config, err := config.New(path, logger)
	if err != nil {
		logger.Error("Failed to create config", "err", err)
		os.Exit(1)
	}

	gtk.Init(nil)

	temp, _ := gtk.LabelNew("")
	awww := awww.New(config.AwwwFlags)
	imgmanager, err := images.NewManager(config.Folder, config.Size*temp.GetScaleFactor())
	if err != nil {
		logger.Error("images", "err", err)
		os.Exit(2)
	}
	temp.Destroy()

	state := state.New()
	state.SetFlags(&flags)
	state.SetConfig(config)

	wallctl := wallctl.New(awww, imgmanager, logger)
	state.SetWallctl(wallctl)

	app := gui.New(state)

	err = signals.StartDaemon(
		func() {
			if !app.Active() {
				logger.Debug("Open Gui (SIGUSR1)")
				app.Open()
			} else {
				logger.Debug("Close Gui (SIGUSR1)")
				app.Close()
			}
		}, func() {
			logger.Debug("random (SIGUSR2)")
			wallctl.SetRandom()
		})
	if err != nil {
		logger.Error("ErrorStarting", "err", err)
	}

	wallctl.SetCurrent()

	gtk.Main()
}

func sendToDaemon(flags *flags.Flags, logger hclog.Logger) {
	sm := signals.NewSignalManager()

	running, err := sm.DStat()
	if err != nil {
		logger.Error("Failed to get process", "err", err)
		os.Exit(3)
	}

	if !running {
		fmt.Println("Daemon not running")
		fmt.Println("run: setsid awww-gui -d")
		os.Exit(0)
	}

	if flags.Kill {
		sm.Send(syscall.SIGTERM)
		os.Exit(0)
	}

	if flags.Random {
		sm.Send(syscall.SIGUSR2)
		os.Exit(0)
	}

	sm.Send(syscall.SIGUSR1)
	os.Exit(0)
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
	prod := filepath.Join(os.Getenv("HOME"), CONFIG)
	_, err := os.Stat(prod)
	if err == nil {
		return prod
	}

	exe, _ := os.Executable()
	dev := filepath.Join(filepath.Dir(exe), "main.conf")
	return dev
}
