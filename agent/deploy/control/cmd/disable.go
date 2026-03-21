package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable cloudsec-agent from starting on boot",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("systemctl", "disable", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		cobra.CheckErr(c.Run())
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}
