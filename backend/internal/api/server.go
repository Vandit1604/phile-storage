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
	peerAddress   string
	peerUUID      string
}

// NewServer initializes the API server
func NewServer(fileStore *storage.FileStore, metadataStore *storage.MetadataStore, peerRegistry *etcd.PeerRegistry, peerUUID, peerAddress string) *Server {
	return &Server{
		fileStore:     fileStore,
		metadataStore: metadataStore,
		peerRegistry:  peerRegistry,
		peerUUID:      peerUUID,
		peerAddress:   peerAddress,
	}
}

func (s *Server) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Step 1: Check if the file exists locally
	filePath, err := s.fileStore.GetFile(s.peerUUID, filename)
	if err == nil {
		log.Printf("✅ File %s found locally on %s", filename, s.peerUUID)
		http.ServeFile(w, r, filePath)
		return
	}

	log.Printf("⚠️ File %s not found locally on %s. Searching on other peers...", filename, s.peerUUID)

	// Step 2: Query other peers for the file
	ctx := context.Background()
	peers, err := s.metadataStore.GetFilePeers(ctx, filename)
	if err != nil || len(peers) == 0 {
		log.Printf("❌ File %s not found on any peer", filename)
		http.Error(w, "File not found on any peer", http.StatusNotFound)
		return
	}

	// Step 3: Avoid infinite loops by keeping track of already-checked peers
	checkedPeers := make(map[string]bool)
	checkedPeers[s.peerUUID] = true // Mark this peer as checked

	for _, peer := range peers {
		if peer == s.peerUUID || checkedPeers[peer] {
			continue // Skip self and already-checked peers
		}

		checkedPeers[peer] = true // Mark peer as checked
		fileURL := fmt.Sprintf("http://%s/download?filename=%s", peer, filename)
		log.Printf("📡 Requesting file %s from peer %s", filename, peer)

		resp, err := http.Get(fileURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("❌ Failed to fetch %s from peer %s. Trying next peer...", filename, peer)
			continue
		}

		// Step 4: Save the file locally
		defer resp.Body.Close()
		err = s.fileStore.SaveFile(s.peerUUID, filename, resp.Body)
		if err != nil {
			log.Printf("❌ Failed to save %s after fetching from peer %s", filename, peer)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Step 5: Update Redis metadata
		err = s.metadataStore.AddFile(ctx, filename, s.peerAddress)
		if err != nil {
			log.Printf("⚠️ Failed to update Redis metadata for %s", filename)
		}

		// Log success and serve the file
		log.Printf("✅ Successfully fetched %s from peer %s and stored locally on %s", filename, peer, s.peerUUID)
		filePath, _ = s.fileStore.GetFile(s.peerUUID, filename)
		http.ServeFile(w, r, filePath)
		return
	}

	// If no peer had the file, return an error
	log.Printf("❌ File %s not available on any peer", filename)
	http.Error(w, "File not found on any peer", http.StatusNotFound)
}

// UploadFileHandler (Now saves file under data/peerUUID/)
func (s *Server) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// ✅ Save using peerUUID instead of peerAddress
	err = s.fileStore.SaveFile(s.peerUUID, header.Filename, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Register in Redis
	ctx := context.Background()
	err = s.metadataStore.AddFile(ctx, header.Filename, s.peerAddress)
	if err != nil {
		http.Error(w, "Failed to update metadata", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("✅ File %s uploaded & registered under peer UUID: %s", header.Filename, s.peerUUID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
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
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", s.UploadFileHandler)
	mux.HandleFunc("/download", s.DownloadFileHandler)
	mux.HandleFunc("/discover", s.DiscoverFileHandler)
	mux.HandleFunc("/peers", s.GetPeersHandler)
	mux.HandleFunc("/files", s.ListFilesHandler)

	// Wrap with CORS middleware
	handler := withCORS(mux)

	log.Printf("🚀 API server running on port %s", port)
	log.Fatal(http.ListenAndServe(port, handler))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // 🔥 allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	files, err := s.metadataStore.ListAllFiles(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
