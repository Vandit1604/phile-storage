// Package p2p gives each peer a real libp2p network identity and a
// decentralized way to find and fetch content by CID.
//
// This runs alongside the existing etcd/Redis/HTTP path rather than replacing
// it: if libp2p discovery turns up nothing, the caller falls back to the older
// mechanism.
package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BlockProvider hands raw block bytes to the stream handler when another peer
// asks for a CID we hold locally.
type BlockProvider func(cidStr string) ([]byte, error)

// Node is a libp2p host plus a Kademlia DHT for content routing.
type Node struct {
	host host.Host
	dht  *dht.IpfsDHT
}

// NewNode starts a libp2p host with a stable identity, a DHT, and local
// (mDNS) peer discovery. The identity key is persisted so a peer keeps its
// PeerID across restarts.
func NewNode(ctx context.Context, peerUUID string, listenPort int, provide BlockProvider) (*Node, error) {
	priv, err := loadOrCreateIdentity(peerUUID)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	kademlia, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		_ = h.Close()
		return nil, fmt.Errorf("new dht: %w", err)
	}
	if err := kademlia.Bootstrap(ctx); err != nil {
		slog.Warn("dht bootstrap", "err", err)
	}

	n := &Node{host: h, dht: kademlia}
	n.registerFetchHandler(provide)

	if err := startMDNS(h); err != nil {
		slog.Warn("mdns discovery disabled", "err", err)
	}

	slog.Info("libp2p node started", "peerID", h.ID().String(), "addrs", n.Addrs())
	return n, nil
}

// ID returns this node's PeerID — its self-certifying cryptographic identity.
func (n *Node) ID() string { return n.host.ID().String() }

// Addrs returns the dialable multiaddrs for this node.
func (n *Node) Addrs() []string {
	out := make([]string, 0, len(n.host.Addrs()))
	id := n.host.ID().String()
	for _, a := range n.host.Addrs() {
		out = append(out, fmt.Sprintf("%s/p2p/%s", a, id))
	}
	return out
}

// AddrInfo bundles this node's PeerID and addresses for another node to dial.
func (n *Node) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{ID: n.host.ID(), Addrs: n.host.Addrs()}
}

// PeersHTTP maps connected peers (and self) to their HTTP addresses, derived
// from libp2p multiaddrs via the port convention. Used for the peer list when
// running without etcd.
func (n *Node) PeersHTTP(selfHTTP string) map[string]string {
	out := map[string]string{n.host.ID().String(): selfHTTP}
	for _, p := range n.host.Network().Peers() {
		if h := httpAddrFromMultiaddrs(n.host.Peerstore().Addrs(p)); h != "" {
			out[p.String()] = h
		}
	}
	return out
}

func (n *Node) Close() error { return n.host.Close() }

// identityPath stores the key next to the peer's blocks.
func identityPath(peerUUID string) string {
	return filepath.Join("data", peerUUID, "identity.key")
}

func loadOrCreateIdentity(peerUUID string) (crypto.PrivKey, error) {
	path := identityPath(peerUUID)
	if raw, err := os.ReadFile(path); err == nil {
		priv, err := crypto.UnmarshalPrivateKey(raw)
		if err != nil {
			return nil, fmt.Errorf("unmarshal identity: %w", err)
		}
		return priv, nil
	}

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}
	raw, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal identity: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("identity dir: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return nil, fmt.Errorf("write identity: %w", err)
	}
	return priv, nil
}
