package storage

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
)

const storagePath = "data/"

// FileStore persists blocks on the local filesystem, one directory per peer.
//
// Blocks are named by their CID, not by a user-supplied filename. Because a
// CID is a fixed, self-describing string, it can never contain path separators
// or "..", so an attacker cannot steer a write outside the peer's directory.
type FileStore struct {
	basePath string
	peerID   string
}

func NewFileStore(peerID string) *FileStore {
	blocksDir := filepath.Join(storagePath, peerID, "blocks")
	if err := os.MkdirAll(blocksDir, 0o755); err != nil {
		slog.Error("create storage directory", "err", err)
		os.Exit(1)
	}
	return &FileStore{basePath: storagePath, peerID: peerID}
}

func (fs *FileStore) blockPath(c cid.Cid) string {
	return filepath.Join(fs.basePath, fs.peerID, "blocks", c.String())
}

// SaveBlock writes a block addressed by its CID. Writing the same CID twice is
// a harmless no-op since the bytes are identical by definition.
func (fs *FileStore) SaveBlock(c cid.Cid, data io.Reader) error {
	if !c.Defined() {
		return fmt.Errorf("save block: undefined cid")
	}
	path := fs.blockPath(c)
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create block %s: %w", c, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, data); err != nil {
		return fmt.Errorf("write block %s: %w", c, err)
	}
	slog.Info("block saved", "cid", c.String(), "peer", fs.peerID)
	return nil
}

// HasBlock reports whether the block is present locally.
func (fs *FileStore) HasBlock(c cid.Cid) bool {
	if !c.Defined() {
		return false
	}
	_, err := os.Stat(fs.blockPath(c))
	return err == nil
}

// GetBlockPath returns the on-disk path of a local block.
func (fs *FileStore) GetBlockPath(c cid.Cid) (string, error) {
	if !fs.HasBlock(c) {
		return "", fmt.Errorf("block %s not found", c)
	}
	return fs.blockPath(c), nil
}

// ReadBlock returns the raw bytes of a local block.
func (fs *FileStore) ReadBlock(c cid.Cid) ([]byte, error) {
	path, err := fs.GetBlockPath(c)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// ListBlockCIDs returns the CIDs of every block already on disk. Used at
// startup to re-announce persisted content to the network.
func (fs *FileStore) ListBlockCIDs() ([]cid.Cid, error) {
	dir := filepath.Join(fs.basePath, fs.peerID, "blocks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list blocks: %w", err)
	}

	var cids []cid.Cid
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if c, err := cid.Decode(e.Name()); err == nil {
			cids = append(cids, c)
		}
	}
	return cids, nil
}
