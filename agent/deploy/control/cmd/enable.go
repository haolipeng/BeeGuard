package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable cloudsec-agent to start on boot",
	Run: func(cmd *cobra.Command, args []string) {
		// Copy service file to systemd directory
		src := serviceFile
		dst := "/etc/systemd/system/" + serviceName + ".service"
		input, err := os.ReadFile(src)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("failed to read service file %s: %w", src, err))
		}
		err = os.WriteFile(dst, input, 0644)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("failed to write service file %s: %w", dst, err))
		}

		c := exec.Command("systemctl", "enable", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		cobra.CheckErr(c.Run())
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
