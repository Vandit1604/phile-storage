package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
)

// Redis key layout:
//
//	cid:<cid>      -> set of peer addresses that hold the block
//	name:<filename> -> the CID a human-facing name currently resolves to
//
// Names point at content rather than the other way around: many names can map
// to the same bytes, and re-uploading identical bytes is idempotent.
const (
	cidKeyPrefix  = "cid:"
	nameKeyPrefix = "name:"
	opTimeout     = 5 * time.Second
)

// ErrNameNotFound is returned when a filename has never been registered.
var ErrNameNotFound = errors.New("name not found")

type MetadataStore struct {
	client *redis.Client
}

// FileEntry is one row of the file map: a name, the content it resolves to,
// and every peer known to hold that content.
type FileEntry struct {
	CID     string   `json:"cid"`
	Holders []string `json:"holders"`
}

func NewMetadataStore(redisAddr string) *MetadataStore {
	return &MetadataStore{
		client: redis.NewClient(&redis.Options{Addr: redisAddr}),
	}
}

// AddHolder records that peerAddress holds the block for c.
func (ms *MetadataStore) AddHolder(ctx context.Context, c cid.Cid, peerAddress string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if err := ms.client.SAdd(ctx, cidKeyPrefix+c.String(), peerAddress).Err(); err != nil {
		return fmt.Errorf("add holder for %s: %w", c, err)
	}
	return nil
}

// GetHolders returns every peer known to hold the block for c.
func (ms *MetadataStore) GetHolders(ctx context.Context, c cid.Cid) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	holders, err := ms.client.SMembers(ctx, cidKeyPrefix+c.String()).Result()
	if err != nil {
		return nil, fmt.Errorf("get holders for %s: %w", c, err)
	}
	return holders, nil
}

// SetName points a filename at the content it currently resolves to.
func (ms *MetadataStore) SetName(ctx context.Context, filename string, c cid.Cid) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if err := ms.client.Set(ctx, nameKeyPrefix+filename, c.String(), 0).Err(); err != nil {
		return fmt.Errorf("set name %q: %w", filename, err)
	}
	return nil
}

// ResolveName returns the CID a filename currently resolves to.
func (ms *MetadataStore) ResolveName(ctx context.Context, filename string) (cid.Cid, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	val, err := ms.client.Get(ctx, nameKeyPrefix+filename).Result()
	if errors.Is(err, redis.Nil) {
		return cid.Undef, ErrNameNotFound
	}
	if err != nil {
		return cid.Undef, fmt.Errorf("resolve name %q: %w", filename, err)
	}
	c, err := cid.Decode(val)
	if err != nil {
		return cid.Undef, fmt.Errorf("decode cid for %q: %w", filename, err)
	}
	return c, nil
}

// ListAll returns the full file map: filename -> { cid, holders }.
func (ms *MetadataStore) ListAll(ctx context.Context) (map[string]FileEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	names, err := ms.client.Keys(ctx, nameKeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("list names: %w", err)
	}

	result := make(map[string]FileEntry, len(names))
	for _, nameKey := range names {
		cidStr, err := ms.client.Get(ctx, nameKey).Result()
		if err != nil {
			continue
		}
		holders, _ := ms.client.SMembers(ctx, cidKeyPrefix+cidStr).Result()
		filename := strings.TrimPrefix(nameKey, nameKeyPrefix)
		result[filename] = FileEntry{CID: cidStr, Holders: holders}
	}
	return result, nil
}
