package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/haiyon/fitobj/fitter"
)

// The pattern to match t('key') or t("key") function calls in source files
var tPattern = regexp.MustCompile(`\bt\(\s*['"]([^'"]+?)['"]`)

// ExtractKeysFromFile extracts all t() function call keys from a single file
func ExtractKeysFromFile(filePath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return keys, nil // Ignore read errors (e.g., binary files)
	}

	// Find all matches in the file content
	matches := tPattern.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			keys[string(match[1])] = true
		}
	}

	return keys, nil
}

// ExtractKeysFromDir recursively extracts all t() function call keys from a directory
func ExtractKeysFromDir(rootDir string) (map[string]bool, error) {
	keys := make(map[string]bool)

	// Walk through all files in the directory recursively
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Extract keys from the file
			fileKeys, err := ExtractKeysFromFile(path)
			if err != nil {
				return err // Propagate the error
			}

			// Add the extracted keys to the result
			for key := range fileKeys {
				keys[key] = true
			}
		}

		return nil
	})

	return keys, err
}

// ExtractKeysFromJSON extracts all keys from a JSON file using flattening
func ExtractKeysFromJSON(filePath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	// Read and parse JSON file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return keys, fmt.Errorf("failed to read JSON file: %v", err)
	}

	var jsonObj map[string]any
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return keys, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Flatten the JSON object to get all keys
	options := fitter.DefaultFlattenOptions()
	flattenedObj := fitter.FlattenMapWithOptions(jsonObj, "", options)

	// Extract keys from the flattened object
	for key := range flattenedObj {
		keys[key] = true
	}

	return keys, nil
}

// ExtractKeysFromJSONDir extracts all keys from JSON files in a directory
func ExtractKeysFromJSONDir(jsonPath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	// Check if the path is a file or directory
	fileInfo, err := os.Stat(jsonPath)
	if err != nil {
		return keys, fmt.Errorf("failed to stat path: %v", err)
	}

	if fileInfo.IsDir() {
		// Process all JSON files in the directory
		entries, err := os.ReadDir(jsonPath)
		if err != nil {
			return keys, fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				fullPath := filepath.Join(jsonPath, entry.Name())
				jsonKeys, err := ExtractKeysFromJSON(fullPath)
				if err != nil {
					return keys, err
				}

				// Add the extracted keys to the result
				for key := range jsonKeys {
					keys[key] = true
				}
			}
		}
	} else {
		// Process a single JSON file
		jsonKeys, err := ExtractKeysFromJSON(jsonPath)
		if err != nil {
			return keys, err
		}

		// Add the extracted keys to the result
		for key := range jsonKeys {
			keys[key] = true
		}
	}

	return keys, nil
}

// CompareKeys compares source keys with JSON keys to find missing and unused keys
func CompareKeys(sourceKeys, jsonKeys map[string]bool) ([]string, []string) {
	var missingInJSON, unusedInSource []string

	// Find keys missing in JSON
	for key := range sourceKeys {
		if !jsonKeys[key] {
			missingInJSON = append(missingInJSON, key)
		}
	}

	// Find keys unused in source
	for key := range jsonKeys {
		if !sourceKeys[key] {
			unusedInSource = append(unusedInSource, key)
		}
	}

	return missingInJSON, unusedInSource
}
