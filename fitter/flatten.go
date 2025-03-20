package fitter

import (
	"fmt"
	"strconv"
)

// FlattenOptions configures the flattening process
type FlattenOptions struct {
	// Separator is the character used to separate nested keys (default: ".")
	Separator string
	// MaxDepth limits the recursion depth (-1 means no limit)
	MaxDepth int
	// IncludeArrayIndices determines whether array indices should be included
	IncludeArrayIndices bool
	// ArrayFormatting specifies how array indices should be formatted
	// "index" = user.addresses.0.street (default)
	// "bracket" = user.addresses[0].street
	ArrayFormatting string
	// BufferSize is the initial capacity of result maps to reduce allocations
	BufferSize int
}

// DefaultFlattenOptions returns the default options for flattening
func DefaultFlattenOptions() FlattenOptions {
	return FlattenOptions{
		Separator:           ".",
		MaxDepth:            -1,
		IncludeArrayIndices: true,
		ArrayFormatting:     "index",
		BufferSize:          16, // Default starting capacity for map
	}
}

// FlattenMap converts a nested map into a flattened structure using default options.
// Example: {"user": {"name": "John"}} becomes {"user.name": "John"}
//
// Parameters:
// - obj: The nested map to flatten
// - prefix: The prefix to use for keys (used in recursion)
//
// Returns:
// - A new map with flattened structure
func FlattenMap(obj map[string]any, prefix string) map[string]any {
	return FlattenMapWithOptions(obj, prefix, DefaultFlattenOptions())
}

// FlattenMapWithOptions converts a nested map into a flattened structure with custom options.
//
// Parameters:
// - obj: The nested map to flatten
// - prefix: The prefix to use for keys (used in recursion)
// - options: Configuration options for the flattening process
//
// Returns:
// - A new map with flattened structure
func FlattenMapWithOptions(obj map[string]any, prefix string, options FlattenOptions) map[string]any {
	// Initialize with specified buffer size for better performance
	result := make(map[string]any, options.BufferSize)
	flatten(obj, prefix, result, options, 0)
	return result
}

// flatten recursively flattens a nested map into a flattened structure
func flatten(obj map[string]any, prefix string, result map[string]any, options FlattenOptions, depth int) {
	// Check depth limit
	if options.MaxDepth >= 0 && depth > options.MaxDepth {
		// Depth limit exceeded, add the object as is
		if prefix != "" {
			result[prefix] = obj
		} else {
			// If prefix is empty, copy the original map
			for k, v := range obj {
				result[k] = v
			}
		}
		return
	}

	for key, value := range obj {
		// Create the prefixed key
		var fullKey string
		if prefix == "" {
			fullKey = key
		} else {
			fullKey = prefix + options.Separator + key
		}

		// Handle different value types
		switch typedValue := value.(type) {
		case map[string]any:
			// Recursively flatten nested maps
			if len(typedValue) == 0 {
				// Empty map, directly add
				result[fullKey] = typedValue
			} else {
				// Non-empty map, recursively flatten
				flatten(typedValue, fullKey, result, options, depth+1)
			}

		case []any:
			// Handle arrays
			if len(typedValue) == 0 {
				// Empty array, directly add
				result[fullKey] = typedValue
			} else if options.IncludeArrayIndices {
				// Add array elements with indices
				for i, item := range typedValue {
					var indexedKey string
					if options.ArrayFormatting == "bracket" {
						indexedKey = fmt.Sprintf("%s[%d]", fullKey, i)
					} else {
						indexedKey = fullKey + options.Separator + strconv.Itoa(i)
					}

					// Handle different item types
					switch itemTyped := item.(type) {
					case map[string]any:
						// Nested map, recursively flatten
						if len(itemTyped) == 0 {
							result[indexedKey] = itemTyped
						} else {
							flatten(itemTyped, indexedKey, result, options, depth+1)
						}
					case []any:
						// Nested array - this case was missing in the original
						if len(itemTyped) == 0 {
							result[indexedKey] = itemTyped
						} else if options.IncludeArrayIndices {
							// Handle nested arrays recursively
							for j, nestedItem := range itemTyped {
								var nestedIndexedKey string
								if options.ArrayFormatting == "bracket" {
									nestedIndexedKey = fmt.Sprintf("%s[%d]", indexedKey, j)
								} else {
									nestedIndexedKey = indexedKey + options.Separator + strconv.Itoa(j)
								}

								if nestedMap, ok := nestedItem.(map[string]any); ok {
									if len(nestedMap) == 0 {
										result[nestedIndexedKey] = nestedMap
									} else {
										flatten(nestedMap, nestedIndexedKey, result, options, depth+2)
									}
								} else {
									result[nestedIndexedKey] = nestedItem
								}
							}
						} else {
							result[indexedKey] = itemTyped
						}
					default:
						// Simple value, directly add
						result[indexedKey] = item
					}
				}
			} else {
				// If not including array indices, directly add
				result[fullKey] = typedValue
			}

		default:
			// Simple value, directly add
			result[fullKey] = value
		}
	}
}
