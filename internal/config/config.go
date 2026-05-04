package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ImgSizePreset struct {
	Cols int
	Rows int
}

var ImgSizePresets = map[string]ImgSizePreset{
	"small":  {Cols: 16, Rows: 8},
	"medium": {Cols: 20, Rows: 10},
	"large":  {Cols: 28, Rows: 14},
}

var ImgSizeNames = []string{"small", "medium", "large"}

var (
	ClientID string
	ImgSize  = "medium"
	ImgCols  = ImgSizePresets["medium"].Cols
	ImgRows  = ImgSizePresets["medium"].Rows
)

type configData struct {
	ClientID string `json:"client_id"`
	ImgSize  string `json:"img_size,omitempty"`
}

func applyPreset(name string) {
	if p, ok := ImgSizePresets[name]; ok {
		ImgSize = name
		ImgCols = p.Cols
		ImgRows = p.Rows
	}
}

// Dir returns the spt config directory, creating it if needed.
// Uses os.UserConfigDir which resolves to:
//   - Windows: %AppData%
//   - macOS:   ~/Library/Application Support
//   - Linux:   $XDG_CONFIG_HOME or ~/.config
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	dir := filepath.Join(base, "spt")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

func configPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Save(clientID string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(configData{ClientID: clientID}, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}
	return nil
}

func Load() error {
	// 1. Try config file
	path, err := configPath()
	if err == nil {
		if data, err := os.ReadFile(path); err == nil {
			var cfg configData
			if err := json.Unmarshal(data, &cfg); err == nil {
				if cfg.ClientID != "" {
					ClientID = cfg.ClientID
				}
				if cfg.ImgSize != "" {
					applyPreset(cfg.ImgSize)
				}
				if ClientID != "" {
					return nil
				}
			}
		}
	}

	// 2. Fall back to environment variable
	ClientID = os.Getenv("SPOTIFY_CLIENT_ID")
	if ClientID != "" {
		return nil
	}

	return fmt.Errorf("SPOTIFY_CLIENT_ID is not configured. Run 'spt init' to set up")
}

func SaveSettings(imgSize string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Read-modify-write must surface errors; previously a missing or
	// corrupt file silently produced a zero-value cfg, which then wrote
	// back an empty client_id and forced the user to re-run `spt init`.
	cfg, err := readConfig(path)
	if err != nil {
		return err
	}
	cfg.ImgSize = imgSize

	if err := writeConfigAtomic(path, cfg); err != nil {
		return err
	}

	applyPreset(imgSize)
	return nil
}

// readConfig loads the on-disk config. A missing file is treated as an
// empty config so first-time writers can compose one. Any other I/O or
// parse error is returned to the caller.
func readConfig(path string) (configData, error) {
	var cfg configData
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("cannot read config file: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("corrupt config file at %s: %w", path, err)
	}
	return cfg, nil
}

// writeConfigAtomic writes cfg to path using a sibling temp file and a
// rename, so concurrent readers never see a partially-written file and
// a crash mid-write cannot leave the config truncated.
func writeConfigAtomic(path string, cfg configData) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temp config file: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot write temp config file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot chmod temp config file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot fsync temp config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("cannot rename config file: %w", err)
	}
	committed = true
	return nil
}
