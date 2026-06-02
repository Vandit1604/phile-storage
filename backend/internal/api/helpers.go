package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vandit1604/phile-storage/internal/content"
	"github.com/ipfs/go-cid"
)

// resolveTarget turns a request into the CID to serve. A request may name the
// content directly (?cid=) or by a human filename (?filename=) that resolves
// through the name index.
func (s *Server) resolveTarget(ctx context.Context, r *http.Request) (cid.Cid, string, error) {
	if cidStr := r.URL.Query().Get("cid"); cidStr != "" {
		c, err := content.Parse(cidStr)
		if err != nil {
			return cid.Undef, "", fmt.Errorf("invalid cid")
		}
		return c, r.URL.Query().Get("filename"), nil
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		return cid.Undef, "", fmt.Errorf("filename or cid is required")
	}
	c, err := s.metadataStore.ResolveName(ctx, filename)
	if err != nil {
		return cid.Undef, "", fmt.Errorf("unknown file")
	}
	return c, filename, nil
}

// fetchAndCache pulls a missing block from the network: the libp2p DHT first,
// then the Redis holder list over HTTP. Either path verifies the hash before
// the bytes are trusted.
func (s *Server) fetchAndCache(ctx context.Context, c cid.Cid) error {
	if s.node != nil {
		fctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		for _, pi := range s.node.FindProviders(fctx, c, providerLimit) {
			data, err := s.node.Fetch(fctx, c, pi)
			if err != nil {
				slog.Debug("libp2p fetch failed", "peer", pi.ID.String(), "err", err)
				continue
			}
			slog.Info("fetched via libp2p", "cid", c.String(), "from", pi.ID.String())
			return s.store(ctx, c, data)
		}
	}

	holders, err := s.metadataStore.GetHolders(ctx, c)
	if err != nil {
		return err
	}
	for _, addr := range holders {
		if addr == s.peerAddress {
			continue
		}
		data, err := s.httpFetch(ctx, c, addr)
		if err != nil {
			slog.Debug("http fetch failed", "addr", addr, "err", err)
			continue
		}
		slog.Info("fetched via http", "cid", c.String(), "from", addr)
		return s.store(ctx, c, data)
	}
	return fmt.Errorf("no provider served %s", c)
}

func (s *Server) httpFetch(ctx context.Context, c cid.Cid, addr string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/block?cid=%s", addr, c.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, s.cfg.MaxUploadSize))
	if err != nil {
		return nil, err
	}
	if !content.Verify(data, c) {
		return nil, fmt.Errorf("integrity check failed for %s", c)
	}
	return data, nil
}

func (s *Server) store(ctx context.Context, c cid.Cid, data []byte) error {
	if err := s.fileStore.SaveBlock(c, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := s.metadataStore.AddHolder(ctx, c, s.peerAddress); err != nil {
		slog.Warn("index fetched block", "err", err)
	}
	s.announce(c)
	return nil
}

// announce tells the DHT we now hold c. It runs in the background because
// providing can be slow and must not block the response.
func (s *Server) announce(c cid.Cid) {
	if s.node == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := s.node.Provide(ctx, c); err != nil {
			slog.Debug("provide failed", "cid", c.String(), "err", err)
		}
	}()
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(s.cfg.CORSOrigins))
	for _, o := range s.cfg.CORSOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
