package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/T4ko0522/spotify-cli/internal/config"
	"golang.org/x/oauth2"
)

func tokenPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

// SaveToken writes the OAuth token to disk atomically: marshal into a
// sibling temp file, fsync, then rename over the destination. PKCE rotates
// the refresh token on every refresh, so a partially-written or lost save
// invalidates future refreshes and forces the user to re-run `spt init`.
func SaveToken(token *oauth2.Token) error {
	path, err := tokenPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal token: %w", err)
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "token-*.json.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temp token file: %w", err)
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
		return fmt.Errorf("cannot write temp token file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot chmod temp token file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot fsync temp token file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temp token file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("cannot rename token file: %w", err)
	}
	committed = true
	return nil
}

func LoadToken() (*oauth2.Token, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no saved token found. Run 'spt init' first")
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("corrupt token file: %w", err)
	}
	return &token, nil
}
