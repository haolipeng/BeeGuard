package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func unsetKey(key string) error {
	configMap := viper.AllSettings()
	delete(configMap, key)
	buf := bytes.NewBuffer(nil)
	for k, v := range configMap {
		fmt.Fprintf(buf, "%v = %v\n", k, v)
	}
	err := viper.ReadConfig(buf)
	if err != nil {
		return err
	}
	return viper.WriteConfig()
}

var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Remove runtime configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().NFlag() == 0 {
			cobra.CheckErr(cmd.Help())
			return
		}
		cmd.Flags().Visit(func(f *pflag.Flag) {
			switch f.Name {
			case "server":
				cobra.CheckErr(unsetKey("specified_server"))
			case "id":
				cobra.CheckErr(unsetKey("specified_agent_id"))
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(unsetCmd)
	unsetCmd.Flags().Bool("server", false, "remove server address")
	unsetCmd.Flags().Bool("id", false, "remove agent ID")
}
