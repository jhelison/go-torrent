package cmd

import (
	"fmt"
	"os"

	"github.com/jhelison/go-torrent/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile  string
	logLevel string

	rootCmd = &cobra.Command{
		Use:   "go-torrent",
		Short: "A small application to download torrent from torrent files",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	initConfig()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-torrent.toml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "update the log level (trace|info|warn|err|disabled)")

	// Additional commands
	rootCmd.AddCommand(DownloadCmd())
}

// initConfig initiates all the configurations used in go-torrent
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".go-torrent".
		viper.AddConfigPath(fmt.Sprintf("%s/.go-torrent", home))
		viper.SetConfigType("toml")
		viper.SetConfigName("config")

		// Save all the configs
		buildConfigs(home)

		// Create a new config if not found
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Create the file if it doesn't exists
				if err := os.MkdirAll(fmt.Sprintf("%s/.go-torrent", home), os.ModePerm); err != nil {
					fmt.Printf("Error creating config directory: %s\n", err)
					os.Exit(1)
				}

				// Write the config to a file
				if err := viper.SafeWriteConfig(); err != nil {
					fmt.Printf("Error creating default config file: %s\n", err)
				}
			}
		}
	}

	// Read the config
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	viper.AutomaticEnv()

	// Set the log level
	logger.SetLogLevel(logLevel)
}

// buildConfigs builds all the configs using Viper
func buildConfigs(home string) {
	viper.SetDefault("log_level", "debug")

	// Download config
	viper.SetDefault("download.deadline", "30s")
	viper.SetDefault("download.max_backlog", 10)
	viper.SetDefault("download.block_size", 16384)
	viper.SetDefault("download.output_path", fmt.Sprintf("%s/Downloads", home))

	// Peers config
	viper.SetDefault("peers.max_retries", 10)
	viper.SetDefault("peers.timeout", "5s")
}
