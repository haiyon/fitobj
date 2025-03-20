package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadJSONFile reads a JSON file and unmarshals it into a map.
//
// Parameters:
// - filePath: Path to the JSON file
//
// Returns:
// - Unmarshalled map and error if any
func ReadJSONFile(filePath string) (map[string]any, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Unmarshal JSON
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}

// WriteJSONFile writes a map to a JSON file with indentation.
//
// Parameters:
// - filePath: Path to write the JSON file
// - data: Map to be written as JSON
//
// Returns:
// - Error if any
func WriteJSONFile(filePath string, data map[string]any) error {
	// Marshal with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist.
//
// Parameters:
// - dirPath: Directory path to ensure exists
//
// Returns:
// - Error if any
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	return nil
}

// ParseJSON parses a JSON string into a map.
//
// Parameters:
// - jsonStr: JSON string to parse
//
// Returns:
// - Parsed map and error if any
func ParseJSON(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return result, nil
}

// PrintJSON prints a map as JSON with indentation.
//
// Parameters:
// - data: Map to print as JSON
func PrintJSON(data map[string]any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("JSON error: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// PrintJSONFile prints the content of a JSON file with indentation.
//
// Parameters:
// - filePath: Path to the JSON file
func PrintJSONFile(filePath string) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return
	}

	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	PrintJSON(data)
}
