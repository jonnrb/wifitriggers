package wifitriggers

import (
	"context"
	"net"

	"go.jonnrb.io/hostapd_grpc/proto"
	"golang.org/x/sync/errgroup"
)

type HostapdReader struct {
	hostapd.HostapdControlClient
}

func (h HostapdReader) ConnectedClients(ctx context.Context) ([]net.HardwareAddr, error) {
	return connectedMACs(ctx, h.HostapdControlClient)
}

func connectedMACs(ctx context.Context, cli hostapd.HostapdControlClient) ([]net.HardwareAddr, error) {
	sockets, err := cli.ListSockets(ctx, &hostapd.ListSocketsRequest{})
	if err != nil {
		return nil, err
	}

	staLists := make(chan []*hostapd.Client)
	g, ctx := errgroup.WithContext(ctx)
	for _, s := range sockets.Socket {
		sockName := s.Name
		g.Go(func() error {
			res, err := cli.ListClients(ctx, &hostapd.ListClientsRequest{
				SocketName: []string{sockName},
			})
			if err != nil {
				return err
			} else {
				staLists <- res.Client
				return nil
			}
		})
	}
	go func() {
		g.Wait()
		close(staLists)
	}()

	// Create set to dedup possibly inconsistent info from a station switching
	// access points.
	addrSet := make(map[string]net.HardwareAddr)
	for staList := range staLists {
		for _, sta := range staList {
			hwAddr, err := net.ParseMAC(sta.Addr)
			if err != nil {
				return nil, err
			}
			addrSet[hwAddr.String()] = hwAddr
		}
	}
	switch err := ctx.Err(); err {
	case nil, context.Canceled:
	default:
		return nil, err
	}

	ret := make([]net.HardwareAddr, len(addrSet))
	ret = ret[:0]
	for _, hwAddr := range addrSet {
		ret = append(ret, hwAddr)
	}
	return ret, nil
}

func MACMustParse(s string) net.HardwareAddr {
	hwAddr, err := net.ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return hwAddr
}
