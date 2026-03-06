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

	registryPath := filepath.Join(filesPath, manifest.RegistryFile)
	rf, err := os.Open(registryPath)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", registryPath, err)
	}
	var registry manifest.Registry
	if err := json.NewDecoder(rf).Decode(&registry); err != nil {
		rf.Close()
		log.Fatalf("Failed to parse %s: %v", registryPath, err)
	}
	rf.Close()

	if len(registry.Packages) == 0 {
		log.Fatal("registry.json contains no packages")
	}

	for _, pkg := range registry.Packages {
		pkgPath := filepath.Join(filesPath, pkg.Folder)

		m, err := manifest.CreateManifest(pkgPath)
		if err != nil {
			log.Fatalf("[%s] Failed to create manifest: %v", pkg.ID, err)
		}

		outPath := filepath.Join(pkgPath, manifest.ManifestFile)
		f, err := os.Create(outPath)
		if err != nil {
			log.Fatalf("[%s] Failed to create %s: %v", pkg.ID, outPath, err)
		}

		if err := json.NewEncoder(f).Encode(m); err != nil {
			f.Close()
			log.Fatalf("[%s] Failed to write manifest: %v", pkg.ID, err)
		}
		f.Close()

		log.Printf("[%s] Written manifest with %d files to %s", pkg.ID, len(m.Files), outPath)
	}
}
