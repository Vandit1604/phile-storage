package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Vandit1604/phile-storage/internal/storage"
	"github.com/ipfs/go-cid"
	ma "github.com/multiformats/go-multiaddr"
)

// PortOffset is the fixed gap between a peer's libp2p port and its HTTP port
// (libp2p = HTTP + PortOffset). main.go assigns ports with this convention, so
// we can recover a peer's HTTP address from its libp2p multiaddr on this host.
const PortOffset = 1000

// Store is the decentralized content index: holder lookup rides the DHT, and
// human filenames are resolved from a node-local map (a node only knows names
// for files it has uploaded). The map is persisted to namesPath so it survives
// restarts. No etcd or Redis involved.
type Store struct {
	node      *Node
	selfHTTP  string
	namesPath string

	mu    sync.RWMutex
	names map[string]cid.Cid
}

func NewStore(node *Node, selfHTTP, namesPath string) *Store {
	s := &Store{
		node:      node,
		selfHTTP:  selfHTTP,
		namesPath: namesPath,
		names:     make(map[string]cid.Cid),
	}
	s.load()
	return s
}

// load restores the name map from disk; a missing or unreadable file just
// leaves the map empty.
func (s *Store) load() {
	data, err := os.ReadFile(s.namesPath)
	if err != nil {
		return
	}
	var raw map[string]string
	if json.Unmarshal(data, &raw) != nil {
		return
	}
	for name, cidStr := range raw {
		if c, err := cid.Decode(cidStr); err == nil {
			s.names[name] = c
		}
	}
	slog.Info("restored name index", "entries", len(s.names))
}

// persist writes a snapshot of the name map; best-effort, never fatal.
func (s *Store) persist(snapshot map[string]string) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}
	if err := os.WriteFile(s.namesPath, data, 0o644); err != nil {
		slog.Warn("persist name index", "err", err)
	}
}

// AddHolder announces to the DHT that this node holds c. The address argument
// is ignored — libp2p identifies the holder by PeerID.
func (s *Store) AddHolder(ctx context.Context, c cid.Cid, _ string) error {
	return s.node.Provide(ctx, c)
}

// GetHolders resolves DHT providers for c into HTTP addresses.
func (s *Store) GetHolders(ctx context.Context, c cid.Cid) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var addrs []string
	for _, pi := range s.node.FindProviders(ctx, c, 10) {
		if h := httpAddrFromMultiaddrs(pi.Addrs); h != "" {
			addrs = append(addrs, h)
		}
	}
	return addrs, nil
}

func (s *Store) SetName(_ context.Context, filename string, c cid.Cid) error {
	s.mu.Lock()
	s.names[filename] = c
	snapshot := make(map[string]string, len(s.names))
	for name, id := range s.names {
		snapshot[name] = id.String()
	}
	s.mu.Unlock()

	s.persist(snapshot)
	return nil
}

func (s *Store) ResolveName(_ context.Context, filename string) (cid.Cid, error) {
	s.mu.RLock()
	c, ok := s.names[filename]
	s.mu.RUnlock()
	if !ok {
		return cid.Undef, fmt.Errorf("name %q not known to this node", filename)
	}
	return c, nil
}

// ListAll returns this node's local file view. Holders is just this node — a
// full network view would need a shared index (gossip), which this mode omits.
func (s *Store) ListAll(_ context.Context) (map[string]storage.FileEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]storage.FileEntry, len(s.names))
	for name, c := range s.names {
		out[name] = storage.FileEntry{CID: c.String(), Holders: []string{s.selfHTTP}}
	}
	return out, nil
}

func (s *Store) GetPeers(_ context.Context) (map[string]string, error) {
	return s.node.PeersHTTP(s.selfHTTP), nil
}

// httpAddrFromMultiaddrs recovers a peer's HTTP address from its libp2p
// multiaddrs using the PortOffset convention. This targets the single-host
// demo: every peer's HTTP API is reachable on loopback, so we keep the port
// and always emit 127.0.0.1 (a peer's advertised LAN/bridge IPs may not be
// reachable from the browser). Returns "" if no TCP port is found.
func httpAddrFromMultiaddrs(addrs []ma.Multiaddr) string {
	for _, a := range addrs {
		portStr, err := a.ValueForProtocol(ma.P_TCP)
		if err != nil {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		return fmt.Sprintf("127.0.0.1:%d", port-PortOffset)
	}
	return ""
}
