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

func StartDaemon(usr1 func(), usr2 func()) error {
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

	go Watch(usr1, usr2)
	return nil
}

type SignalManager struct {
	pid     int
	process *os.Process
}

func NewSignalManager() *SignalManager {
	return &SignalManager{}
}

func (sm *SignalManager) DStat() (bool, error) {
	data, err := os.ReadFile(lockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read lock file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false, fmt.Errorf("invalid PID in lock file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, nil
	}

	sm.pid = pid
	sm.process = process
	return true, nil
}

func (sm *SignalManager) Send(sig os.Signal) error {
	return sm.process.Signal(sig)
}

func Watch(onUSR1 func(), onUSR2 func()) {
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
			case syscall.SIGUSR2:
				if onUSR2 != nil {
					glib.IdleAdd(func() {
						onUSR2()
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
