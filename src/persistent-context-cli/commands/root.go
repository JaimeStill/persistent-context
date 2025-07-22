package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	webURL  string
	directMode bool
)

var rootCmd = &cobra.Command{
	Use:   "persistent-context-cli",
	Short: "CLI tool for investigating and managing the persistent context system",
	Long: `A command-line interface for the Persistent Context system that provides
direct access to memory inspection, consolidation testing, and system monitoring.

This tool can operate in two modes:
- Direct mode: Connects directly to Qdrant and Ollama (default)
- HTTP mode: Connects through the web service API`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.persistent-context/cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&webURL, "web-url", "http://localhost:8543", "URL of the persistent context web service")
	rootCmd.PersistentFlags().BoolVar(&directMode, "direct", false, "Use direct mode (bypass HTTP)")

	viper.BindPFlag("web_url", rootCmd.PersistentFlags().Lookup("web-url"))
	viper.BindPFlag("direct_mode", rootCmd.PersistentFlags().Lookup("direct"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		configDir := home + "/.persistent-context"
		viper.AddConfigPath(configDir)
		viper.SetConfigName("cli")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}