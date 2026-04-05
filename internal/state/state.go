package state

import (
	"awww-gui/internal/awww"
	"awww-gui/internal/config"
	"awww-gui/internal/flags"
	"awww-gui/internal/images"
	"sync"
)

type State struct {
	mu     *sync.RWMutex
	images *images.Manager
	config *config.Conf
	flags  *flags.Flags
	awww   *awww.Awww
}

func New() *State {
	return &State{
		mu: &sync.RWMutex{},
	}
}

func (s *State) GetImages() *images.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.images
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

func (s *State) GetAwww() *awww.Awww {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.awww
}

func (s *State) SetImages(images *images.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.images = images
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

func (s *State) SetAwww(awww *awww.Awww) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.awww = awww
}
