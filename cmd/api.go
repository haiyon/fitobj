package cmd

import (
	"fmt"

	"github.com/haiyon/fitobj/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start API server mode",
	Long: `Start a RESTful API server for JSON processing.

The server provides endpoints for flattening and unflattening JSON data.

Example:
  fitobj api --port=8080
  fitobj api --port=3000 --separator="__"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port := viper.GetString("api.port")

		fmt.Println("Starting fitobj in API mode...")

		options := api.Options{
			Port:          port,
			FlattenOpts:   buildFlattenOptions(),
			UnflattenOpts: buildUnflattenOptions(),
		}

		return api.StartServerWithOptions(options)
	},
}

func init() {
	apiCmd.Flags().String("port", "8080", "port for API server")
	viper.BindPFlag("api.port", apiCmd.Flags().Lookup("port"))

	rootCmd.AddCommand(apiCmd)
}
