package fitter

import (
	"regexp"
	"strconv"
	"strings"
)

// UnflattenOptions configures the unflattening process
type UnflattenOptions struct {
	Separator              string // separator for flattened keys
	DetectArrays           bool   // auto convert numeric indices to arrays
	SupportBracketNotation bool   // support key[0] notation
	BufferSize             int    // initial capacity for result maps
}

// DefaultUnflattenOptions returns the default options for unflattening
func DefaultUnflattenOptions() UnflattenOptions {
	return UnflattenOptions{
		Separator:              ".",
		DetectArrays:           true,
		SupportBracketNotation: true,
		BufferSize:             16,
	}
}

// UnflattenMap converts a flattened map back into a nested structure
func UnflattenMap(obj map[string]any) map[string]any {
	return UnflattenMapWithOptions(obj, DefaultUnflattenOptions())
}

// UnflattenMapWithOptions converts a flattened map back into a nested structure with custom options
func UnflattenMapWithOptions(obj map[string]any, options UnflattenOptions) map[string]any {
	result := make(map[string]any)

	// Preprocess keys to convert bracket notation if needed
	processedObj := make(map[string]any, len(obj))
	for k, v := range obj {
		if options.SupportBracketNotation {
			k = convertBracketToDot(k, options.Separator)
		}
		processedObj[k] = v
	}

	// Process each key-value pair
	for key, value := range processedObj {
		parts := strings.Split(key, options.Separator)
		assignToNested(result, parts, value, options)
	}

	// Convert numeric maps to arrays
	if options.DetectArrays {
		return convertNumericMapsToArrays(result)
	}

	return result
}

// convertBracketToDot converts "user[0].name" to "user.0.name"
func convertBracketToDot(key, separator string) string {
	re := regexp.MustCompile(`\[([0-9]+)\]`)
	return re.ReplaceAllString(key, separator+"$1")
}

// assignToNested sets a value at a path in a nested structure
func assignToNested(obj map[string]any, parts []string, value any, options UnflattenOptions) {
	if len(parts) == 0 {
		return
	}

	part := parts[0]

	if len(parts) == 1 {
		obj[part] = value
		return
	}

	// Check if we need to create an array or object
	nextIsNumeric := false
	nextIndex := -1

	if options.DetectArrays {
		if idx, err := strconv.Atoi(parts[1]); err == nil {
			nextIsNumeric = true
			nextIndex = idx
		}
	}

	if nextIsNumeric {
		// Handle array creation/extension
		var arr []any
		if existing, ok := obj[part]; ok {
			if existingArr, ok := existing.([]any); ok {
				arr = existingArr
			} else {
				arr = make([]any, nextIndex+1)
			}
		} else {
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
			nextObj = make(map[string]any)
			arr[nextIndex] = nextObj
		}

		obj[part] = arr
		assignToNested(nextObj, parts[2:], value, options)
	} else {
		// Handle object creation
		var nextObj map[string]any
		if existing, ok := obj[part]; ok {
			if existingMap, ok := existing.(map[string]any); ok {
				nextObj = existingMap
			} else {
				nextObj = make(map[string]any)
				obj[part] = nextObj
			}
		} else {
			nextObj = make(map[string]any)
			obj[part] = nextObj
		}

		assignToNested(nextObj, parts[1:], value, options)
	}
}

// convertNumericMapsToArrays recursively converts maps with consecutive numeric keys to arrays
func convertNumericMapsToArrays(obj map[string]any) map[string]any {
	for key, value := range obj {
		switch val := value.(type) {
		case map[string]any:
			processedMap := convertNumericMapsToArrays(val)

			// Check if this map should become an array
			if shouldConvertToArray(processedMap) {
				obj[key] = convertMapToArray(processedMap)
			} else {
				obj[key] = processedMap
			}

		case []any:
			for i, item := range val {
				if nestedMap, ok := item.(map[string]any); ok {
					processedItem := convertNumericMapsToArrays(nestedMap)
					if shouldConvertToArray(processedItem) {
						val[i] = convertMapToArray(processedItem)
					} else {
						val[i] = processedItem
					}
				}
			}
		}
	}

	return obj
}

// shouldConvertToArray checks if a map should be converted to an array
func shouldConvertToArray(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}

	maxIdx := -1
	for k := range m {
		if idx, err := strconv.Atoi(k); err == nil {
			if idx > maxIdx {
				maxIdx = idx
			}
		} else {
			return false
		}
	}

	// Check for consecutive indices starting from 0
	for i := 0; i <= maxIdx; i++ {
		if _, exists := m[strconv.Itoa(i)]; !exists {
			return false
		}
	}

	return true
}

// convertMapToArray converts a numeric-keyed map to an array
func convertMapToArray(m map[string]any) []any {
	maxIdx := -1
	for k := range m {
		if idx, err := strconv.Atoi(k); err == nil && idx > maxIdx {
			maxIdx = idx
		}
	}

	arr := make([]any, maxIdx+1)
	for k, v := range m {
		if idx, err := strconv.Atoi(k); err == nil {
			arr[idx] = v
		}
	}

	return arr
}
