package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Vandit1604/phile-storage/internal/etcd"
	"github.com/Vandit1604/phile-storage/internal/storage"
)

// Server wraps the API server
type Server struct {
	fileStore     *storage.FileStore
	metadataStore *storage.MetadataStore
	peerRegistry  *etcd.PeerRegistry
	peerAddress   string // Address of this peer
}

// NewServer initializes the API server
func NewServer(fileStore *storage.FileStore, metadataStore *storage.MetadataStore, peerRegistry *etcd.PeerRegistry, peerAddress string) *Server {
	return &Server{
		fileStore:     fileStore,
		metadataStore: metadataStore,
		peerRegistry:  peerRegistry,
		peerAddress:   peerAddress,
	}
}

func (s *Server) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the filename from query parameters
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Retrieve the file path
	filePath, err := s.fileStore.GetFile(filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve the file for download
	http.ServeFile(w, r, filePath)
}

// UploadFileHandler handles file uploads
func (s *Server) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file
	err = s.fileStore.SaveFile(header.Filename, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Register file in Redis
	ctx := context.Background()
	err = s.metadataStore.AddFile(ctx, header.Filename, s.peerAddress)
	if err != nil {
		http.Error(w, "Failed to update metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ File %s uploaded & registered on %s", header.Filename, s.peerAddress)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File %s uploaded & registered successfully", header.Filename)
}

// DiscoverFileHandler finds which peers have a given file
func (s *Server) DiscoverFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	peers, err := s.metadataStore.GetFilePeers(ctx, filename)
	if err != nil || len(peers) == 0 {
		http.Error(w, "File not found on any peer", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// GetPeersHandler returns the list of active peers
func (s *Server) GetPeersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
	}

	ctx := context.Background()

	peers, err := s.peerRegistry.GetPeers(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch peers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// Start initializes all API routes and starts the HTTP server
func (s *Server) Start(port string) {
	mux := http.NewServeMux() // Use a new ServeMux for each peer
	mux.HandleFunc("/upload", s.UploadFileHandler)
	mux.HandleFunc("/download", s.DownloadFileHandler)
	mux.HandleFunc("/discover", s.DiscoverFileHandler)
	mux.HandleFunc("/peers", s.GetPeersHandler)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("🚀 API server running on port %s", port)
	log.Fatal(server.ListenAndServe())
}
