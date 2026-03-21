package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cloudsec-agent service status",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("systemctl", "status", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		// systemctl status returns non-zero if service is not running,
		// don't treat that as a fatal error
		c.Run()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
