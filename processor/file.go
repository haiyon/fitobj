package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/utils"
)

// Options configures the file processing behavior
type Options struct {
	// Workers is the number of worker goroutines for parallel processing
	Workers int
	// FlattenOpts contains options for the flatten operation
	FlattenOpts fitter.FlattenOptions
	// UnflattenOpts contains options for the unflatten operation
	UnflattenOpts fitter.UnflattenOptions
}

// DefaultOptions returns the default options for processing
func DefaultOptions() Options {
	return Options{
		Workers:       4, // Default to 4 workers
		FlattenOpts:   fitter.DefaultFlattenOptions(),
		UnflattenOpts: fitter.DefaultUnflattenOptions(),
	}
}

// ProcessFile processes a single JSON file by either flattening or unflattening it.
//
// Parameters:
// - inputPath: Path to the input JSON file
// - outputPath: Path to write the output JSON file
// - unflatten: Whether to unflatten (true) or flatten (false) the JSON
//
// Returns:
// - Error if any
func ProcessFile(inputPath, outputPath string, unflatten bool) error {
	return ProcessFileWithOptions(inputPath, outputPath, unflatten, DefaultOptions())
}

// ProcessFileWithOptions processes a single JSON file with custom options.
//
// Parameters:
// - inputPath: Path to the input JSON file
// - outputPath: Path to write the output JSON file
// - unflatten: Whether to unflatten (true) or flatten (false) the JSON
// - options: Configuration options for processing
//
// Returns:
// - Error if any
func ProcessFileWithOptions(inputPath, outputPath string, unflatten bool, options Options) error {
	// Read and parse the input JSON file
	jsonData, err := utils.ReadJSONFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file %s: %v", inputPath, err)
	}

	var processedData map[string]any
	if unflatten {
		// Unflatten the JSON structure
		processedData = fitter.UnflattenMapWithOptions(jsonData, options.UnflattenOpts)
	} else {
		// Flatten the JSON structure
		// Make sure to pass an empty string as the prefix for root level flattening
		processedData = fitter.FlattenMapWithOptions(jsonData, "", options.FlattenOpts)
	}

	// Write the processed data to the output file
	if err := utils.WriteJSONFile(outputPath, processedData); err != nil {
		return fmt.Errorf("failed to write output file %s: %v", outputPath, err)
	}

	return nil
}

// ProcessDirectory processes all JSON files in a directory.
//
// Parameters:
// - inputDir: Path to the directory containing input JSON files
// - outputDir: Path to the directory to write output JSON files
// - unflatten: Whether to unflatten (true) or flatten (false) the JSON files
//
// Returns:
// - Error if any
func ProcessDirectory(inputDir, outputDir string, unflatten bool) error {
	return ProcessDirectoryWithOptions(inputDir, outputDir, unflatten, DefaultOptions())
}

// ProcessDirectoryWithOptions processes all JSON files in a directory with custom options.
//
// Parameters:
// - inputDir: Path to the directory containing input JSON files
// - outputDir: Path to the directory to write output JSON files
// - unflatten: Whether to unflatten (true) or flatten (false) the JSON files
// - options: Configuration options for processing
//
// Returns:
// - Error if any
func ProcessDirectoryWithOptions(inputDir, outputDir string, unflatten bool, options Options) error {
	// Validate input directory
	inputInfo, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("input directory error: %v", err)
	}
	if !inputInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", inputDir)
	}

	// Ensure output directory exists
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Read directory contents
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	// Filter for JSON files
	var jsonFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	if len(jsonFiles) == 0 {
		fmt.Printf("Warning: No JSON files found in '%s'\n", inputDir)
		return nil
	}

	// Set up concurrency
	numWorkers := options.Workers
	if numWorkers <= 0 {
		numWorkers = 1
	}

	// Create a channel to distribute work
	filesChan := make(chan string, len(jsonFiles))

	// Create a channel to collect results
	resultsChan := make(chan struct {
		filename string
		err      error
	}, len(jsonFiles))

	// Create a wait group to wait for all workers
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range filesChan {
				inputPath := filepath.Join(inputDir, file)
				outputPath := filepath.Join(outputDir, file)

				err := ProcessFileWithOptions(inputPath, outputPath, unflatten, options)

				resultsChan <- struct {
					filename string
					err      error
				}{file, err}
			}
		}()
	}

	// Send files to the workers
	for _, file := range jsonFiles {
		filesChan <- file
	}
	close(filesChan)

	// Create a goroutine to close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Process results as they come in
	successCount := 0
	errorCount := 0

	for result := range resultsChan {
		if result.err != nil {
			fmt.Printf("Error processing file '%s': %v\n", result.filename, result.err)
			errorCount++
		} else {
			fmt.Printf("Processed: %s\n", result.filename)
			successCount++
		}
	}

	fmt.Printf("Processing completed. Processed %d files (%d successful, %d failed)\n",
		len(jsonFiles), successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d files failed to process", errorCount)
	}

	return nil
}
