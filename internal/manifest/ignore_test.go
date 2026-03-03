package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIgnoreFile(t *testing.T) {
	dir := t.TempDir()

	content := `# This is a comment
*.log
temp/

# Soft files
?config.json
?settings/*.ini
`
	err := os.WriteFile(filepath.Join(dir, ManifestIgnoreFile), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	rules, err := ParseIgnoreFile(dir)
	if err != nil {
		t.Fatal(err)
	}

	if rules == nil {
		t.Fatal("expected rules, got nil")
	}

	if len(rules.rules) != 4 {
		t.Errorf("expected 4 rules, got %d", len(rules.rules))
	}

	// Check first rule (ignore *.log)
	if rules.rules[0].Pattern != "*.log" || rules.rules[0].Soft {
		t.Errorf("unexpected rule[0]: %+v", rules.rules[0])
	}

	// Check soft rule
	if rules.rules[2].Pattern != "config.json" || !rules.rules[2].Soft {
		t.Errorf("unexpected rule[2]: %+v", rules.rules[2])
	}
}

func TestParseIgnoreFile_NotExists(t *testing.T) {
	dir := t.TempDir()

	rules, err := ParseIgnoreFile(dir)
	if err != nil {
		t.Fatal(err)
	}

	if rules != nil {
		t.Error("expected nil rules for missing file")
	}
}

func TestIgnoreRules_Match(t *testing.T) {
	rules := &IgnoreRules{
		rules: []IgnoreRule{
			{Pattern: "*.log", Soft: false},
			{Pattern: "temp/", Soft: false},
			{Pattern: "config.json", Soft: true},
			{Pattern: "**/*.ini", Soft: true},
			{Pattern: "cache/", Soft: false},
			{Pattern: "important.dat", Soft: false},
			{Pattern: "important.dat", Soft: true}, // Later rule overrides
		},
	}

	tests := []struct {
		path        string
		wantMatched bool
		wantSoft    bool
	}{
		// Ignored files
		{"app.log", true, false},
		{"logs/debug.log", true, false},
		{"temp/file.txt", true, false},
		{"temp/subdir/file.txt", true, false},
		{"cache/data.bin", true, false},

		// Soft files
		{"config.json", true, true},
		{"settings/game.ini", true, true},
		{"deep/nested/config.ini", true, true},

		// Required files (no match)
		{"main.exe", false, false},
		{"data/game.dat", false, false},
		{"readme.txt", false, false},

		// Last rule wins
		{"important.dat", true, true},
	}

	for _, tt := range tests {
		matched, isSoft := rules.Match(tt.path)
		if matched != tt.wantMatched || isSoft != tt.wantSoft {
			t.Errorf("Match(%q) = (%v, %v), want (%v, %v)",
				tt.path, matched, isSoft, tt.wantMatched, tt.wantSoft)
		}
	}
}

func TestIgnoreRules_Match_Nil(t *testing.T) {
	var rules *IgnoreRules
	matched, isSoft := rules.Match("anything.txt")
	if matched || isSoft {
		t.Error("nil rules should not match anything")
	}
}

func TestCreateManifestWithIgnore(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.exe":           "main executable",
		"data/game.dat":      "game data",
		"config.json":        "config content",
		"settings/video.ini": "video settings",
		"debug.log":          "debug logs",
		"temp/cache.bin":     "cache data",
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create .manifestignore
	ignoreContent := `*.log
temp/
?config.json
?settings/*.ini
`
	err := os.WriteFile(filepath.Join(dir, ManifestIgnoreFile), []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	manifest, err := CreateManifest(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Build a map of files in manifest
	manifestFiles := make(map[string]*HashedFile)
	for _, f := range manifest.Files {
		manifestFiles[f.FilePath] = f
	}

	// Check required files are present
	if _, ok := manifestFiles["main.exe"]; !ok {
		t.Error("main.exe should be in manifest")
	}
	if _, ok := manifestFiles["data/game.dat"]; !ok {
		t.Error("data/game.dat should be in manifest")
	}

	// Check ignored files are not present
	if _, ok := manifestFiles["debug.log"]; ok {
		t.Error("debug.log should NOT be in manifest")
	}
	if _, ok := manifestFiles["temp/cache.bin"]; ok {
		t.Error("temp/cache.bin should NOT be in manifest")
	}

	// Check existing files are present with correct mode
	if f, ok := manifestFiles["config.json"]; !ok {
		t.Error("config.json should be in manifest")
	} else if f.Mode != ModeExisting {
		t.Errorf("config.json mode = %q, want %q", f.Mode, ModeExisting)
	}

	if f, ok := manifestFiles["settings/video.ini"]; !ok {
		t.Error("settings/video.ini should be in manifest")
	} else if f.Mode != ModeExisting {
		t.Errorf("settings/video.ini mode = %q, want %q", f.Mode, ModeExisting)
	}

	// Check .manifestignore itself is not in manifest
	if _, ok := manifestFiles[ManifestIgnoreFile]; ok {
		t.Error(".manifestignore should NOT be in manifest")
	}

	// Check exact files have correct mode
	if f := manifestFiles["main.exe"]; f.Mode != ModeExact {
		t.Errorf("main.exe mode = %q, want %q", f.Mode, ModeExact)
	}
}
