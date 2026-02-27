package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestFile(t *testing.T, dir string, name string, size int64) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Write random-ish data in chunks
	chunk := make([]byte, 64*1024)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	remaining := size
	for remaining > 0 {
		writeSize := int64(len(chunk))
		if writeSize > remaining {
			writeSize = remaining
		}
		_, err := f.Write(chunk[:writeSize])
		if err != nil {
			t.Fatal(err)
		}
		remaining -= writeSize
	}

	return path
}

func TestHashFileEquivalence(t *testing.T) {
	dir := t.TempDir()
	testFile := createTestFile(t, dir, "test.bin", 1024*1024) // 1MB

	oldResult, err := hashFileReadAll(testFile, "test.bin")
	if err != nil {
		t.Fatal(err)
	}

	newResult, err := HashFile(testFile, "test.bin")
	if err != nil {
		t.Fatal(err)
	}

	if oldResult.Hash != newResult.Hash {
		t.Errorf("Hash mismatch: old=%s new=%s", oldResult.Hash, newResult.Hash)
	}
	if oldResult.FileSize != newResult.FileSize {
		t.Errorf("Size mismatch: old=%d new=%d", oldResult.FileSize, newResult.FileSize)
	}
}

func BenchmarkHashFile_Streaming(b *testing.B) {
	benchmarkHashFile(b, HashFile)
}

func BenchmarkHashFile_ReadAll(b *testing.B) {
	benchmarkHashFile(b, hashFileReadAll)
}

func benchmarkHashFile(b *testing.B, hashFunc func(string, string) (*HashedFile, error)) {
	sizes := []struct {
		name string
		size int64
	}{
		{"1MB", 1 * 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			dir := b.TempDir()
			testFile := createBenchFile(b, dir, "test.bin", s.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := hashFunc(testFile, "test.bin")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func createBenchFile(b *testing.B, dir string, name string, size int64) string {
	b.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	chunk := make([]byte, 64*1024)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	remaining := size
	for remaining > 0 {
		writeSize := int64(len(chunk))
		if writeSize > remaining {
			writeSize = remaining
		}
		_, err := f.Write(chunk[:writeSize])
		if err != nil {
			b.Fatal(err)
		}
		remaining -= writeSize
	}

	return path
}
