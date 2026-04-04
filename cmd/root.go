package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "wallpaper-cli",
	Short: "A resource-efficient CLI tool for downloading wallpapers",
	Long: `wallpaper-cli is a resource-efficient CLI tool for downloading high-quality
anime and general wallpapers from multiple sources with smart filtering,
deduplication, and organization.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/wallpaper-cli/config.json)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + "/.config/wallpaper-cli")
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
