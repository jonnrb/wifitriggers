package wifitriggers

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
)

type APReader interface {
	ConnectedClients(ctx context.Context) ([]net.HardwareAddr, error)
}

type Cond func(connectedClients []net.HardwareAddr) bool

// Takes the logical OR of two Conds. E.g.:
//
//   condA.Or(condB)
//
func (a Cond) Or(b Cond) Cond {
	return func(connectedClients []net.HardwareAddr) bool {
		return a(connectedClients) || b(connectedClients)
	}
}

// Takes the logical AND of two Conds. E.g.:
//
//   condA.And(condB)
//
func (a Cond) And(b Cond) Cond {
	return func(connectedClients []net.HardwareAddr) bool {
		return a(connectedClients) && b(connectedClients)
	}
}

type Action func(ctx context.Context) error

type key int

var actionGroupKey key

// Returns an action that runs the two actions it is comprised of concurrently.
// golang.org/x/sync/errgroup's semantics are used: the context will be
// cancelled if any member action fails and the returned error will be the
// error that caused the context to be cancelled.
//
//     actionA.And(actionB).And(actionC.And(actionD))
//
func (a Action) And(b Action) Action {
	return func(ctx context.Context) error {
		g, ok := ctx.Value(actionGroupKey).(*errgroup.Group)
		if ok {
			g.Go(func() error { return a(ctx) })
			return b(ctx)
		}

		g, ctx = errgroup.WithContext(ctx)
		ctx = context.WithValue(ctx, actionGroupKey, g)

		g.Go(func() error { return a(ctx) })
		g.Go(func() error { return b(ctx) })

		return g.Wait()
	}
}
