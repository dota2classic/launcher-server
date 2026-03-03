package manifest

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type HashedFile struct {
	FilePath string   `json:"path"`
	Hash     string   `json:"hash"`
	FileSize int64    `json:"size"`
	Mode     FileMode `json:"mode"` // "exact" = must match hash, "existing" = must exist, hash may differ
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
		Mode:     ModeExact,
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

	type fileJob struct {
		fullPath     string
		relativePath string
		mode         FileMode
	}

	// First pass: collect all files to hash
	var jobs []fileJob
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

		// Skip meta files
		if relativePath == ManifestIgnoreFile || relativePath == ManifestFile {
			return nil
		}

		// Check ignore rules
		matched, isSoft := ignoreRules.Match(relativePath)
		if matched && !isSoft {
			return nil
		}

		mode := ModeExact
		if matched && isSoft {
			mode = ModeExisting
		}

		jobs = append(jobs, fileJob{fullPath, relativePath, mode})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Second pass: hash files in parallel
	type result struct {
		hf  *HashedFile
		err error
	}

	results := make([]result, len(jobs))
	jobCh := make(chan int, len(jobs))
	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)

	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobCh {
				job := jobs[idx]
				hf, err := HashFile(job.fullPath, job.relativePath)
				if err == nil {
					hf.Mode = job.mode
				}
				results[idx] = result{hf, err}
			}
		}()
	}
	wg.Wait()

	files := make([]*HashedFile, 0, len(jobs))
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		files = append(files, r.hf)
	}

	log.Printf("Hashed directory %s. Total files: %d", basePath, len(files))
	return &Manifest{Files: files}, nil
}
