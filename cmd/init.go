package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/T4ko0522/spotify-cli/internal/auth"
	"github.com/T4ko0522/spotify-cli/internal/config"
	"github.com/spf13/cobra"
)

func (a *App) newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure Client ID and authenticate with Spotify",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			// Step 1: 既存の Client ID があれば検証して維持するか確認
			if err := config.Load(); err == nil && config.ClientID != "" {
				fmt.Printf("Existing Client ID found: %s\n", config.ClientID)
				fmt.Print("Use this Client ID? [Y/n]: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				s := strings.TrimSpace(strings.ToLower(input))
				if s == "" || s == "y" || s == "yes" {
					// 既存の Client ID を維持してログインへ
					fmt.Println("Using existing Client ID.")
					return auth.Login()
				}
			}

			// Step 2: 新規または差し替えで Client ID 入力
			fmt.Print("Enter your Spotify Client ID: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			clientID := strings.TrimSpace(input)
			if clientID == "" {
				return fmt.Errorf("client ID cannot be empty")
			}
			if err := config.Save(clientID); err != nil {
				return err
			}
			fmt.Println("Client ID saved.")

			// Step 3: config をロードして Login 実行
			if err := config.Load(); err != nil {
				return err
			}
			return auth.Login()
		},
	}
}
