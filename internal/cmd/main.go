package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"launcher-host/internal/manifest"
)

func main() {
	var filesPath string
	if len(os.Args) > 1 {
		filesPath = os.Args[1]
	} else {
		filesPath = os.Getenv("LAUNCHER_FILES_PATH")
	}
	if filesPath == "" {
		log.Fatal("usage: launcher-host <files-path>  or set LAUNCHER_FILES_PATH")
	}

	m, err := manifest.CreateManifest(filesPath)
	if err != nil {
		log.Fatalf("Failed to create manifest: %v", err)
	}

	outPath := filepath.Join(filesPath, manifest.ManifestFile)
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", outPath, err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(m); err != nil {
		log.Fatalf("Failed to write manifest: %v", err)
	}

	log.Printf("Written manifest with %d files to %s", len(m.Files), outPath)
}
