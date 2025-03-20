package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/haiyon/fitobj/api"
	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/processor"
)

func main() {
	// Define command line flags
	inputDir := flag.String("input", "", "Path to input directory containing JSON files")
	outputDir := flag.String("output", "", "Path to output directory for processed JSON files")
	unflatten := flag.Bool("unflatten", false, "Unflatten the JSON structure (default: flatten)")
	apiMode := flag.Bool("api", false, "Start in API server mode")
	port := flag.String("port", "8080", "Port for API server")

	// Custom separator option
	separator := flag.String("separator", ".", "Separator character for flattened keys")

	// Array formatting option
	arrayFormat := flag.String("array-format", "index", "Array index format: 'index' (items.0) or 'bracket' (items[0])")

	// Performance options
	workers := flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines for parallel processing")
	bufferSize := flag.Int("buffer", 16, "Initial buffer size for maps to reduce allocations")
	cpuProfile := flag.String("cpuprofile", "", "Write CPU profile to file")
	memProfile := flag.String("memprofile", "", "Write memory profile to file")

	// Parse command line flags
	flag.Parse()

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Printf("Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
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
		fmt.Println("Usage: fitobj -input=<input-dir> -output=<output-dir> [-unflatten] [-separator=.] [-array-format=index|bracket]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Process files in the directory
	operation := "Flattening"
	if *unflatten {
		operation = "Unflattening"
	}

	fmt.Printf("%s JSON files from %s to %s\n", operation, *inputDir, *outputDir)
	fmt.Printf("Using separator: '%s', array format: '%s', with %d workers\n",
		*separator, *arrayFormat, *workers)

	if err := processor.ProcessDirectoryWithOptions(*inputDir, *outputDir, *unflatten, processorOptions); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fmt.Printf("Could not create memory profile: %v\n", err)
		}
		defer f.Close()
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("Could not write memory profile: %v\n", err)
		}
	}
}
