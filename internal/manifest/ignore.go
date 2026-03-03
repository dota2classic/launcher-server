package manifest

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type FileMode string

const (
	ModeExact    FileMode = "exact"    // File must match exactly (same hash)
	ModeExisting FileMode = "existing" // File must exist, but hash may differ; downloaded if missing
)

const ManifestIgnoreFile = ".manifestignore"
const ManifestFile = "manifest.json"

type IgnoreRule struct {
	Pattern string
	Soft    bool // true = soft file, false = ignored file
}

type IgnoreRules struct {
	rules []IgnoreRule
}

// ParseIgnoreFile reads and parses a .manifestignore file.
// Returns nil if the file doesn't exist.
func ParseIgnoreFile(basePath string) (*IgnoreRules, error) {
	ignorePath := filepath.Join(basePath, ManifestIgnoreFile)

	f, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	rules := &IgnoreRules{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := IgnoreRule{}

		// Check for soft prefix
		if strings.HasPrefix(line, "?") {
			rule.Soft = true
			rule.Pattern = strings.TrimPrefix(line, "?")
		} else {
			rule.Soft = false
			rule.Pattern = line
		}

		// Normalize pattern
		rule.Pattern = filepath.ToSlash(rule.Pattern)

		rules.rules = append(rules.rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// Match checks if a path matches any rule and returns:
// - (true, true) if path matches a soft rule
// - (true, false) if path matches an ignore rule
// - (false, false) if path doesn't match any rule
func (ir *IgnoreRules) Match(relativePath string) (matched bool, isSoft bool) {
	if ir == nil {
		return false, false
	}

	// Normalize path
	relativePath = filepath.ToSlash(relativePath)

	// Check rules in order (last match wins, like gitignore)
	for _, rule := range ir.rules {
		if matchPattern(rule.Pattern, relativePath) {
			matched = true
			isSoft = rule.Soft
		}
	}

	return matched, isSoft
}

// matchPattern checks if a path matches a pattern.
// Supports:
// - Exact matches: "file.txt"
// - Wildcards: "*.log", "temp/*"
// - Directory matches: "dir/" matches all files under dir
// - Double wildcards: "**/*.txt" matches any depth
func matchPattern(pattern, path string) bool {
	// Directory pattern (ends with /)
	if strings.HasSuffix(pattern, "/") {
		dir := strings.TrimSuffix(pattern, "/")
		if path == dir || strings.HasPrefix(path, dir+"/") {
			return true
		}
		return false
	}

	// Handle ** pattern for any depth matching
	if strings.Contains(pattern, "**") {
		return matchDoubleWildcard(pattern, path)
	}

	// Use filepath.Match for simple patterns
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Also try matching just the filename for patterns without path separator
	if !strings.Contains(pattern, "/") {
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}

	return false
}

// matchDoubleWildcard handles ** patterns
func matchDoubleWildcard(pattern, path string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Multiple ** not supported, fall back to simple match
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	prefix := parts[0]
	suffix := parts[1]

	// Remove leading slash from suffix if present
	suffix = strings.TrimPrefix(suffix, "/")

	// Check prefix
	if prefix != "" {
		prefix = strings.TrimSuffix(prefix, "/")
		if !strings.HasPrefix(path, prefix+"/") && path != prefix {
			return false
		}
	}

	// Check suffix
	if suffix != "" {
		// Try to match suffix against the path or any component
		if strings.HasSuffix(path, "/"+suffix) || strings.HasSuffix(path, suffix) {
			return true
		}
		// Also try glob match on filename
		matched, _ := filepath.Match(suffix, filepath.Base(path))
		return matched
	}

	return true
}
