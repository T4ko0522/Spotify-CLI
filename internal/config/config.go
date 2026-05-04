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

	// Merge with existing config so unrelated fields (e.g. img_size) are
	// preserved when only the client ID is being updated.
	var cfg configData
	if existing, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(existing, &cfg)
	}
	cfg.ClientID = clientID

	data, err := json.MarshalIndent(cfg, "", "  ")
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

	var cfg configData
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	cfg.ImgSize = imgSize

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	applyPreset(imgSize)
	return nil
}
