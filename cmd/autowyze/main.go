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
	IFTTTKey        string
	Backend         string
	MACs            []net.HardwareAddr
	CameraMACs      []net.HardwareAddr
	SlackWebhookURL string
}

func parseMACs(csv string) (macs []net.HardwareAddr) {
	for _, macStr := range strings.Split(csv, ",") {
		macs = append(macs, wifitriggers.MACMustParse(macStr))
	}
	return
}

func parseFlags() Flags {
	var f Flags
	flag.StringVar(&f.IFTTTKey, "iftttKey", "", "IFTTT API (\"Maker\") key")
	flag.StringVar(&f.Backend, "backend", "hostapd_grpc:8080", "hostapd_grpc backend")
	macStrs := flag.String("trackedMACs", "", "MACs to consider as somebody \"home\"")
	cameraMACStrs := flag.String("cameraMACs", "", "(Optional) notification will be sent when any of these MACs goes offline")
	flag.StringVar(&f.SlackWebhookURL, "slackWebhookURL", "", "Required if \"-cameraMACs\" is set; the Slack webhook URL to send camera disconnected notifications to")

	flag.Parse()

	f.MACs = parseMACs(*macStrs)
	f.CameraMACs = parseMACs(*cameraMACStrs)

	return f
}

// If the sets intersect, somebody's home.
func hasIntersection(s, u []net.HardwareAddr) bool {
	for _, a := range s {
		for _, b := range u {
			if bytes.Equal(a, b) {
				return true
			}
		}
	}
	return false
}

// If s is a subset of u, all devices are present.
func isSubset(s, u []net.HardwareAddr) bool {
	lookup := make(map[string]struct{})
	for _, a := range s {
		lookup[a.String()] = struct{}{}
	}

	for _, a := range u {
		k := a.String()
		if _, ok := lookup[k]; ok {
			delete(lookup, k)
			if len(lookup) == 0 {
				return true
			}
		}
	}
	return len(lookup) == 0
}

func setUpPresenceDetection(f Flags) wifitriggers.BindingChain {
	var (
		SomebodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return hasIntersection(f.MACs, c) })
		NobodyIsHome = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return !hasIntersection(f.MACs, c) })
	)

	s := wifitriggers.SwitchOnIFTTT{
		Key:    f.IFTTTKey,
		OnCmd:  "arm_wyzecam",
		OffCmd: "disarm_wyzecam",
	}

	return wifitriggers.NilBindingChain.
		AddBinding(wifitriggers.If(SomebodyIsHome).Then(s.OffAction())).
		AddBinding(wifitriggers.If(NobodyIsHome).Then(s.OnAction()))
}

func maybeSetUpCameraOfflineDetection(f Flags) wifitriggers.BindingChain {
	if len(f.CameraMACs) == 0 {
		return wifitriggers.NilBindingChain
	}

	var (
		CamerasAreConnected = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return isSubset(f.CameraMACs, c) })
		CamerasAreDisconnected = wifitriggers.Cond(
			func(c []net.HardwareAddr) bool { return !isSubset(f.CameraMACs, c) })
	)

	c := wifitriggers.SwitchOnSlack{
		WebhookURL: f.SlackWebhookURL,
		OnMsg:      "All cameras are connected.",
		OffMsg:     "Some cameras are disconnected.",
	}

	return wifitriggers.NilBindingChain.
		AddBinding(wifitriggers.If(CamerasAreConnected).Then(c.OnAction())).
		AddBinding(wifitriggers.If(CamerasAreDisconnected).Then(c.OffAction()))
}

func main() {
	f := parseFlags()

	cc, err := grpc.Dial(f.Backend, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	r := wifitriggers.HostapdReader{hostapd.NewHostapdControlClient(cc)}
	b := setUpPresenceDetection(f).And(maybeSetUpCameraOfflineDetection(f))

	d := wifitriggers.Driver{
		APReader:     r,
		BindingChain: b,
		Interval:     2 * time.Second,
	}

	if err := d.Run(context.Background()); err != nil {
		log.Println("Error running:", err)
	}
}
