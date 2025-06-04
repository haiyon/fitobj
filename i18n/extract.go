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

// RemoveKeysFromPath removes specified keys from a nested JSON structure in the given path
func RemoveKeysFromPath(value map[string]any, keyPath string, separator string) bool {
	parts := splitKeyPath(keyPath, separator)
	if len(parts) == 0 {
		return false
	}

	// Navigate to the parent of the target key
	current := value
	for _, part := range parts[:len(parts)-1] {
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]any); ok {
				current = nextMap
			} else {
				// Path doesn't exist or isn't navigable
				return false
			}
		} else {
			// Path doesn't exist
			return false
		}
	}

	// Remove the target key
	targetKey := parts[len(parts)-1]
	if _, exists := current[targetKey]; exists {
		delete(current, targetKey)
		return true
	}

	return false
}

// splitKeyPath splits a key path by separator (e.g., "hello.world" -> ["hello", "world"])
func splitKeyPath(keyPath, separator string) []string {
	if keyPath == "" {
		return []string{}
	}

	var parts []string
	current := ""

	for i := 0; i < len(keyPath); i++ {
		if i+len(separator) <= len(keyPath) && keyPath[i:i+len(separator)] == separator {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			i += len(separator) - 1 // Skip separator
		} else {
			current += string(keyPath[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// CleanupUnusedKeys removes unused keys from JSON files in the specified path
func CleanupUnusedKeys(jsonPath string, unusedKeys []string, separator string) error {
	if len(unusedKeys) == 0 {
		return nil
	}

	// Check if the path is a file or directory
	fileInfo, err := os.Stat(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to stat path: %v", err)
	}

	if fileInfo.IsDir() {
		// Process all JSON files in the directory
		entries, err := os.ReadDir(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				fullPath := filepath.Join(jsonPath, entry.Name())
				if err := cleanupJSONFile(fullPath, unusedKeys, separator); err != nil {
					return fmt.Errorf("failed to cleanup file %s: %v", fullPath, err)
				}
			}
		}
	} else {
		// Process a single JSON file
		if err := cleanupJSONFile(jsonPath, unusedKeys, separator); err != nil {
			return fmt.Errorf("failed to cleanup file %s: %v", jsonPath, err)
		}
	}

	return nil
}

// cleanupJSONFile removes unused keys from a single JSON file
func cleanupJSONFile(filePath string, unusedKeys []string, separator string) error {
	// Read the JSON file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	var jsonObj map[string]any
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Remove unused keys
	removedCount := 0
	for _, key := range unusedKeys {
		if RemoveKeysFromPath(jsonObj, key, separator) {
			removedCount++
		}
	}

	// Only write back if changes were made
	if removedCount > 0 {
		// Marshal back to JSON with proper formatting
		updatedData, err := json.MarshalIndent(jsonObj, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}

		// Write back to file
		if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
			return fmt.Errorf("failed to write JSON file: %v", err)
		}

		fmt.Printf("✅ Removed %d unused keys from %s\n", removedCount, filePath)
	}

	return nil
}
