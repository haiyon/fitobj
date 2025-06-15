package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/haiyon/fitobj/fitter"
)

// Pattern to match t('key') or t("key") function calls in source files
var tPattern = regexp.MustCompile(`\bt\(\s*['"]([^'"]+?)['"]`)

// ExtractKeysFromFile extracts all t() function call keys from a single file
func ExtractKeysFromFile(filePath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return keys, nil // Ignore read errors (e.g., binary files)
	}

	matches := tPattern.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			key := strings.TrimSpace(string(match[1]))
			if key != "" {
				keys[key] = true
			}
		}
	}

	return keys, nil
}

// ExtractKeysFromDir recursively extracts all t() function call keys from a directory
func ExtractKeysFromDir(rootDir string) (map[string]bool, error) {
	keys := make(map[string]bool)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process text-like files
		if !d.IsDir() && isTextFile(path) {
			fileKeys, err := ExtractKeysFromFile(path)
			if err != nil {
				return err
			}

			for key := range fileKeys {
				keys[key] = true
			}
		}

		return nil
	})

	return keys, err
}

// isTextFile checks if a file is likely a text file based on its extension
func isTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	textExts := []string{
		".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte",
		".py", ".rb", ".php", ".java", ".kt", ".go",
		".html", ".htm", ".xml", ".css", ".scss", ".sass",
		".md", ".txt", ".json", ".yaml", ".yml",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	return false
}

// ExtractKeysFromJSON extracts all keys from a JSON file using flattening
func ExtractKeysFromJSON(filePath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return keys, fmt.Errorf("failed to read JSON file: %v", err)
	}

	if len(jsonData) == 0 {
		return keys, nil
	}

	var jsonObj map[string]any
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return keys, fmt.Errorf("failed to parse JSON: %v", err)
	}

	options := fitter.DefaultFlattenOptions()
	flattenedObj := fitter.FlattenMapWithOptions(jsonObj, "", options)

	for key := range flattenedObj {
		keys[key] = true
	}

	return keys, nil
}

// ExtractKeysFromJSONDir extracts all keys from JSON files in a directory
func ExtractKeysFromJSONDir(jsonPath string) (map[string]bool, error) {
	keys := make(map[string]bool)

	fileInfo, err := os.Stat(jsonPath)
	if err != nil {
		return keys, fmt.Errorf("failed to stat path: %v", err)
	}

	if fileInfo.IsDir() {
		entries, err := os.ReadDir(jsonPath)
		if err != nil {
			return keys, fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				fullPath := filepath.Join(jsonPath, entry.Name())
				jsonKeys, err := ExtractKeysFromJSON(fullPath)
				if err != nil {
					fmt.Printf("Warning: Failed to process %s: %v\n", fullPath, err)
					continue
				}

				for key := range jsonKeys {
					keys[key] = true
				}
			}
		}
	} else {
		jsonKeys, err := ExtractKeysFromJSON(jsonPath)
		if err != nil {
			return keys, err
		}

		for key := range jsonKeys {
			keys[key] = true
		}
	}

	return keys, nil
}

// CompareKeys compares source keys with JSON keys to find missing and unused keys
func CompareKeys(sourceKeys, jsonKeys map[string]bool) ([]string, []string) {
	var missingInJSON, unusedInSource []string

	for key := range sourceKeys {
		if !jsonKeys[key] {
			missingInJSON = append(missingInJSON, key)
		}
	}

	for key := range jsonKeys {
		if !sourceKeys[key] {
			unusedInSource = append(unusedInSource, key)
		}
	}

	// Sort for consistent output
	sort.Strings(missingInJSON)
	sort.Strings(unusedInSource)

	return missingInJSON, unusedInSource
}

// RemoveKeysFromPath removes specified keys from a nested JSON structure
func RemoveKeysFromPath(value map[string]any, keyPath string, separator string) bool {
	parts := splitKeyPath(keyPath, separator)
	if len(parts) == 0 {
		return false
	}

	// Handle single part (top-level key)
	if len(parts) == 1 {
		if _, exists := value[parts[0]]; exists {
			delete(value, parts[0])
			return true
		}
		return false
	}

	// Navigate to the parent of the target key
	current := value
	parents := []map[string]any{value}
	var parentKeys []string

	for _, part := range parts[:len(parts)-1] {
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]any); ok {
				current = nextMap
				parents = append(parents, current)
				parentKeys = append(parentKeys, part)
			} else {
				return false
			}
		} else {
			return false
		}
	}

	// Remove the target key
	targetKey := parts[len(parts)-1]
	if _, exists := current[targetKey]; exists {
		delete(current, targetKey)

		// Clean up empty parent objects from bottom to top
		for i := len(parents) - 1; i > 0; i-- {
			if len(parents[i]) == 0 {
				parentKey := parentKeys[i-1]
				delete(parents[i-1], parentKey)
			} else {
				break // Stop if parent is not empty
			}
		}

		return true
	}

	return false
}

// splitKeyPath splits a key path by separator
func splitKeyPath(keyPath, separator string) []string {
	if keyPath == "" {
		return []string{}
	}

	if separator == "." {
		return strings.Split(keyPath, separator)
	}

	// Handle custom separators
	var parts []string
	current := ""
	sepLen := len(separator)

	for i := 0; i < len(keyPath); {
		if i+sepLen <= len(keyPath) && keyPath[i:i+sepLen] == separator {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			i += sepLen
		} else {
			current += string(keyPath[i])
			i++
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

	fileInfo, err := os.Stat(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to stat path: %v", err)
	}

	if fileInfo.IsDir() {
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
		if err := cleanupJSONFile(jsonPath, unusedKeys, separator); err != nil {
			return fmt.Errorf("failed to cleanup file %s: %v", jsonPath, err)
		}
	}

	return nil
}

// cleanupJSONFile removes unused keys from a single JSON file
func cleanupJSONFile(filePath string, unusedKeys []string, separator string) error {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	if len(jsonData) == 0 {
		return nil // Skip empty files
	}

	var jsonObj map[string]any
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	originalSize := len(jsonData)
	removedCount := 0

	for _, key := range unusedKeys {
		if RemoveKeysFromPath(jsonObj, key, separator) {
			removedCount++
		}
	}

	if removedCount > 0 {
		updatedData, err := json.MarshalIndent(jsonObj, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}

		if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
			return fmt.Errorf("failed to write JSON file: %v", err)
		}

		fmt.Printf("✅ Removed %d unused keys from %s (size: %d -> %d bytes)\n",
			removedCount, filePath, originalSize, len(updatedData))
	}

	return nil
}
