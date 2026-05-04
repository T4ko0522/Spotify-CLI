package cmd

import (
	"errors"
	"fmt"

	"github.com/T4ko0522/spotify-cli/internal/player"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:     "stop",
	Aliases: []string{"s"},
	Short:   "Pause playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		err := runWithDeviceSelection(ctx, spotifyPlayer, func() error {
			return spotifyPlayer.Pause(ctx)
		})
		if err != nil {
			if errors.Is(err, player.ErrAlreadyPaused) {
				fmt.Println("Already paused.")
				return nil
			}
			return err
		}
		fmt.Println("Playback paused.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
