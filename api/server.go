package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/haiyon/fitobj/fitter"
)

// Options configures the API server behavior
type Options struct {
	// Port to run the server on
	Port string
	// FlattenOpts contains options for the flatten operation
	FlattenOpts fitter.FlattenOptions
	// UnflattenOpts contains options for the unflatten operation
	UnflattenOpts fitter.UnflattenOptions
}

// DefaultOptions returns the default options for the API server
func DefaultOptions() Options {
	return Options{
		Port:          "8080",
		FlattenOpts:   fitter.DefaultFlattenOptions(),
		UnflattenOpts: fitter.DefaultUnflattenOptions(),
	}
}

// Request defines the structure for API requests.
type Request struct {
	Data        map[string]any `json:"data"`                  // JSON data to process
	Unflatten   bool           `json:"unflatten"`             // Whether to unflatten (true) or flatten (false)
	Separator   string         `json:"separator,omitempty"`   // Custom separator (optional)
	ArrayFormat string         `json:"arrayFormat,omitempty"` // Array format: "index" or "bracket" (optional)
}

// Response defines the structure for API responses.
type Response struct {
	Data    map[string]any `json:"data"`              // Processed JSON data
	Success bool           `json:"success"`           // Whether the operation was successful
	Message string         `json:"message,omitempty"` // Error message, if any
}

// server holds the configuration for the API server
type server struct {
	options Options
}

// newServer creates a new server with the given options
func newServer(options Options) *server {
	return &server{
		options: options,
	}
}

// ProcessHandler handles API requests to process JSON data.
func (s *server) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var request Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Check if data is provided
	if request.Data == nil {
		http.Error(w, "No data provided in request", http.StatusBadRequest)
		return
	}

	// Create a copy of the options to avoid modifying the server's defaults
	flattenOpts := fitter.DefaultFlattenOptions()
	flattenOpts.Separator = s.options.FlattenOpts.Separator
	flattenOpts.ArrayFormatting = s.options.FlattenOpts.ArrayFormatting
	flattenOpts.BufferSize = s.options.FlattenOpts.BufferSize
	unflattenOpts := fitter.DefaultUnflattenOptions()
	unflattenOpts.Separator = s.options.UnflattenOpts.Separator
	unflattenOpts.SupportBracketNotation = s.options.UnflattenOpts.SupportBracketNotation
	unflattenOpts.BufferSize = s.options.UnflattenOpts.BufferSize

	// Message variable
	var message string

	// Apply custom options if provided
	if request.Separator != "" {
		flattenOpts.Separator = request.Separator
		unflattenOpts.Separator = request.Separator
	}

	if request.ArrayFormat != "" {
		if request.ArrayFormat == "index" || request.ArrayFormat == "bracket" {
			flattenOpts.ArrayFormatting = request.ArrayFormat
			unflattenOpts.SupportBracketNotation = request.ArrayFormat == "bracket"
		} else {
			// Invalid array format, use default and add warning
			message = "Warning: Invalid array format specified, using default ('index')."
		}
	}

	// Process the data
	var result map[string]any

	if request.Unflatten {
		// For unflattening, pass the data directly
		result = fitter.UnflattenMapWithOptions(request.Data, unflattenOpts)
	} else {
		// For flattening, pass empty string as prefix (this is crucial)
		result = fitter.FlattenMapWithOptions(request.Data, "", flattenOpts)
	}

	// Prepare and send response
	response := Response{
		Data:    result,
		Success: true,
		Message: message,
	}

	// Set proper content type and encode response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// StartServer starts the API server on the specified port.
func StartServer(port string) error {
	options := DefaultOptions()
	options.Port = port
	return StartServerWithOptions(options)
}

// StartServerWithOptions starts the API server with custom options.
func StartServerWithOptions(options Options) error {
	if options.FlattenOpts.MaxDepth == 0 {
		options.FlattenOpts.MaxDepth = -1
	}
	if !options.FlattenOpts.IncludeArrayIndices {
		options.FlattenOpts.IncludeArrayIndices = true
	}
	s := newServer(options)

	// Register handlers
	http.HandleFunc("/process", s.ProcessHandler)

	// Start server
	fmt.Printf("API server running at http://localhost:%s/process\n", options.Port)
	fmt.Printf("Using separator: '%s', array format: '%s'\n",
		options.FlattenOpts.Separator,
		options.FlattenOpts.ArrayFormatting)

	return http.ListenAndServe(":"+options.Port, nil)
}
