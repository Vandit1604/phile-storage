package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const storagePath = "data/" // Directory to store files

// FileStore handles storing and retrieving files
type FileStore struct {
	basePath string
	peerID   string
}

// NewFileStore initializes a new FileStore
func NewFileStore(peerID string) *FileStore {
	// Ensure storage directory exists
	peerStoragePath := filepath.Join(storagePath, peerID)
	if err := os.MkdirAll(peerStoragePath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	return &FileStore{
		basePath: storagePath,
		peerID:   peerID,
	}
}

// SaveFile stores an uploaded file
func (fs *FileStore) SaveFile(peerUUID, filename string, data io.Reader) error {
	peerDir := filepath.Join(fs.basePath, peerUUID) // ✅ Use peerUUID instead of IP
	if err := os.MkdirAll(peerDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create peer directory: %v", err)
	}

	filePath := filepath.Join(peerDir, filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, data)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("✅ File %s saved at %s", filename, filePath)
	return nil
}

// GetFile returns the file path for retrieval
func (fs *FileStore) GetFile(peerUUID, filename string) (string, error) {
	filePath := filepath.Join(fs.basePath, peerUUID, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found")
	}
	return filePath, nil
}
