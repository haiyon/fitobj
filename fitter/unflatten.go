package fitter

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Regular expression to match array indices in bracket notation
	// Matches patterns like "key[0]" and extracts the index
	bracketArrayRegex = regexp.MustCompile(`^(.*)\[(\d+)\]$`)
)

// UnflattenOptions configures the unflattening process
type UnflattenOptions struct {
	// Separator is the character used to separate flattened keys
	Separator string
	// DetectArrays automatically converts numeric indices to array elements
	DetectArrays bool
	// SupportBracketNotation enables parsing of array indices in bracket notation: key[0]
	SupportBracketNotation bool
	// BufferSize is the initial capacity of result maps to reduce allocations
	BufferSize int
}

// DefaultUnflattenOptions returns the default options for unflattening
func DefaultUnflattenOptions() UnflattenOptions {
	return UnflattenOptions{
		Separator:              ".",
		DetectArrays:           true,
		SupportBracketNotation: true,
		BufferSize:             16, // Default starting capacity for map
	}
}

// UnflattenMap converts a flattened map back into a nested structure.
// Example: {"user.name": "John"} becomes {"user": {"name": "John"}}
//
// Parameters:
// - obj: The flattened map to convert to a nested structure
//
// Returns:
// - A new map with nested structure
func UnflattenMap(obj map[string]any) map[string]any {
	return UnflattenMapWithOptions(obj, DefaultUnflattenOptions())
}

// UnflattenMapWithOptions converts a flattened map back into a nested structure with custom options.
//
// Parameters:
// - obj: The flattened map to convert to a nested structure
// - options: Configuration options for the unflattening process
//
// Returns:
// - A new map with nested structure
func UnflattenMapWithOptions(obj map[string]any, options UnflattenOptions) map[string]any {
	result := make(map[string]any, options.BufferSize)

	// If the object is empty, return empty map
	if len(obj) == 0 {
		return result
	}

	// First, analyze all keys to identify arrays
	arrayPaths := make(map[string][]int)
	if options.DetectArrays {
		for key := range obj {
			parts := splitKey(key, options)

			// Check each part for potential array indices
			for i := 0; i < len(parts); i++ {
				path := strings.Join(parts[:i], options.Separator)
				if idx, err := strconv.Atoi(parts[i]); err == nil {
					indices := arrayPaths[path]
					arrayPaths[path] = appendUniqueInt(indices, idx)
				}
			}
		}
	}

	// Process each key-value pair
	for key, value := range obj {
		parts := splitKey(key, options)

		if len(parts) == 0 {
			continue
		}

		// Start with root result map
		current := result

		// Process all parts except the last one to build the nested structure
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]

			// Check if this part should be treated as an array index
			if idx, err := strconv.Atoi(part); err == nil && options.DetectArrays {
				// This part is a numeric index, check if parent exists
				if i == 0 {
					// Special case: top-level array (rare)
					continue
				}

				// Check if parent already exists
				parentKey := parts[i-1]
				if parent, exists := current[parentKey]; exists {
					// Parent exists, check if it's already an array
					if arr, isArray := parent.([]any); isArray {
						// Ensure array capacity
						for len(arr) <= idx {
							arr = append(arr, nil)
						}

						// Update the array in the parent
						current[parentKey] = arr

						// If next part exists, ensure this index contains a map
						if i < len(parts)-2 {
							if arr[idx] == nil {
								arr[idx] = make(map[string]any, options.BufferSize)
							}

							// Move to the next level
							if mapVal, ok := arr[idx].(map[string]any); ok {
								current = mapVal
							} else {
								// Convert to map if needed
								newMap := make(map[string]any, options.BufferSize)
								arr[idx] = newMap
								current = newMap
							}
						}
					} else {
						// Parent exists but is not an array yet
						// Create a new array
						arr := make([]any, idx+1)

						// If parent is a map, merge into first element
						if mapVal, ok := parent.(map[string]any); ok && len(mapVal) > 0 {
							arr[0] = mapVal
						} else {
							// Otherwise just set first element to parent value
							arr[0] = parent
						}

						// Update parent with new array
						current[parentKey] = arr

						// If next part exists, ensure this index contains a map
						if i < len(parts)-2 {
							if arr[idx] == nil {
								arr[idx] = make(map[string]any, options.BufferSize)
							}

							// Move to the next level
							if mapVal, ok := arr[idx].(map[string]any); ok {
								current = mapVal
							} else {
								// Convert to map if needed
								newMap := make(map[string]any, options.BufferSize)
								arr[idx] = newMap
								current = newMap
							}
						}
					}
				} else {
					// Parent doesn't exist, create a new array
					arr := make([]any, idx+1)

					// Update parent with new array
					current[parentKey] = arr

					// If next part exists, ensure this index contains a map
					if i < len(parts)-2 {
						newMap := make(map[string]any, options.BufferSize)
						arr[idx] = newMap
						current = newMap
					}
				}

				// Skip next part as we've handled it
				i++
				continue
			}

			// Not an array index, standard object property
			if _, exists := current[part]; !exists {
				// Check if this part should be an array based on next part
				if i+1 < len(parts) {
					if _, isArray := arrayPaths[strings.Join(parts[:i+1], options.Separator)]; isArray {
						// Next level should be an array
						current[part] = make([]any, 0, 4) // Start with small capacity
					} else {
						// Regular object
						current[part] = make(map[string]any, options.BufferSize)
					}
				} else {
					// Last part, create empty map
					current[part] = make(map[string]any, options.BufferSize)
				}
			}

			// Check if the current part's value is an array
			if arr, isArray := current[part].([]any); isArray {
				// If array, we don't navigate into it unless next part is numeric
				if i+1 < len(parts) {
					if idx, err := strconv.Atoi(parts[i+1]); err == nil {
						// Next part is numeric, ensure capacity
						for len(arr) <= idx {
							arr = append(arr, nil)
						}

						// Update array in case it was extended
						current[part] = arr

						// If index doesn't have a map, create one
						if arr[idx] == nil {
							arr[idx] = make(map[string]any, options.BufferSize)
						}

						// Navigate to the map at this index
						if mapVal, ok := arr[idx].(map[string]any); ok {
							current = mapVal
						} else {
							// Convert to map if needed
							newMap := make(map[string]any, options.BufferSize)
							arr[idx] = newMap
							current = newMap
						}

						// Skip next part as we've handled it
						i++
					}
				}
				continue
			}

			// Navigate to the next level
			if next, ok := current[part].(map[string]any); ok {
				current = next
			}
		}

		// Now handle the last part to set the actual value
		lastPart := parts[len(parts)-1]

		// Check if last part is numeric (array index)
		if idx, err := strconv.Atoi(lastPart); err == nil && options.DetectArrays && len(parts) > 1 {
			// Get parent key
			parentKey := parts[len(parts)-2]

			if parent, exists := current[parentKey]; exists {
				// Parent exists, check if it's already an array
				if arr, isArray := parent.([]any); isArray {
					// Ensure array capacity
					for len(arr) <= idx {
						arr = append(arr, nil)
					}

					// Set the value at this index
					arr[idx] = value

					// Update the array in the parent
					current[parentKey] = arr
				} else {
					// Parent exists but is not an array yet
					// Create a new array
					arr := make([]any, idx+1)

					// If parent is a map, merge into first element
					if mapVal, ok := parent.(map[string]any); ok && len(mapVal) > 0 {
						arr[0] = mapVal
					} else {
						// Otherwise just set first element to parent value
						arr[0] = parent
					}

					// Set the value at this index
					arr[idx] = value

					// Update parent with new array
					current[parentKey] = arr
				}
			}
		} else {
			// Regular property, just set it
			current[lastPart] = value
		}
	}

	// Post-process arrays
	processArrays(result, arrayPaths, options.Separator)

	return result
}

