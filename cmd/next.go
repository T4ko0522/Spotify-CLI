package cmd

import (
	"fmt"
	"time"

	"github.com/T4ko0522/spotify-cli/internal/player"
	"github.com/spf13/cobra"
)

func (a *App) newNextCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "next",
		Aliases: []string{"n"},
		Short:   "Skip to next track",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := runWithDeviceSelection(ctx, a.player, func() error {
				return a.player.Next(ctx)
			}); err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
			playing, err := a.player.NowPlaying(ctx)
			if err != nil {
				fmt.Println("Skipped to next track.")
				return nil
			}
			if playing.Item != nil {
				fmt.Printf("Now playing: %s - %s\n", playing.Item.Name, player.FormatArtists(playing.Item.Artists))
			}
			return nil
		},
	}
}
