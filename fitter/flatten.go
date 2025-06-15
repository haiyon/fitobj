package fitter

import (
	"fmt"
	"strconv"
)

// FlattenOptions configures the flattening process
type FlattenOptions struct {
	Separator           string // separator for nested keys (default: ".")
	MaxDepth            int    // max recursion depth (-1 = no limit)
	IncludeArrayIndices bool   // whether to include array indices
	ArrayFormatting     string // "index" or "bracket"
	BufferSize          int    // initial capacity for result maps
}

// DefaultFlattenOptions returns the default options for flattening
func DefaultFlattenOptions() FlattenOptions {
	return FlattenOptions{
		Separator:           ".",
		MaxDepth:            -1,
		IncludeArrayIndices: true,
		ArrayFormatting:     "index",
		BufferSize:          16,
	}
}

// FlattenMap converts a nested map into a flattened structure using default options
func FlattenMap(obj map[string]any, prefix string) map[string]any {
	return FlattenMapWithOptions(obj, prefix, DefaultFlattenOptions())
}

// FlattenMapWithOptions converts a nested map into a flattened structure with custom options
func FlattenMapWithOptions(obj map[string]any, prefix string, options FlattenOptions) map[string]any {
	result := make(map[string]any, options.BufferSize)
	flatten(obj, prefix, result, options, 0)
	return result
}

// flatten recursively flattens a nested map
func flatten(obj map[string]any, prefix string, result map[string]any, options FlattenOptions, depth int) {
	// Check depth limit
	if options.MaxDepth >= 0 && depth > options.MaxDepth {
		if prefix != "" {
			result[prefix] = obj
		} else {
			for k, v := range obj {
				result[k] = v
			}
		}
		return
	}

	for key, value := range obj {
		var fullKey string
		if prefix == "" {
			fullKey = key
		} else {
			fullKey = prefix + options.Separator + key
		}

		switch typedValue := value.(type) {
		case map[string]any:
			if len(typedValue) == 0 {
				result[fullKey] = typedValue
			} else {
				flatten(typedValue, fullKey, result, options, depth+1)
			}

		case []any:
			if len(typedValue) == 0 {
				result[fullKey] = typedValue
			} else if options.IncludeArrayIndices {
				flattenArray(typedValue, fullKey, result, options, depth)
			} else {
				result[fullKey] = typedValue
			}

		default:
			result[fullKey] = value
		}
	}
}

// flattenArray handles array flattening with proper recursion
func flattenArray(arr []any, prefix string, result map[string]any, options FlattenOptions, depth int) {
	for i, item := range arr {
		var indexedKey string
		if options.ArrayFormatting == "bracket" {
			indexedKey = fmt.Sprintf("%s[%d]", prefix, i)
		} else {
			indexedKey = prefix + options.Separator + strconv.Itoa(i)
		}

		switch itemTyped := item.(type) {
		case map[string]any:
			if len(itemTyped) == 0 {
				result[indexedKey] = itemTyped
			} else {
				flatten(itemTyped, indexedKey, result, options, depth+1)
			}
		case []any:
			if len(itemTyped) == 0 {
				result[indexedKey] = itemTyped
			} else {
				flattenArray(itemTyped, indexedKey, result, options, depth+1)
			}
		default:
			result[indexedKey] = item
		}
	}
}
