package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"net"
	"strings"
	"time"

	hostapd "go.jonnrb.io/hostapd_grpc/proto"
	"go.jonnrb.io/wifitriggers"
	"google.golang.org/grpc"
)

type Flags struct {
	IFTTTKey string
	Backend  string
	MACs     []net.HardwareAddr
}

func parseFlags() Flags {
	var f Flags
	flag.StringVar(&f.IFTTTKey, "iftttKey", "", "IFTTT API (\"Maker\") key")
	flag.StringVar(&f.Backend, "backend", "hostapd_grpc:8080", "hostapd_grpc backend")
	macStrs := flag.String("trackedMACs", "", "MACs to consider as somebody \"home\"")

	flag.Parse()

	for _, macStr := range strings.Split(*macStrs, ",") {
		f.MACs = append(f.MACs, wifitriggers.MACMustParse(macStr))
	}

	return f
}

func isSomebodyHome(f Flags, set []net.HardwareAddr) bool {
	// If the sets intersect, somebody's home.
	for _, a := range set {
		for _, b := range f.MACs {
			if bytes.Equal(a, b) {
				return true
			}
		}
	}
	return false
}

func main() {
	f := parseFlags()

	s := wifitriggers.SwitchOnIFTTT{
		Key:    f.IFTTTKey,
		OnCmd:  "arm_wyzecam",
		OffCmd: "disarm_wyzecam",
	}

	var (
		SomebodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return isSomebodyHome(f, c) })
		NobodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return !isSomebodyHome(f, c) })
	)

	cc, err := grpc.Dial(f.Backend, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	d := wifitriggers.Driver{
		APReader: wifitriggers.HostapdReader{hostapd.NewHostapdControlClient(cc)},
		BindingChain: wifitriggers.NilBindingChain.
			AddBinding(wifitriggers.If(SomebodyIsHome).Then(s.OffAction())).
			AddBinding(wifitriggers.If(NobodyIsHome).Then(s.OnAction())),
		Interval: 2 * time.Second,
	}
	if err := d.Run(context.Background()); err != nil {
		log.Println("Error running:", err)
	}
}
