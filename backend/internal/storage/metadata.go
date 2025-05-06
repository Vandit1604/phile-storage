package storage

import (
	"context"
	"fmt"
	"strings"

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
func (ms *MetadataStore) AddFile(ctx context.Context, filename, peerAddress string) error {
	key := fmt.Sprintf("file:%s", filename)
	return ms.client.SAdd(ctx, key, peerAddress).Err()
}

// GetFilePeers returns the peers that have reference filename
func (ms *MetadataStore) GetFilePeers(ctx context.Context, filename string) ([]string, error) {
	peers, err := ms.client.SMembers(ctx, filename).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}

	return peers, nil
}

func (m *MetadataStore) ListAllFiles(ctx context.Context) (map[string][]string, error) {
	keys, err := m.client.Keys(ctx, "file:*").Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, key := range keys {
		peers, err := m.client.SMembers(ctx, key).Result()
		if err != nil {
			continue
		}
		filename := strings.TrimPrefix(key, "file:")
		result[filename] = peers
	}
	return result, nil
}
