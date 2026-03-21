package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	serviceName  = "cloudsec-agent"
	agentWorkDir = "/opt/cloudsec/agent/"
	agentFile    = agentWorkDir + "bin/agent"
	agentConfig  = agentWorkDir + "agent.yaml"
	cfgFile      = agentWorkDir + "specified_env"
	serviceFile  = agentWorkDir + serviceName + ".service"
	pidFile      = "/var/run/" + serviceName + ".pid"
)

var rootCmd = &cobra.Command{
	Use:   "cloudsecctl",
	Short: "cloudsec-agent service control tool",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("props")
	viper.ReadInConfig()
}
