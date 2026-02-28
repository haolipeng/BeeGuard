package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart cloudsec-agent service",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("systemctl", "restart", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		cobra.CheckErr(c.Run())
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
