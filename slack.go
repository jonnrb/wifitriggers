package wifitriggers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"

	"golang.org/x/net/context/ctxhttp"
)

type SwitchOnSlack struct {
	WebhookURL string
	OnMsg      string
	OffMsg     string

	mu    sync.Mutex
	state switchState
}

func (s *SwitchOnSlack) gotoState(ctx context.Context, n switchState, msg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == n {
		return nil
	}

	if err := msgSlack(ctx, s.WebhookURL, msg); err != nil {
		log.Printf("Error delivering Slack message %q: %v", msg, err)
		s.state = unknownState
	} else {
		log.Printf("Successfully delivered slack message %q", msg)
		s.state = n
	}
	return nil
}

func msgSlack(ctx context.Context, webhookURL, msg string) error {
	b, err := json.Marshal(struct {
		Text string `json:"text"`
	}{
		msg,
	})
	if err != nil {
		panic(err)
	}
	res, err := ctxhttp.Post(ctx, nil, webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (s *SwitchOnSlack) OnAction() Action {
	return func(ctx context.Context) error {
		return s.gotoState(ctx, onState, s.OnMsg)
	}
}

func (s *SwitchOnSlack) OffAction() Action {
	return func(ctx context.Context) error {
		return s.gotoState(ctx, offState, s.OffMsg)
	}
}
