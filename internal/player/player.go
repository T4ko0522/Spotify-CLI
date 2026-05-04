package player

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zmb3/spotify/v2"
)

var ErrAlreadyPlaying = errors.New("already playing")
var ErrAlreadyPaused = errors.New("already paused")

// AmbiguousDeviceError is returned by EnsureDevice when no device is
// currently active and more than one is available, so the player can't
// pick one safely on its own. The package is intentionally I/O-free —
// the caller (CLI or TUI) decides how to surface the choice and is
// expected to call TransferPlayback with the picked device. This is
// what lets the TUI invoke player methods without fighting bubbletea
// over stdin.
type AmbiguousDeviceError struct {
	Devices []spotify.PlayerDevice
}

func (e *AmbiguousDeviceError) Error() string {
	return fmt.Sprintf("multiple devices available (%d); select one", len(e.Devices))
}

type Player struct {
	Client *spotify.Client
}

func New(client *spotify.Client) *Player {
	return &Player{Client: client}
}

// EnsureDevice confirms an active Spotify device exists. If none is
// active and exactly one device is available, playback is transferred
// to it automatically. With multiple available devices it returns
// AmbiguousDeviceError carrying the list, so the caller can run a UI.
// This function never reads stdin nor writes to stdout.
func (p *Player) EnsureDevice(ctx context.Context) error {
	state, err := p.PlayerState(ctx)
	if err != nil {
		return err
	}
	if state.Device.ID != "" {
		return nil
	}

	devices, err := p.Devices(ctx)
	if err != nil {
		return err
	}

	switch len(devices) {
	case 0:
		return errors.New("no devices found. Open Spotify on a device first")
	case 1:
		return p.Client.TransferPlayback(ctx, devices[0].ID, true)
	default:
		return &AmbiguousDeviceError{Devices: devices}
	}
}

// TransferPlayback moves active playback to the specified device.
func (p *Player) TransferPlayback(ctx context.Context, deviceID spotify.ID) error {
	if err := p.Client.TransferPlayback(ctx, deviceID, true); err != nil {
		return fmt.Errorf("failed to transfer playback: %w", err)
	}
	return nil
}

func (p *Player) Play(ctx context.Context) error {
	if err := p.EnsureDevice(ctx); err != nil {
		return err
	}
	state, err := p.PlayerState(ctx)
	if err != nil {
		return err
	}
	if state.Playing {
		return ErrAlreadyPlaying
	}
	if err := p.Client.Play(ctx); err != nil {
		return fmt.Errorf("failed to resume playback: %w", err)
	}
	return nil
}

func (p *Player) Pause(ctx context.Context) error {
	if err := p.EnsureDevice(ctx); err != nil {
		return err
	}
	state, err := p.PlayerState(ctx)
	if err != nil {
		return err
	}
	if !state.Playing {
		return ErrAlreadyPaused
	}
	if err := p.Client.Pause(ctx); err != nil {
		return fmt.Errorf("failed to pause playback: %w", err)
	}
	return nil
}

func (p *Player) Next(ctx context.Context) error {
	if err := p.EnsureDevice(ctx); err != nil {
		return err
	}
	if err := p.Client.Next(ctx); err != nil {
		return fmt.Errorf("failed to skip to next track: %w", err)
	}
	return nil
}

func (p *Player) Previous(ctx context.Context) error {
	if err := p.EnsureDevice(ctx); err != nil {
		return err
	}
	if err := p.Client.Previous(ctx); err != nil {
		return fmt.Errorf("failed to skip to previous track: %w", err)
	}
	return nil
}

func (p *Player) SetVolume(ctx context.Context, percent int) error {
	if err := p.EnsureDevice(ctx); err != nil {
		return err
	}
	if err := p.Client.Volume(ctx, percent); err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}
	return nil
}

func (p *Player) NowPlaying(ctx context.Context) (*spotify.CurrentlyPlaying, error) {
	playing, err := p.Client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get currently playing: %w", err)
	}
	return playing, nil
}

func (p *Player) PlayerState(ctx context.Context) (*spotify.PlayerState, error) {
	state, err := p.Client.PlayerState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get player state: %w", err)
	}
	return state, nil
}

func (p *Player) Devices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	devices, err := p.Client.PlayerDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	return devices, nil
}

func FormatArtists(artists []spotify.SimpleArtist) string {
	names := make([]string, len(artists))
	for i, a := range artists {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func FormatProgress(ms, total int) string {
	current := fmt.Sprintf("%d:%02d", ms/60000, (ms/1000)%60)
	end := fmt.Sprintf("%d:%02d", total/60000, (total/1000)%60)
	return current + " / " + end
}
