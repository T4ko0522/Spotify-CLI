package cmd

import (
	"fmt"
	"strconv"

	"github.com/T4ko0522/spotify-cli/internal/tui"
	"github.com/spf13/cobra"
)

func (a *App) newVolumeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "volume [0-100]",
		Aliases: []string{"v"},
		Short:   "Get or set playback volume",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return tui.RunVolume(a.player)
			}
			ctx := cmd.Context()
			vol, err := strconv.Atoi(args[0])
			if err != nil || vol < 0 || vol > 100 {
				return fmt.Errorf("volume must be a number between 0 and 100")
			}
			if err := runWithDeviceSelection(ctx, a.player, func() error {
				return a.player.SetVolume(ctx, vol)
			}); err != nil {
				return err
			}
			fmt.Printf("Volume: %d%%\n", vol)
			return nil
		},
	}
}
