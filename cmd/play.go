package cmd

import (
	"errors"
	"fmt"

	"github.com/T4ko0522/spotify-cli/internal/player"
	"github.com/spf13/cobra"
)

func (a *App) newPlayCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "play",
		Aliases: []string{"p"},
		Short:   "Resume playback",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			err := runWithDeviceSelection(ctx, a.player, func() error {
				return a.player.Play(ctx)
			})
			if err != nil {
				if errors.Is(err, player.ErrAlreadyPlaying) {
					fmt.Println("Already playing.")
					return nil
				}
				return err
			}
			fmt.Println("Playback resumed.")
			return nil
		},
	}
}
