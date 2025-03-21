package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/haiyon/fitobj/api"
	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/processor"
)

var (
	// Version information
	version = "dev"
	// Define command line flags
	versionFlag = flag.Bool("version", false, "Display version information")
	// Paths
	inputDir  = flag.String("input", "", "Path to input directory containing JSON files")
	outputDir = flag.String("output", "", "Path to output directory for processed JSON files")
	// Operation
	reverse = flag.Bool("reverse", false, "Reverse the transformation (flatten->nested, nested->flatten)")
	// API mode
	apiMode = flag.Bool("api", false, "Start in API server mode")
	port    = flag.String("port", "8080", "Port for API server")
	// Options
	separator   = flag.String("separator", ".", "Separator character for flattened keys")
	arrayFormat = flag.String("array-format", "index", "Array index format: 'index' (items.0) or 'bracket' (items[0])")
	// Performance
	workers    = flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines for parallel processing")
	bufferSize = flag.Int("buffer", 16, "Initial buffer size for maps to reduce allocations")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Check if version flag is set
	if *versionFlag {
		fmt.Printf("fitobj version is %s\n", version)
		return
	}

	// Processing options
	flattenOpts := fitter.DefaultFlattenOptions()
	flattenOpts.Separator = *separator
	flattenOpts.ArrayFormatting = *arrayFormat
	flattenOpts.BufferSize = *bufferSize

	unflattenOpts := fitter.DefaultUnflattenOptions()
	unflattenOpts.Separator = *separator
	unflattenOpts.SupportBracketNotation = *arrayFormat == "bracket"
	unflattenOpts.BufferSize = *bufferSize

	processorOptions := processor.Options{
		Workers:       *workers,
		FlattenOpts:   flattenOpts,
		UnflattenOpts: unflattenOpts,
	}

	// API server mode
	if *apiMode {
		fmt.Println("Starting fitobj in API mode...")

		apiOptions := api.Options{
			Port:          *port,
			FlattenOpts:   processorOptions.FlattenOpts,
			UnflattenOpts: processorOptions.UnflattenOpts,
		}

		if err := api.StartServerWithOptions(apiOptions); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// File processing mode
	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Error: Input and output directories must be specified")
		fmt.Println("Usage: fitobj -input=<input-dir> -output=<output-dir> [-reverse] [-separator=.] [-array-format=index|bracket]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Process files in the directory
	operation := "Flattening"
	if *reverse {
		operation = "Reversing (unflattening)"
	}

	fmt.Printf("%s JSON files from %s to %s\n", operation, *inputDir, *outputDir)
	fmt.Printf("Using separator: '%s', array format: '%s', with %d workers\n",
		*separator, *arrayFormat, *workers)

	if err := processor.ProcessDirectoryWithOptions(*inputDir, *outputDir, *reverse, processorOptions); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
