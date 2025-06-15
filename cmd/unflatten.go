package cmd

import (
	"fmt"

	"github.com/haiyon/fitobj/processor"
	"github.com/spf13/cobra"
)

var unflattenCmd = &cobra.Command{
	Use:   "unflatten [input-dir] [output-dir]",
	Short: "Unflatten JSON objects back to nested structure",
	Long: `Unflatten converts flat key-value pairs back into nested JSON objects.

Example:
  fitobj unflatten ./flattened ./nested
  fitobj unflatten ./flat ./nested --separator="__"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := args[0]
		outputDir := args[1]

		fmt.Printf("Unflattening JSON files from %s to %s\n", inputDir, outputDir)
		fmt.Printf("Using separator: '%s', array format: '%s', workers: %d\n",
			getSeparator(), getArrayFormat(), getWorkers())

		options := buildProcessorOptions()
		return processor.ProcessDirectoryWithOptions(inputDir, outputDir, true, options)
	},
}

func init() {
	rootCmd.AddCommand(unflattenCmd)
}
