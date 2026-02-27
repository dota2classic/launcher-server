package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"awesomeProject/internal/manifest"
)

type Server struct {
	basePath        string
	currentManifest *manifest.Manifest
	mu              sync.RWMutex
}

func New(basePath string) (*Server, error) {
	s := &Server{
		basePath: basePath,
	}

	if err := s.recalculateManifest(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) recalculateManifest() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := manifest.CreateManifest(s.basePath)
	if err != nil {
		return err
	}

	s.currentManifest = m
	log.Printf("Recalculated manifest: %d files", len(m.Files))
	return nil
}

// GET /manifest - returns current manifest
func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.currentManifest)
}

// POST /manifest/recalculate - triggers manifest recalculation
func (s *Server) handleRecalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.recalculateManifest(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"files":  len(s.currentManifest.Files),
	})
}

// GET /files/{path} - serves files
func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Strip /files/ prefix
	relativePath := r.URL.Path[len("/files/"):]
	if relativePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	cleanPath := filepath.Clean(relativePath)
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(s.basePath, cleanPath)

	// Verify file is within basePath
	if !isSubPath(s.basePath, fullPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func isSubPath(basePath, targetPath string) bool {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && (len(rel) < 2 || rel[:2] != "..")
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/manifest", s.handleManifest)
	mux.HandleFunc("/manifest/recalculate", s.handleRecalculate)
	mux.HandleFunc("/files/", s.handleFiles)
	return mux
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Server starting on %s", addr)
	log.Printf("  GET  /manifest             - get current manifest")
	log.Printf("  POST /manifest/recalculate - recalculate manifest")
	log.Printf("  GET  /files/{path}         - download file")
	return http.ListenAndServe(addr, s.Handler())
}
