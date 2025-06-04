package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/haiyon/fitobj/api"
	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/i18n"
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
	// i18n feature
	i18nMode      = flag.Bool("i18n", false, "Extract and compare i18n keys")
	sourceDir     = flag.String("source-dir", "", "Source directory to extract t() function calls from")
	jsonPath      = flag.String("json-path", "", "Path to JSON file or directory containing translation keys")
	cleanupUnused = flag.Bool("cleanup", false, "Automatically remove unused keys from JSON files (use with -i18n)")
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

	// i18n mode
	if *i18nMode {
		if *sourceDir == "" || *jsonPath == "" {
			fmt.Println("Usage: fitobj -i18n -source-dir=<source-dir> -json-path=<json-path> [-cleanup]")
			flag.PrintDefaults()
			os.Exit(1)
		}

		fmt.Printf("Extracting and comparing i18n keys...\n")
		fmt.Printf("Source directory: %s\n", *sourceDir)
		fmt.Printf("JSON path: %s\n", *jsonPath)
		if *cleanupUnused {
			fmt.Printf("Cleanup mode: Enabled (unused keys will be removed)\n")
		}

		// Extract keys from source files
		sourceKeys, err := i18n.ExtractKeysFromDir(*sourceDir)
		if err != nil {
			fmt.Printf("Error extracting keys from source: %v\n", err)
			os.Exit(1)
		}

		// Extract keys from JSON files
		jsonKeys, err := i18n.ExtractKeysFromJSONDir(*jsonPath)
		if err != nil {
			fmt.Printf("Error extracting keys from JSON: %v\n", err)
			os.Exit(1)
		}

		// Compare keys
		missingInJSON, unusedInSource := i18n.CompareKeys(sourceKeys, jsonKeys)

		// Print results
		fmt.Printf("\n🔍 Total keys in source: %d\n", len(sourceKeys))
		fmt.Printf("📚 Total keys in JSON: %d\n", len(jsonKeys))

		fmt.Printf("\n❌ Missing in JSON (%d):\n", len(missingInJSON))
		for _, key := range missingInJSON {
			fmt.Println(key)
		}

		fmt.Printf("\n🟡 Unused in Source (%d):\n", len(unusedInSource))
		for _, key := range unusedInSource {
			fmt.Println(key)
		}

		// Auto-cleanup unused keys if requested
		if *cleanupUnused && len(unusedInSource) > 0 {
			fmt.Printf("\n🧹 Cleaning up unused keys...\n")
			if err := i18n.CleanupUnusedKeys(*jsonPath, unusedInSource, *separator); err != nil {
				fmt.Printf("❌ Error during cleanup: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Cleanup completed successfully!\n")
		} else if *cleanupUnused && len(unusedInSource) == 0 {
			fmt.Printf("\n✅ No unused keys to cleanup!\n")
		}

		return
	}

	// Check if cleanup flag is used without i18n mode
	if *cleanupUnused && !*i18nMode {
		fmt.Println("Error: -cleanup flag can only be used with -i18n mode")
		os.Exit(1)
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
