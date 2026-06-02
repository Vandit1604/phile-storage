package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Vandit1604/phile-storage/internal/api"
	"github.com/Vandit1604/phile-storage/internal/config"
	"github.com/Vandit1604/phile-storage/internal/content"
	"github.com/Vandit1604/phile-storage/internal/etcd"
	"github.com/Vandit1604/phile-storage/internal/p2p"
	"github.com/Vandit1604/phile-storage/internal/storage"
)

// backend bundles the shared services a peer needs. In decentralized mode the
// etcd/redis fields are nil and each peer builds its own libp2p-backed store.
type backend struct {
	cfg      config.Config
	registry *etcd.PeerRegistry     // nil in decentralized mode
	metadata *storage.MetadataStore // nil in decentralized mode
}

func startPeer(ctx context.Context, b backend, index int) {
	cfg := b.cfg
	httpPort := cfg.BasePort + index
	libp2pPort := cfg.BasePort + p2p.PortOffset + index
	// A stable, port-derived ID keeps each peer's data directory (blocks +
	// identity key + name index) across restarts.
	peerID := fmt.Sprintf("peer-%d", httpPort)
	peerAddress := fmt.Sprintf("127.0.0.1:%d", httpPort)

	fileStore := storage.NewFileStore(peerID)

	// The libp2p host serves blocks it holds locally to peers asking by CID.
	provide := func(cidStr string) ([]byte, error) {
		c, err := content.Parse(cidStr)
		if err != nil {
			return nil, err
		}
		return fileStore.ReadBlock(c)
	}

	node, err := p2p.NewNode(ctx, peerID, libp2pPort, provide)
	if err != nil {
		slog.Error("libp2p is required but failed to start", "peer", peerID, "err", err)
		return
	}

	var meta api.MetadataStore
	var peers api.PeerSource
	if cfg.UseEtcdRedis {
		if err := b.registry.RegisterPeerWithHeartbeat(peerID, peerAddress); err != nil {
			slog.Error("register peer", "peer", peerID, "err", err)
			return
		}
		meta, peers = b.metadata, b.registry
	} else {
		namesPath := filepath.Join("data", peerID, "names.json")
		store := p2p.NewStore(node, peerAddress, namesPath)
		meta, peers = store, store
	}

	// Re-announce blocks already on disk so persisted content is discoverable
	// again after a restart.
	go reannounce(ctx, node, fileStore)

	server := api.NewServer(fileStore, meta, peers, node, cfg, peerID, peerAddress)
	go server.Start(fmt.Sprintf(":%d", httpPort))

	slog.Info("peer running", "http", peerAddress, "id", peerID, "mode", mode(cfg))
}

// reannounce tells the DHT about every block this peer already holds on disk.
func reannounce(ctx context.Context, node *p2p.Node, fileStore *storage.FileStore) {
	cids, err := fileStore.ListBlockCIDs()
	if err != nil || len(cids) == 0 {
		return
	}
	for _, c := range cids {
		annCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		_ = node.Provide(annCtx, c)
		cancel()
	}
	slog.Info("re-announced persisted blocks", "count", len(cids))
}

func mode(cfg config.Config) string {
	if cfg.UseEtcdRedis {
		return "centralized (etcd+redis)"
	}
	return "decentralized (libp2p)"
}

func main() {
	numPeers := flag.Int("peers", 1, "number of peer nodes to start")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()
	b := backend{cfg: cfg}

	// Only stand up etcd + Redis when explicitly enabled. The default path
	// needs no external infrastructure.
	if cfg.UseEtcdRedis {
		client, err := etcd.NewEtcdClient(cfg.EtcdEndpoints)
		if err != nil {
			slog.Error("connect etcd", "err", err)
			os.Exit(1)
		}
		defer client.Close()
		b.registry = etcd.NewPeerRegistry(client, 10)
		b.metadata = storage.NewMetadataStore(cfg.RedisAddr)
	}
	slog.Info("starting phile-storage", "mode", mode(cfg), "peers", *numPeers)

	for i := 0; i < *numPeers; i++ {
		go startPeer(ctx, b, i)
		time.Sleep(500 * time.Millisecond)
	}

	select {}
}
