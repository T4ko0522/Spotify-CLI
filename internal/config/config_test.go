package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// readBack reads cfg from disk via os.ReadFile (which follows symlinks).
func readBack(t *testing.T, path string) configData {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	var c configData
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return c
}

func TestWriteConfigAtomic_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := writeConfigAtomic(path, configData{ClientID: "abc", ImgSize: "small"}); err != nil {
		t.Fatalf("writeConfigAtomic: %v", err)
	}

	got := readBack(t, path)
	if got.ClientID != "abc" || got.ImgSize != "small" {
		t.Fatalf("unexpected content: %+v", got)
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("expected regular file, got symlink")
	}
}

func TestWriteConfigAtomic_ExistingRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte(`{"client_id":"old"}`), 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := writeConfigAtomic(path, configData{ClientID: "new", ImgSize: "large"}); err != nil {
		t.Fatalf("writeConfigAtomic: %v", err)
	}

	got := readBack(t, path)
	if got.ClientID != "new" || got.ImgSize != "large" {
		t.Fatalf("unexpected content: %+v", got)
	}
}

// TestWriteConfigAtomic_SymlinkPreserved is the regression test for the
// PR2 review. Before the fix, writing to a symlinked config replaced
// the symlink with a regular file, silently breaking dotfiles setups.
func TestWriteConfigAtomic_SymlinkPreserved(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires elevated privileges or developer mode")
	}

	root := t.TempDir()
	realDir := filepath.Join(root, "dotfiles", "spt")
	cfgDir := filepath.Join(root, ".config", "spt")
	if err := os.MkdirAll(realDir, 0700); err != nil {
		t.Fatalf("mkdir realDir: %v", err)
	}
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("mkdir cfgDir: %v", err)
	}

	realPath := filepath.Join(realDir, "config.json")
	linkPath := filepath.Join(cfgDir, "config.json")

	if err := os.WriteFile(realPath, []byte(`{"client_id":"old","img_size":"small"}`), 0600); err != nil {
		t.Fatalf("seed real: %v", err)
	}
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := writeConfigAtomic(linkPath, configData{ClientID: "new", ImgSize: "large"}); err != nil {
		t.Fatalf("writeConfigAtomic: %v", err)
	}

	// The symlink itself must still be a symlink pointing at realPath.
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("lstat linkPath: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was replaced with a regular file (regression)")
	}
	resolved, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if resolved != realPath {
		t.Fatalf("symlink target changed: got %q want %q", resolved, realPath)
	}

	// Both the symlink view and the real file must contain the new
	// content (i.e. the underlying file was written, not replaced).
	if got := readBack(t, linkPath); got.ClientID != "new" || got.ImgSize != "large" {
		t.Fatalf("via symlink: %+v", got)
	}
	if got := readBack(t, realPath); got.ClientID != "new" || got.ImgSize != "large" {
		t.Fatalf("via real path: %+v", got)
	}

	// Make sure no temp file leaked beside the symlink (it must be in
	// the real dir).
	entries, err := os.ReadDir(cfgDir)
	if err != nil {
		t.Fatalf("readdir cfgDir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "config.json" {
			t.Errorf("unexpected leftover in cfgDir: %s", e.Name())
		}
	}
}

// TestWriteConfigAtomic_BrokenSymlink covers a symlink whose target
// does not exist yet. The symlink must be preserved and the target
// file should be created on the path the link points at.
func TestWriteConfigAtomic_BrokenSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires elevated privileges or developer mode")
	}

	root := t.TempDir()
	realDir := filepath.Join(root, "dotfiles", "spt")
	cfgDir := filepath.Join(root, ".config", "spt")
	if err := os.MkdirAll(realDir, 0700); err != nil {
		t.Fatalf("mkdir realDir: %v", err)
	}
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("mkdir cfgDir: %v", err)
	}

	realPath := filepath.Join(realDir, "config.json") // intentionally not created
	linkPath := filepath.Join(cfgDir, "config.json")

	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := writeConfigAtomic(linkPath, configData{ClientID: "fresh", ImgSize: "medium"}); err != nil {
		t.Fatalf("writeConfigAtomic: %v", err)
	}

	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("lstat linkPath: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was replaced with a regular file (regression)")
	}
	if _, err := os.Stat(realPath); err != nil {
		t.Fatalf("real target was not created: %v", err)
	}
	if got := readBack(t, realPath); got.ClientID != "fresh" || got.ImgSize != "medium" {
		t.Fatalf("real path content: %+v", got)
	}
}

func TestSaveSettings_PropagatesCorruptConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if runtime.GOOS == "darwin" {
		// os.UserConfigDir on darwin uses ~/Library/Application Support.
		// Override HOME so it points at our temp dir.
		t.Setenv("HOME", dir)
	}

	// Pre-populate spt/config.json with garbage so Unmarshal fails.
	configDir, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := SaveSettings("large"); err == nil {
		t.Fatalf("expected SaveSettings to surface the parse error, got nil")
	}
}
