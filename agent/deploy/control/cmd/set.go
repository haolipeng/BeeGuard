package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set runtime configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().NFlag() == 0 {
			cobra.CheckErr(cmd.Help())
			return
		}
		cmd.Flags().Visit(func(f *pflag.Flag) {
			switch f.Name {
			case "server":
				viper.Set("SPECIFIED_SERVER", f.Value.String())
			case "id":
				viper.Set("SPECIFIED_AGENT_ID", f.Value.String())
			}
			cobra.CheckErr(viper.WriteConfig())
		})
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().String("server", "", "server address (host:port)")
	setCmd.Flags().String("id", "", "agent ID")
}
