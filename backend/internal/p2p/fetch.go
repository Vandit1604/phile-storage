package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vandit1604/phile-storage/internal/content"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// fetchProtocol is the wire protocol for requesting a block by CID:
// the client writes "<cid>\n", the server streams the raw block back.
const fetchProtocol protocol.ID = "/phile/fetch/1.0.0"

func (n *Node) registerFetchHandler(provide BlockProvider) {
	n.host.SetStreamHandler(fetchProtocol, func(s network.Stream) {
		defer s.Close()

		cidStr, err := bufio.NewReader(s).ReadString('\n')
		if err != nil {
			slog.Debug("fetch: read request", "err", err)
			return
		}
		cidStr = strings.TrimSpace(cidStr)

		data, err := provide(cidStr)
		if err != nil {
			slog.Debug("fetch: not served", "cid", cidStr, "err", err)
			return
		}
		if _, err := s.Write(data); err != nil {
			slog.Debug("fetch: write block", "err", err)
		}
	})
}

// Provide announces to the DHT that this node holds c.
func (n *Node) Provide(ctx context.Context, c cid.Cid) error {
	if err := n.dht.Provide(ctx, c, true); err != nil {
		return fmt.Errorf("provide %s: %w", c, err)
	}
	return nil
}

// FindProviders asks the network who holds c, returning up to max providers.
func (n *Node) FindProviders(ctx context.Context, c cid.Cid, max int) []peer.AddrInfo {
	out := make([]peer.AddrInfo, 0, max)
	for pi := range n.dht.FindProvidersAsync(ctx, c, max) {
		if pi.ID == n.host.ID() {
			continue
		}
		out = append(out, pi)
	}
	return out
}

// Fetch pulls a block from a provider over a stream and returns it only if the
// bytes hash to c. Wrong or tampered content is rejected, never returned.
func (n *Node) Fetch(ctx context.Context, c cid.Cid, from peer.AddrInfo) ([]byte, error) {
	if err := n.host.Connect(ctx, from); err != nil {
		return nil, fmt.Errorf("connect %s: %w", from.ID, err)
	}

	s, err := n.host.NewStream(ctx, from.ID, fetchProtocol)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", from.ID, err)
	}
	defer s.Close()

	if _, err := fmt.Fprintf(s, "%s\n", c.String()); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if err := s.CloseWrite(); err != nil {
		return nil, fmt.Errorf("close write: %w", err)
	}

	data, err := io.ReadAll(s)
	if err != nil {
		return nil, fmt.Errorf("read block: %w", err)
	}
	if !content.Verify(data, c) {
		return nil, fmt.Errorf("integrity check failed for %s", c)
	}
	return data, nil
}
