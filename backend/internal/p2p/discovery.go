package p2p

import (
	"context"
	"log/slog"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// mdnsServiceTag scopes local discovery to this application.
const mdnsServiceTag = "phile-storage"

// mdnsNotifee dials any peer found on the local network so the DHT has someone
// to talk to without a public bootstrap list — ideal for the localhost demo.
type mdnsNotifee struct {
	h host.Host
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == m.h.ID() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.h.Connect(ctx, pi); err != nil {
		slog.Debug("mdns connect failed", "peer", pi.ID.String(), "err", err)
		return
	}
	slog.Info("peer discovered", "peer", pi.ID.String())
}

func startMDNS(h host.Host) error {
	return mdns.NewMdnsService(h, mdnsServiceTag, &mdnsNotifee{h: h}).Start()
}
