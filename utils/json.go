package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadJSONFile reads a JSON file and unmarshals it into a map
func ReadJSONFile(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	if len(data) == 0 {
		return make(map[string]any), nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}

// WriteJSONFile writes a map to a JSON file with indentation
func WriteJSONFile(filePath string, data map[string]any) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := EnsureDirectoryExists(dir); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	return nil
}

// ParseJSON parses a JSON string into a map
func ParseJSON(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return result, nil
}

// PrintJSON prints a map as JSON with indentation
func PrintJSON(data map[string]any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("JSON error: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// PrintJSONFile prints the content of a JSON file with indentation
func PrintJSONFile(filePath string) {
	data, err := ReadJSONFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return
	}

	PrintJSON(data)
}

// IsJSONFile checks if a file is a JSON file based on its extension
func IsJSONFile(filename string) bool {
	return filepath.Ext(filename) == ".json"
}
