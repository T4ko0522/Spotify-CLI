package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/T4ko0522/spotify-cli/internal/auth"
	"github.com/T4ko0522/spotify-cli/internal/config"
	"github.com/T4ko0522/spotify-cli/internal/player"
	"github.com/T4ko0522/spotify-cli/internal/tui"
	"github.com/spf13/cobra"
	spotify "github.com/zmb3/spotify/v2"
)

type App struct {
	client     *spotify.Client
	player     *player.Player
	showLyrics bool
}

func NewRootCommand(version string) *cobra.Command {
	app := &App{}
	rootCmd := &cobra.Command{
		Use:           "spt",
		Short:         "Spotify CLI controller",
		Long:          "A command-line tool to control Spotify playback.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.showLyrics {
				return tui.RunLyrics(app.client)
			}
			return tui.Run(app.client)
		},
		PersistentPreRunE: app.loadSpotify,
	}

	rootCmd.Flags().BoolVarP(&app.showLyrics, "lyrics", "l", false, "Show lyrics for the currently playing track")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	app.addCommands(rootCmd)

	return rootCmd
}

func (a *App) addCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		a.newDevicesCommand(),
		a.newInitCommand(),
		a.newNextCommand(),
		a.newNowCommand(),
		a.newPauseCommand(),
		a.newPlayCommand(),
		a.newPrevCommand(),
		a.newSettingsCommand(),
		a.newVolumeCommand(),
	)
}

func (a *App) loadSpotify(cmd *cobra.Command, args []string) error {
	// init/settings commands handle config on their own.
	if cmd.Name() == "init" || cmd.Name() == "settings" {
		return nil
	}
	if err := config.Load(); err != nil {
		return err
	}
	ctx := context.Background()
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("%w\nRun 'spt init' to authenticate", err)
	}
	a.client = spotify.New(httpClient)
	a.player = player.New(a.client)
	return nil
}

func Execute(version string) {
	if err := NewRootCommand(version).Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
