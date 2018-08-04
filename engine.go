package wifitriggers

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
)

type Engine struct {
	bindings []Binding
}

type EngineBuilder struct {
	e Engine
}

type If Cond

type Binding struct {
	Cond
	Action
}

func (i If) Then(action Action) Binding {
	return Binding{Cond(i), action}
}

func (eb *EngineBuilder) Bind(b Binding) *EngineBuilder {
	eb.e.bindings = append(eb.e.bindings, b)
	return eb
}

func (eb *EngineBuilder) Build() *Engine {
	return &eb.e
}

func (e *Engine) Run(ctx context.Context, connectedClients []net.HardwareAddr) error {
	grp, ctx := errgroup.WithContext(ctx)
	for _, b := range e.bindings {
		cond, action := b.Cond, b.Action
		grp.Go(func() error {
			if !cond(connectedClients) {
				return nil
			}
			return action(ctx)
		})
	}
	return grp.Wait()
}
