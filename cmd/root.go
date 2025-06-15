package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "fitobj",
	Short: "A lightweight tool for flattening and unflattening JSON objects",
	Long: `fitobj is a CLI tool for JSON manipulation with features including:
- Flatten nested JSON objects with customizable separators
- Unflatten objects back to nested structures
- Batch processing with parallel execution
- i18n key management and cleanup
- RESTful API server mode`,
	Version: version,
}

func Execute() {
	// Disable completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fitobj.yaml)")
	rootCmd.PersistentFlags().String("separator", ".", "separator character for flattened keys")
	rootCmd.PersistentFlags().String("array-format", "index", "array format: 'index' or 'bracket'")
	rootCmd.PersistentFlags().Int("workers", runtime.NumCPU(), "number of workers for parallel processing")
	rootCmd.PersistentFlags().Int("buffer", 16, "initial buffer size for maps")

	// Bind flags to viper
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".fitobj")
		}
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