// Helper function to split a key into parts, handling bracket notation if enabled
func splitKey(key string, options UnflattenOptions) []string {
	if !options.SupportBracketNotation {
		return strings.Split(key, options.Separator)
	}

	// Custom split that understands bracket notation
	var parts []string
	currentIndex := 0

	for currentIndex < len(key) {
		// Find separator
		sepIndex := strings.Index(key[currentIndex:], options.Separator)

		if sepIndex == -1 {
			// No more separators, process remaining part
			remaining := key[currentIndex:]

			// Check for bracket notation
			if match := bracketArrayRegex.FindStringSubmatch(remaining); match != nil {
				if match[1] != "" {
					parts = append(parts, match[1])
				}
				parts = append(parts, match[2])
			} else {
				parts = append(parts, remaining)
			}

			break
		}

		// Get current part
		sepIndex += currentIndex
		part := key[currentIndex:sepIndex]

		// Check for bracket notation
		if match := bracketArrayRegex.FindStringSubmatch(part); match != nil {
			if match[1] != "" {
				parts = append(parts, match[1])
			}
			parts = append(parts, match[2])
		} else {
			parts = append(parts, part)
		}

		// Move to next part
		currentIndex = sepIndex + len(options.Separator)
	}

	return parts
}

// Helper function to append an integer to a slice only if it's not already present
func appendUniqueInt(slice []int, value int) []int {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

// Helper function to process array structures after the initial unflattening
func processArrays(obj map[string]any, arrayPaths map[string][]int, separator string) {
	for key, value := range obj {
		switch v := value.(type) {
		case map[string]any:
			// Recursively process nested maps
			processArrays(v, arrayPaths, separator)

			// Check for numeric keys in this map, which might indicate it should be an array
			numericKeys := false
			for k := range v {
				if _, err := strconv.Atoi(k); err == nil {
					numericKeys = true
					break
				}
			}

			if numericKeys {
				// Find the highest index
				maxIndex := -1
				for k := range v {
					if idx, err := strconv.Atoi(k); err == nil && idx > maxIndex {
						maxIndex = idx
					}
				}

				if maxIndex >= 0 {
					// Create an array with appropriate size
					arr := make([]any, maxIndex+1)

					// Fill array with values from map
					for k, val := range v {
						if idx, err := strconv.Atoi(k); err == nil && idx <= maxIndex {
							arr[idx] = val
						}
					}

					// Replace map with array in parent
					obj[key] = arr
				}
			}

		case []any:
			// Process each element of the array
			for i, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					processArrays(itemMap, arrayPaths, separator)
					v[i] = itemMap
				}
			}
		}
	}
}
