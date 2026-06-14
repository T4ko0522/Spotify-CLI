package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) newDevicesCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "devices",
		Aliases: []string{"d"},
		Short:   "List available Spotify devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			devices, err := a.player.Devices(cmd.Context())
			if err != nil {
				return err
			}
			if len(devices) == 0 {
				fmt.Println("No active devices found. Open Spotify on a device first.")
				return nil
			}
			fmt.Printf("%-30s %-15s %-8s %s\n", "NAME", "TYPE", "VOLUME", "ACTIVE")
			fmt.Printf("%-30s %-15s %-8s %s\n", "----", "----", "------", "------")
			for _, d := range devices {
				active := " "
				if d.Active {
					active = "*"
				}
				fmt.Printf("%-30s %-15s %-8d %s\n", d.Name, d.Type, d.Volume, active)
			}
			return nil
		},
	}
}
