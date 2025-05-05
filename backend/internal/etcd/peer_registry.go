package etcd

import (
	"context"
	"errors"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// PeerRegistry manages peer registration in Etcd
type PeerRegistry struct {
	client *EtcdClient
	ttl    int64 // TTL in seconds
}

// NewPeerRegistry initializes the peer registry
func NewPeerRegistry(client *EtcdClient, ttl int64) *PeerRegistry {
	return &PeerRegistry{
		client: client,
		ttl:    ttl,
	}
}

// RegisterPeerWithHeartbeat registers a peer and keeps it alive
func (pr *PeerRegistry) RegisterPeerWithHeartbeat(peerID, address string) error {
	ctx := context.Background()
	key := fmt.Sprintf("/peers/%s", peerID)

	// Create lease
	lease, err := pr.client.client.Grant(ctx, pr.ttl)
	if err != nil {
		return fmt.Errorf("failed to create lease: %v", err)
	}

	// Store peer with lease
	_, err = pr.client.client.Put(ctx, key, address, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register peer: %v", err)
	}

	log.Printf("✅ Peer %s registered at %s", peerID, address)

	// Create keep-alive channel
	keepAliveChan, err := pr.client.client.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		return fmt.Errorf("failed to start keep-alive: %v", err)
	}

	// Start a goroutine to handle lease keep-alive responses
	go func() {
		for {
			select {
			case kaResp, ok := <-keepAliveChan:
				if !ok {
					log.Printf("⚠️ KeepAlive channel closed for peer %s. Retrying registration...", peerID)

					// Re-register on failure
					lease, err = pr.client.client.Grant(ctx, pr.ttl)
					if err != nil {
						log.Printf("❌ Failed to create new lease for %s: %v", peerID, err)
						return
					}

					_, err = pr.client.client.Put(ctx, key, address, clientv3.WithLease(lease.ID))
					if err != nil {
						log.Printf("❌ Failed to re-register peer %s: %v", peerID, err)
						return
					}

					keepAliveChan, err = pr.client.client.KeepAlive(context.Background(), lease.ID)
					if err != nil {
						log.Printf("❌ Failed to restart KeepAlive for peer %s: %v", peerID, err)
						return
					}

					log.Printf("✅ Peer %s successfully re-registered after KeepAlive failure", peerID)
				} else if kaResp == nil {
					log.Printf("⚠️ Received nil KeepAlive response for peer %s. Checking lease status...", peerID)
				}
			}
		}
	}()

	return nil
}

// Returns all the peers list in etcd with prefix "/peers/"
func (pr *PeerRegistry) GetPeers(ctx context.Context) (map[string]string, error) {
	resp, err := pr.client.client.Get(ctx, "/peers/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	peers := make(map[string]string)
	if resp.Count == 0 {
		return nil, errors.New("No Peer in etcd")
	}

	for _, kv := range resp.Kvs {
		peers[string(kv.Key)] = string(kv.Value)
	}

	return peers, nil
}
