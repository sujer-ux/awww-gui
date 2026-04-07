package state

import (
	"awww-gui/internal/config"
	"awww-gui/internal/flags"
	"awww-gui/internal/wallctl"
	"sync"

	"github.com/hashicorp/go-hclog"
)

type State struct {
	mu      *sync.RWMutex
	config  *config.Conf
	flags   *flags.Flags
	wallctl *wallctl.Control
	logger  hclog.Logger
}

func New() *State {
	return &State{
		mu: &sync.RWMutex{},
	}
}

func (s *State) GetConfig() *config.Conf {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *State) GetFlags() *flags.Flags {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.flags
}

func (s *State) GetWallctl() *wallctl.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wallctl
}

func (s *State) GetLogger() hclog.Logger {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logger
}

func (s *State) SetConfig(config *config.Conf) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

func (s *State) SetFlags(flags *flags.Flags) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags = flags
}

func (s *State) SetWallctl(wallctl *wallctl.Control) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wallctl = wallctl
}

func (s *State) SetLogger(logger hclog.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
}
