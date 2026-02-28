package cmd

import (
	"os/exec"

	"github.com/spf13/cobra"
)

var serviceReloadCmd = &cobra.Command{
	Use:   "service-reload",
	Short: "Reload systemd daemon configuration",
	Run: func(cmd *cobra.Command, args []string) {
		exec.Command("systemctl", "daemon-reload").Run()
	},
}

func init() {
	rootCmd.AddCommand(serviceReloadCmd)
}
