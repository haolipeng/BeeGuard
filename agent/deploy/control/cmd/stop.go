package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop cloudsec-agent service",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("systemctl", "stop", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		cobra.CheckErr(c.Run())
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
