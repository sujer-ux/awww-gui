package signals

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const lockFile = "/tmp/awww-gui.lock"

func StartDaemon(usr func()) error {
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return fmt.Errorf("daemon already running")
	}

	f.Truncate(0)
	f.Seek(0, 0)
	fmt.Fprintf(f, "%d", os.Getpid())
	f.Sync()

	go Watch(usr)
	return nil
}

func SendSignal() error {
	data, err := os.ReadFile(lockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daemon not running")
		}
		return fmt.Errorf("failed to read lock file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID in lock file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("daemon not running (stale lock file)")
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("daemon not running (process %d is dead)", pid)
	}

	return process.Signal(syscall.SIGUSR1)
}

func Watch(onUSR1 func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGUSR1:
				if onUSR1 != nil {
					glib.IdleAdd(func() {
						onUSR1()
					})
				}
			case syscall.SIGINT, syscall.SIGTERM:
				glib.IdleAdd(func() {
					gtk.MainQuit()
				})
			}
		}
	}()
}
