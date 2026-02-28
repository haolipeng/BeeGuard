package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start cloudsec-agent service",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("systemctl", "start", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		cobra.CheckErr(c.Run())
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
