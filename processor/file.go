package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/utils"
)

// Options configures the file processing behavior
type Options struct {
	Workers       int
	FlattenOpts   fitter.FlattenOptions
	UnflattenOpts fitter.UnflattenOptions
}

// DefaultOptions returns the default options for processing
func DefaultOptions() Options {
	return Options{
		Workers:       4,
		FlattenOpts:   fitter.DefaultFlattenOptions(),
		UnflattenOpts: fitter.DefaultUnflattenOptions(),
	}
}

// ProcessFile processes a single JSON file
func ProcessFile(inputPath, outputPath string, unflatten bool) error {
	return ProcessFileWithOptions(inputPath, outputPath, unflatten, DefaultOptions())
}

// ProcessFileWithOptions processes a single JSON file with custom options
func ProcessFileWithOptions(inputPath, outputPath string, unflatten bool, options Options) error {
	// Read and parse the input JSON file
	jsonData, err := utils.ReadJSONFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file %s: %v", inputPath, err)
	}

	var processedData map[string]any
	if unflatten {
		processedData = fitter.UnflattenMapWithOptions(jsonData, options.UnflattenOpts)
	} else {
		processedData = fitter.FlattenMapWithOptions(jsonData, "", options.FlattenOpts)
	}

	// Write the processed data to the output file
	if err := utils.WriteJSONFile(outputPath, processedData); err != nil {
		return fmt.Errorf("failed to write output file %s: %v", outputPath, err)
	}

	return nil
}

// ProcessDirectory processes all JSON files in a directory
func ProcessDirectory(inputDir, outputDir string, unflatten bool) error {
	return ProcessDirectoryWithOptions(inputDir, outputDir, unflatten, DefaultOptions())
}

// ProcessDirectoryWithOptions processes all JSON files in a directory with custom options
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

	// Create channels
	filesChan := make(chan string, len(jsonFiles))
	resultsChan := make(chan ProcessResult, len(jsonFiles))

	// Counters for progress tracking
	var processed, failed int64

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range filesChan {
				inputPath := filepath.Join(inputDir, file)
				outputPath := filepath.Join(outputDir, file)

				err := ProcessFileWithOptions(inputPath, outputPath, unflatten, options)

				result := ProcessResult{Filename: file, Error: err}
				resultsChan <- result

				if err != nil {
					atomic.AddInt64(&failed, 1)
				} else {
					atomic.AddInt64(&processed, 1)
				}
			}
		}()
	}

	// Send files to workers
	for _, file := range jsonFiles {
		filesChan <- file
	}
	close(filesChan)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Process results
	for result := range resultsChan {
		if result.Error != nil {
			fmt.Printf("Error processing file '%s': %v\n", result.Filename, result.Error)
		} else {
			fmt.Printf("Processed: %s\n", result.Filename)
		}
	}

	successCount := atomic.LoadInt64(&processed)
	errorCount := atomic.LoadInt64(&failed)

	fmt.Printf("Processing completed. Processed %d files (%d successful, %d failed)\n",
		len(jsonFiles), successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d files failed to process", errorCount)
	}

	return nil
}

// ProcessResult represents the result of processing a single file
type ProcessResult struct {
	Filename string
	Error    error
}
