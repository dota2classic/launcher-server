package manifest

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type HashedFile struct {
	FilePath string   `json:"path"`
	Hash     string   `json:"hash"`
	FileSize int64    `json:"size"`
	Mode     FileMode `json:"mode,omitempty"` // empty = required, "soft" = download only if missing
}

type Manifest struct {
	Files []*HashedFile `json:"files"`
}

func HashFile(fullPath, relativePath string) (*HashedFile, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Error opening file: %s", err)
		return nil, err
	}
	defer f.Close()

	h := md5.New()
	size, err := io.Copy(h, f)
	if err != nil {
		log.Printf("Error hashing file: %s", err)
		return nil, err
	}

	return &HashedFile{
		FilePath: relativePath,
		Hash:     hex.EncodeToString(h.Sum(nil)),
		FileSize: size,
	}, nil
}

// hashFileReadAll is the old implementation for benchmarking comparison
func hashFileReadAll(fullPath, relativePath string) (*HashedFile, error) {
	h := md5.New()

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	h.Write(content)

	return &HashedFile{
		FilePath: relativePath,
		Hash:     hex.EncodeToString(h.Sum(nil)),
		FileSize: int64(len(content)),
	}, nil
}

func CreateManifest(basePath string) (*Manifest, error) {
	// Parse .manifestignore if it exists
	ignoreRules, err := ParseIgnoreFile(basePath)
	if err != nil {
		return nil, err
	}

	var files []*HashedFile

	err = filepath.WalkDir(basePath, func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(basePath, fullPath)
		if err != nil {
			return err
		}
		// Normalize to forward slashes for cross-platform compatibility
		relativePath = filepath.ToSlash(relativePath)

		// Skip the .manifestignore file itself
		if relativePath == ManifestIgnoreFile {
			return nil
		}

		// Check ignore rules
		matched, isSoft := ignoreRules.Match(relativePath)
		if matched && !isSoft {
			// Ignored file - skip entirely
			return nil
		}

		hf, err := HashFile(fullPath, relativePath)
		if err != nil {
			return err
		}

		// Mark soft files
		if matched && isSoft {
			hf.Mode = ModeSoft
		}

		files = append(files, hf)
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Hashed directory %s. Total files: %d", basePath, len(files))
	return &Manifest{Files: files}, nil
}
