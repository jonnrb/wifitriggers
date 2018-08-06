package wifitriggers

import (
	"context"
	"log"
	"time"
)

type Driver struct {
	APReader     APReader
	BindingChain BindingChain
	Interval     time.Duration
}

func tickDuringContext(ctx context.Context, d time.Duration) <-chan time.Time {
	t := time.NewTicker(d)
	go func() {
		<-ctx.Done()
		t.Stop()
	}()
	return t.C
}

func (d *Driver) Run(ctx context.Context) error {
	for _ = range tickDuringContext(ctx, d.Interval) {
		err := func() error {
			ctx, cancel := context.WithTimeout(ctx, d.Interval)
			defer cancel()

			c, err := d.APReader.ConnectedClients(ctx)

			// Treat errors as nobody connected for now.
			// TODO(jonnrb): Introduce some sort of transient error state.
			if err != nil {
				log.Println("Could not get connected clients from AP:", err)
			}

			if err := d.BindingChain(c)(ctx); err != nil {
				log.Println("Error running actions:", err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return ctx.Err()
}
