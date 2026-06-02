package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vandit1604/phile-storage/internal/config"
	"github.com/Vandit1604/phile-storage/internal/content"
	"github.com/Vandit1604/phile-storage/internal/p2p"
	"github.com/Vandit1604/phile-storage/internal/storage"
	"github.com/ipfs/go-cid"
)

// providerLimit caps how many DHT providers we try before giving up.
const providerLimit = 5

// MetadataStore is the content index the server depends on. It is satisfied by
// the Redis-backed store (centralized mode) or the libp2p-backed store
// (decentralized mode).
type MetadataStore interface {
	AddHolder(ctx context.Context, c cid.Cid, peerAddress string) error
	GetHolders(ctx context.Context, c cid.Cid) ([]string, error)
	SetName(ctx context.Context, filename string, c cid.Cid) error
	ResolveName(ctx context.Context, filename string) (cid.Cid, error)
	ListAll(ctx context.Context) (map[string]storage.FileEntry, error)
}

// PeerSource lists the active peers — etcd in centralized mode, the libp2p
// peerstore in decentralized mode.
type PeerSource interface {
	GetPeers(ctx context.Context) (map[string]string, error)
}

type Server struct {
	fileStore     *storage.FileStore
	metadataStore MetadataStore
	peerRegistry  PeerSource
	node          *p2p.Node
	cfg           config.Config
	httpClient    *http.Client
	peerAddress   string
	peerUUID      string
}

func NewServer(
	fileStore *storage.FileStore,
	metadataStore MetadataStore,
	peerRegistry PeerSource,
	node *p2p.Node,
	cfg config.Config,
	peerUUID, peerAddress string,
) *Server {
	return &Server{
		fileStore:     fileStore,
		metadataStore: metadataStore,
		peerRegistry:  peerRegistry,
		node:          node,
		cfg:           cfg,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		peerUUID:      peerUUID,
		peerAddress:   peerAddress,
	}
}

// UploadFileHandler stores a file, indexes it by its CID, and announces it.
func (s *Server) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxUploadSize)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	c, data, err := content.ComputeReader(file)
	if err != nil {
		http.Error(w, "failed to hash file", http.StatusInternalServerError)
		return
	}

	if err := s.fileStore.SaveBlock(c, bytes.NewReader(data)); err != nil {
		http.Error(w, "failed to store file", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	if err := s.metadataStore.AddHolder(ctx, c, s.peerAddress); err != nil {
		http.Error(w, "failed to index file", http.StatusInternalServerError)
		return
	}
	if err := s.metadataStore.SetName(ctx, header.Filename, c); err != nil {
		http.Error(w, "failed to register name", http.StatusInternalServerError)
		return
	}
	s.announce(c)

	slog.Info("file uploaded", "filename", header.Filename, "cid", c.String(), "peer", s.peerUUID)
	writeJSON(w, http.StatusOK, map[string]string{"filename": header.Filename, "cid": c.String()})
}

// DownloadFileHandler serves a file by name or CID, fetching it from the
// network (libp2p first, HTTP second) and verifying its hash before caching.
func (s *Server) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	c, filename, err := s.resolveTarget(ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !s.fileStore.HasBlock(c) {
		if err := s.fetchAndCache(ctx, c); err != nil {
			slog.Warn("download failed", "cid", c.String(), "err", err)
			http.Error(w, "file not available on any peer", http.StatusNotFound)
			return
		}
	}

	path, err := s.fileStore.GetBlockPath(c)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	if filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	}
	http.ServeFile(w, r, path)
}

// BlockHandler serves a local block by CID with no discovery or recursion.
// Peers use it as the HTTP transport when fetching from each other.
func (s *Server) BlockHandler(w http.ResponseWriter, r *http.Request) {
	c, err := content.Parse(r.URL.Query().Get("cid"))
	if err != nil {
		http.Error(w, "valid cid required", http.StatusBadRequest)
		return
	}
	path, err := s.fileStore.GetBlockPath(c)
	if err != nil {
		http.Error(w, "block not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path)
}

// DiscoverFileHandler reports the CID a name resolves to and who holds it.
func (s *Server) DiscoverFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	c, err := s.metadataStore.ResolveName(ctx, filename)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cid": c.String(), "holders": s.holders(ctx, c)})
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func (s *Server) GetPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers, err := s.peerRegistry.GetPeers(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch peers", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, peers)
}

func (s *Server) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := s.metadataStore.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch files", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, files)
}

// P2PInfoHandler exposes this node's libp2p identity.
func (s *Server) P2PInfoHandler(w http.ResponseWriter, r *http.Request) {
	if s.node == nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": true,
		"peerID":  s.node.ID(),
		"addrs":   s.node.Addrs(),
	})
}

func (s *Server) Start(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", s.UploadFileHandler)
	mux.HandleFunc("/download", s.DownloadFileHandler)
	mux.HandleFunc("/block", s.BlockHandler)
	mux.HandleFunc("/discover", s.DiscoverFileHandler)
	mux.HandleFunc("/search", s.SearchHandler)
	mux.HandleFunc("/peers", s.GetPeersHandler)
	mux.HandleFunc("/files", s.ListFilesHandler)
	mux.HandleFunc("/p2p/info", s.P2PInfoHandler)

	slog.Info("api server listening", "port", port, "peer", s.peerUUID)
	if err := http.ListenAndServe(port, s.withCORS(mux)); err != nil {
		slog.Error("api server stopped", "err", err)
	}
}
