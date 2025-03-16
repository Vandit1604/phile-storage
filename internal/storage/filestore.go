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
}

// NewFileStore initializes a new FileStore
func NewFileStore() *FileStore {
	// Ensure storage directory exists
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	return &FileStore{
		basePath: storagePath,
	}
}

// SaveFile stores an uploaded file
func (fs *FileStore) SaveFile(filename string, data io.Reader) error {
	filePath := filepath.Join(fs.basePath, filename)

	// Create a new file
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the file data
	_, err = io.Copy(outFile, data)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("File %s saved successfully", filename)
	return nil
}

// GetFile returns the file path for retrieval
func (fs *FileStore) GetFile(filename string) (string, error) {
	filePath := filepath.Join(fs.basePath, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found")
	}
	return filePath, nil
}
