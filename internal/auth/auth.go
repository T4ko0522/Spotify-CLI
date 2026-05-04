package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/T4ko0522/spotify-cli/internal/config"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectURI = "http://127.0.0.1:8888/callback"

var scopes = []string{
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
}

func newAuthenticator() *spotifyauth.Authenticator {
	return spotifyauth.New(
		spotifyauth.WithClientID(config.ClientID),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(scopes...),
	)
}

func generateVerifier() (string, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func Login() error {
	auth := newAuthenticator()

	verifier, err := generateVerifier()
	if err != nil {
		return fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	challenge := generateChallenge(verifier)

	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8888", Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("state"); s != state {
			errCh <- fmt.Errorf("state mismatch")
			_, _ = fmt.Fprint(w, "State mismatch. Please try again.")
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			errCh <- fmt.Errorf("authorization denied: %s", e)
			_, _ = fmt.Fprintf(w, "Authorization denied: %s", e)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			_, _ = fmt.Fprint(w, "No authorization code received.")
			return
		}
		_, _ = fmt.Fprint(w, "Login successful! You can close this tab.")
		codeCh <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)

	fmt.Println("Opening browser for Spotify login...")
	if err := openBrowser(url); err != nil {
		fmt.Printf("Open this URL in your browser:\n%s\n", url)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = server.Shutdown(context.Background())
		return err
	}

	_ = server.Shutdown(context.Background())

	token, err := auth.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("Logged in successfully.")
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default: // windows
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

func oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID: config.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.spotify.com/authorize",
			TokenURL:  "https://accounts.spotify.com/api/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectURI,
		Scopes:      scopes,
	}
}

// GetClient returns an HTTP client whose token source automatically
// refreshes and persists tokens to disk under an inter-process lock.
func GetClient(ctx context.Context) (*http.Client, error) {
	token, err := LoadToken()
	if err != nil {
		return nil, err
	}
	src := &persistingTokenSource{
		ctx:     ctx,
		cfg:     oauthConfig(),
		current: token,
	}
	return oauth2.NewClient(ctx, src), nil
}

// persistingTokenSource serializes Spotify PKCE token refreshes across
// concurrent processes. PKCE rotates the refresh token on each refresh,
// so two `spt` processes that load the same token and both attempt to
// refresh will invalidate each other's refresh token. We avoid that by
// taking a file lock for any refresh, re-reading the latest token from
// disk under the lock, and atomically writing the result back before
// releasing the lock.
type persistingTokenSource struct {
	ctx     context.Context
	cfg     *oauth2.Config
	mu      sync.Mutex
	current *oauth2.Token
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil && p.current.Valid() {
		return p.current, nil
	}

	lock, err := acquireTokenLock()
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Release() }()

	diskTok, err := LoadToken()
	if err != nil {
		return nil, err
	}

	// Another process may have already refreshed while we were waiting
	// for the lock. If the on-disk token is still valid, just use it.
	if diskTok.Valid() {
		p.current = diskTok
		return diskTok, nil
	}

	src := p.cfg.TokenSource(p.ctx, diskTok)
	fresh, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed; run 'spt init' to re-authenticate: %w", err)
	}

	if err := SaveToken(fresh); err != nil {
		// The refresh succeeded so the in-memory token is usable for
		// this invocation, but the rotated refresh token is now lost
		// from disk. Warn the user that the next run will likely need
		// re-authentication, instead of silently swallowing the error.
		fmt.Fprintf(os.Stderr, "warning: token refreshed but failed to persist: %v\n", err)
		fmt.Fprintln(os.Stderr, "the next 'spt' invocation may require 'spt init' to re-authenticate")
	}

	p.current = fresh
	return fresh, nil
}
