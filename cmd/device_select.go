package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/T4ko0522/spotify-cli/internal/player"
	"github.com/zmb3/spotify/v2"
)

// runWithDeviceSelection executes action and, if it fails with
// player.AmbiguousDeviceError, prompts the user on stdin for a device
// choice, transfers playback to the selected device, and retries action
// once. CLI commands wrap their player calls with this so the player
// package can stay free of stdin/stdout, which is required for the TUI
// path that runs through bubbletea.
func runWithDeviceSelection(ctx context.Context, p *player.Player, action func() error) error {
	err := action()
	var ambig *player.AmbiguousDeviceError
	if !errors.As(err, &ambig) {
		return err
	}

	selected, err := promptDeviceSelection(ambig.Devices)
	if err != nil {
		return err
	}
	fmt.Printf("Transferring playback to %s...\n", selected.Name)
	if err := p.TransferPlayback(ctx, selected.ID); err != nil {
		return err
	}
	return action()
}

func promptDeviceSelection(devices []spotify.PlayerDevice) (spotify.PlayerDevice, error) {
	fmt.Println("Multiple devices found:")
	for i, d := range devices {
		fmt.Printf("  %d: %s (%s)\n", i+1, d.Name, d.Type)
	}
	fmt.Print("Select device number: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return spotify.PlayerDevice{}, fmt.Errorf("failed to read selection: %w", err)
	}
	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || choice < 1 || choice > len(devices) {
		return spotify.PlayerDevice{}, fmt.Errorf("invalid selection: %q", strings.TrimSpace(line))
	}
	return devices[choice-1], nil
}
