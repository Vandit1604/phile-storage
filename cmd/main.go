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

func startPeer(peerID, peerAddress string) {
	// Etcd connection (do not defer client.Close())
	etcdEndpoints := []string{"localhost:2379"}
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Etcd: %v", err)
	}

	// Create Peer Registry
	peerRegistry := etcd.NewPeerRegistry(client, 10)

	// Register peer in Etcd
	err = peerRegistry.RegisterPeerWithHeartbeat(peerID, peerAddress)
	if err != nil {
		log.Fatalf("❌ Failed to register peer: %v", err)
	}

	// Initialize file storage and metadata store
	fileStore := storage.NewFileStore()
	metadataStore := storage.NewMetadataStore("localhost:6379")

	// Initialize API server
	server := api.NewServer(fileStore, metadataStore, peerRegistry, peerAddress)

	// Start server in a separate routine
	go server.Start(peerAddress[9:]) // Extracts port from "127.0.0.1:500X"

	log.Printf("🚀 Peer %s running on %s", peerID, peerAddress)
}

func main() {
	// Command-line flag to specify number of peers
	numPeers := flag.Int("peers", 1, "Number of peer nodes to start")
	flag.Parse()

	// Start multiple peers as goroutines
	for i := 0; i < *numPeers; i++ {
		port := 5001 + i
		peerID := uuid.New().String()
		peerAddress := fmt.Sprintf("127.0.0.1:%d", port)

		go startPeer(peerID, peerAddress) // Start each peer in a goroutine

		// Add a short delay to avoid race conditions
		time.Sleep(500 * time.Millisecond)
	}

	// Keep the main process running
	select {}
}
