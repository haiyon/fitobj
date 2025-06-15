package cmd

import (
	"fmt"

	"github.com/haiyon/fitobj/processor"
	"github.com/spf13/cobra"
)

var flattenCmd = &cobra.Command{
	Use:   "flatten [input-dir] [output-dir]",
	Short: "Flatten nested JSON objects",
	Long: `Flatten converts nested JSON objects into flat key-value pairs.

Example:
  fitobj flatten ./nested ./flattened
  fitobj flatten ./data ./output --separator="__" --array-format=bracket`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := args[0]
		outputDir := args[1]

		fmt.Printf("Flattening JSON files from %s to %s\n", inputDir, outputDir)
		fmt.Printf("Using separator: '%s', array format: '%s', workers: %d\n",
			getSeparator(), getArrayFormat(), getWorkers())

		options := buildProcessorOptions()
		return processor.ProcessDirectoryWithOptions(inputDir, outputDir, false, options)
	},
}

func init() {
	rootCmd.AddCommand(flattenCmd)
}
