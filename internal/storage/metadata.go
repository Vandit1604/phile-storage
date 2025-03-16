package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// MetadataStore handles the file metadata storage in redis
type MetadataStore struct {
	client *redis.Client
}

// NewMetadataStore initializes a Redis Client
func NewMetadataStore(redisAddr string) *MetadataStore {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &MetadataStore{
		client: client,
	}
}

// AddFile adds a file and associates it with a peer
func (ms *MetadataStore) AddFile(ctx context.Context, filename string, peerAddress string) error {
	_, err := ms.client.SAdd(ctx, filename, peerAddress).Result()
	if err != nil {
		return fmt.Errorf("failed to add file to metadata store: %v", err)
	}
	log.Printf("File %s added to metadata by %s", filename, peerAddress)
	return nil
}

// GetFilePeers returns the peers that have reference filename
func (ms *MetadataStore) GetFilePeers(ctx context.Context, filename string) ([]string, error) {
	peers, err := ms.client.SMembers(ctx, filename).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}

	return peers, nil
}
