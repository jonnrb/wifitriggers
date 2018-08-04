package wifitriggers

import (
	"context"
	"path"

	"golang.org/x/net/context/ctxhttp"
)

func runIFTTT(ctx context.Context, key, action string) error {
	res, err := ctxhttp.Get(ctx, nil, "https://maker.ifttt.com/"+path.Join("trigger", action, "with/key", key))
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}
