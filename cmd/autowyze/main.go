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

var (
	key     = flag.String("iftttKey", "", "IFTTT API (\"Maker\") key")
	backend = flag.String("backend", "hostapd_grpc:8080", "hostapd_grpc backend")
	macStrs = flag.String("trackedMACs", "", "MACs to consider as somebody \"home\"")
)

var macs []net.HardwareAddr

func isSomebodyHome(set []net.HardwareAddr) bool {
	// If the sets intersect, somebody's home.
	for _, a := range set {
		for _, b := range macs {
			if bytes.Equal(a, b) {
				return true
			}
		}
	}
	return false
}

func main() {
	flag.Parse()

	for _, macStr := range strings.Split(*macStrs, ",") {
		macs = append(macs, wifitriggers.MACMustParse(macStr))
	}

	s := wifitriggers.SwitchOnIFTTT{
		Key:    *key,
		OnCmd:  "arm_wyzecam",
		OffCmd: "disarm_wyzecam",
	}

	var (
		SomebodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return isSomebodyHome(c) })
		NobodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return !isSomebodyHome(c) })
	)

	cc, err := grpc.Dial(*backend, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	d := wifitriggers.Driver{
		APReader: wifitriggers.HostapdReader{hostapd.NewHostapdControlClient(cc)},
		Engine: new(wifitriggers.EngineBuilder).
			Bind(wifitriggers.If(SomebodyIsHome).Then(s.OffAction())).
			Bind(wifitriggers.If(NobodyIsHome).Then(s.OnAction())).
			Build(),
		Interval: 2 * time.Second,
	}
	if err := d.Run(context.Background()); err != nil {
		log.Println("Error running:", err)
	}
}
