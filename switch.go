package wifitriggers

import (
	"context"
	"log"
	"sync"
)

type switchState int

const (
	unknownState switchState = iota
	offState
	onState
)

type SwitchOnIFTTT struct {
	Key    string
	OnCmd  string
	OffCmd string

	mu    sync.Mutex
	state switchState
}

func (s *SwitchOnIFTTT) gotoState(ctx context.Context, n switchState, cmd string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == n {
		return nil
	}

	if err := runIFTTT(ctx, s.Key, cmd); err != nil {
		log.Printf("Error running IFTTT action %q: %v", cmd, err)
		s.state = unknownState
	} else {
		log.Printf("Successfully ran IFTTT action %q", cmd)
		s.state = n
	}
	return nil
}

func (s *SwitchOnIFTTT) OnAction() Action {
	return func(ctx context.Context) error {
		return s.gotoState(ctx, onState, s.OnCmd)
	}
}

func (s *SwitchOnIFTTT) OffAction() Action {
	return func(ctx context.Context) error {
		return s.gotoState(ctx, offState, s.OffCmd)
	}
}
