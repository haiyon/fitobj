package fitter

import (
	"regexp"
	"strconv"
	"strings"
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
	result := make(map[string]any)

	// Preprocess keys to convert bracket notation if needed
	processedObj := make(map[string]any, len(obj))
	for k, v := range obj {
		if options.SupportBracketNotation {
			// Convert bracket notation to dot notation with actual separator
			k = convertBracketToDot(k, options.Separator)
		}
		processedObj[k] = v
	}

	// Process each key-value pair
	for key, value := range processedObj {
		// Split the key by separator
		parts := strings.Split(key, options.Separator)

		// Set value in the nested structure
		assignToNested(result, parts, value, options)
	}

	// Final pass to convert numeric maps to arrays
	return convertNumericMapsToArrays(result)
}

// convertBracketToDot converts "user[0].name" to "user.0.name"
func convertBracketToDot(key, separator string) string {
	// Replace bracket notation with dot notation
	re := regexp.MustCompile(`\[([0-9]+)\]`)
	return re.ReplaceAllString(key, separator+"$1")
}

// assignToNested sets a value at a path in a nested structure
func assignToNested(obj map[string]any, parts []string, value any, options UnflattenOptions) {
	if len(parts) == 0 {
		return
	}

	// Get the current part (first part of the path)
	part := parts[0]

	if len(parts) == 1 {
		// Last part, set the value directly
		obj[part] = value
		return
	}

	// Check if next part is numeric (potentially an array index)
	nextIsNumeric := false
	nextIndex := -1

	if options.DetectArrays {
		if idx, err := strconv.Atoi(parts[1]); err == nil {
			nextIsNumeric = true
			nextIndex = idx
		}
	}

	// Create or update the nested structure
	if nextIsNumeric {
		// Next part is an array index
		var arr []any

		if existing, ok := obj[part]; ok {
			// Try to use existing array
			if existingArr, ok := existing.([]any); ok {
				arr = existingArr
			} else {
				// Convert to array
				arr = make([]any, nextIndex+1)
			}
		} else {
			// Create new array
			arr = make([]any, nextIndex+1)
		}

		// Ensure capacity
		for len(arr) <= nextIndex {
			arr = append(arr, nil)
		}

		// Get or create map at the index
		var nextObj map[string]any

		if arr[nextIndex] == nil {
			nextObj = make(map[string]any)
			arr[nextIndex] = nextObj
		} else if mapVal, ok := arr[nextIndex].(map[string]any); ok {
			nextObj = mapVal
		} else {
			// If index exists but is not a map, and we need to go deeper
			// create a new map (this will overwrite the existing value)
			nextObj = make(map[string]any)
			arr[nextIndex] = nextObj
		}

		// Update the array
		obj[part] = arr

		// Process remaining parts
		assignToNested(nextObj, parts[2:], value, options)
	} else {
		// Next part is not an array index
		var nextObj map[string]any

		if existing, ok := obj[part]; ok {
			// Try to use existing map
			if existingMap, ok := existing.(map[string]any); ok {
				nextObj = existingMap
			} else {
				// Convert to map
				nextObj = make(map[string]any)
				obj[part] = nextObj
			}
		} else {
			// Create new map
			nextObj = make(map[string]any)
			obj[part] = nextObj
		}

		// Process remaining parts
		assignToNested(nextObj, parts[1:], value, options)
	}
}

// convertNumericMapsToArrays recursively converts maps with consecutive numeric keys to arrays
func convertNumericMapsToArrays(obj map[string]any) map[string]any {
	for key, value := range obj {
		switch val := value.(type) {
		case map[string]any:
			// First recursively process nested maps
			processedMap := convertNumericMapsToArrays(val)

			// Then check if the processed map should be converted to array
			allNumeric := true
			hasKeys := false
			maxIdx := -1

			for k := range processedMap {
				hasKeys = true
				if idx, err := strconv.Atoi(k); err == nil {
					if idx > maxIdx {
						maxIdx = idx
					}
				} else {
					allNumeric = false
					break
				}
			}

			if allNumeric && hasKeys {
				// Create array with correct size
				arr := make([]any, maxIdx+1)

				// Fill array with values from map
				for k, v := range processedMap {
					idx, _ := strconv.Atoi(k)
					arr[idx] = v
				}

				// Replace map with array
				obj[key] = arr
			} else {
				// Keep as map but with processed values
				obj[key] = processedMap
			}

		case []any:
			// Process each element in the array
			for i, item := range val {
				if nestedMap, ok := item.(map[string]any); ok {
					// Process nested maps in arrays
					processedItem := convertNumericMapsToArrays(nestedMap)

					// Check if result should be an array
					allNumeric := true
					hasKeys := false
					maxIdx := -1

					for k := range processedItem {
						hasKeys = true
						if idx, err := strconv.Atoi(k); err == nil {
							if idx > maxIdx {
								maxIdx = idx
							}
						} else {
							allNumeric = false
							break
						}
					}

					if allNumeric && hasKeys {
						// Convert to array
						nestedArr := make([]any, maxIdx+1)
						for k, v := range processedItem {
							idx, _ := strconv.Atoi(k)
							nestedArr[idx] = v
						}
						val[i] = nestedArr
					} else {
						val[i] = processedItem
					}
				}
				// If not a map, keep as is
			}
		}
	}

	return obj
}
