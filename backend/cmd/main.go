package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Vandit1604/phile-storage/internal/api"
	"github.com/Vandit1604/phile-storage/internal/etcd"
	"github.com/Vandit1604/phile-storage/internal/storage"
	"github.com/google/uuid"
)

func startPeer(peerUUID, peerAddress string, peerRegistry *etcd.PeerRegistry, metadataStore *storage.MetadataStore) {
	// Register peer using UUID
	err := peerRegistry.RegisterPeerWithHeartbeat(peerUUID, peerAddress)
	if err != nil {
		log.Fatalf("❌ Failed to register peer: %v", err)
	}

	// Initialize file storage (Unique per peer)
	fileStore := storage.NewFileStore(peerUUID)

	// Initialize API server
	server := api.NewServer(fileStore, metadataStore, peerRegistry, peerUUID, peerAddress)

	// Start server in a separate routine
	go server.Start(peerAddress[9:]) // Extracts port from "127.0.0.1:500X"

	log.Printf("🚀 Peer %s running at %s (UUID: %s)", peerAddress, peerAddress, peerUUID)
}

func main() {
	// Command-line flag to specify number of peers
	numPeers := flag.Int("peers", 1, "Number of peer nodes to start")
	flag.Parse()

	// ✅ Initialize Etcd and Redis **once** and share across all peers
	etcdEndpoints := []string{"localhost:2379"}
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Etcd: %v", err)
	}
	defer client.Close()

	peerRegistry := etcd.NewPeerRegistry(client, 10)
	metadataStore := storage.NewMetadataStore("localhost:6379")

	// ✅ Start each peer and pass shared `peerRegistry` and `metadataStore`
	for i := 0; i < *numPeers; i++ {
		port := 5001 + i
		peerUUID := uuid.New().String()
		peerAddress := fmt.Sprintf("127.0.0.1:%d", port)

		go startPeer(peerUUID, peerAddress, peerRegistry, metadataStore) // Pass shared instances

		time.Sleep(500 * time.Millisecond)
	}

	select {} // Keep the main process running
}
