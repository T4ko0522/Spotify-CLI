//go:build windows

package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/T4ko0522/spotify-cli/internal/config"
	"golang.org/x/sys/windows"
)

type fileLock struct {
	f *os.File
}

func acquireTokenLock() (*fileLock, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "token.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open token lock file: %w", err)
	}
	var ol windows.Overlapped
	if err := windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1, 0,
		&ol,
	); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("cannot acquire token lock: %w", err)
	}
	return &fileLock{f: f}, nil
}

func (l *fileLock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	var ol windows.Overlapped
	_ = windows.UnlockFileEx(windows.Handle(l.f.Fd()), 0, 1, 0, &ol)
	err := l.f.Close()
	l.f = nil
	return err
}
