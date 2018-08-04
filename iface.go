package wifitriggers

import (
	"context"
	"net"
)

type APReader interface {
	ConnectedClients(ctx context.Context) ([]net.HardwareAddr, error)
}

type Cond func(connectedClients []net.HardwareAddr) bool

type Action func(ctx context.Context) error
