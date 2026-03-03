package server

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"launcher-host/internal/manifest"
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

func (s *Server) Recalculate() error {
	return s.recalculateManifest()
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
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		json.NewEncoder(gz).Encode(s.currentManifest)
	} else {
		json.NewEncoder(w).Encode(s.currentManifest)
	}
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


func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/manifest", s.handleManifest)
	mux.HandleFunc("/manifest/recalculate", s.handleRecalculate)
	return mux
}

func (s *Server) newHTTPServer(addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		// No WriteTimeout: file downloads can be large and take time.
	}
}

func (s *Server) logRoutes(addr, scheme string) {
	log.Printf("Server starting on %s://%s", scheme, addr)
	log.Printf("  GET  /manifest             - get current manifest")
	log.Printf("  POST /manifest/recalculate - recalculate manifest")
}

func (s *Server) ListenAndServe(addr string) error {
	s.logRoutes(addr, "http")
	return s.newHTTPServer(addr).ListenAndServe()
}

func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	s.logRoutes(addr, "https")
	srv := s.newHTTPServer(addr)
	srv.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		// NextProtos is set automatically by net/http when HTTP/2 is configured.
	}
	// Go's net/http automatically enables HTTP/2 for ListenAndServeTLS.
	return srv.ListenAndServeTLS(certFile, keyFile)
}
