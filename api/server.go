package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/haiyon/fitobj/fitter"
)

// Options configures the API server behavior
type Options struct {
	Port          string
	FlattenOpts   fitter.FlattenOptions
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

// Request defines the structure for API requests
type Request struct {
	Data        map[string]any `json:"data"`
	Reverse     bool           `json:"reverse"`
	Separator   string         `json:"separator,omitempty"`
	ArrayFormat string         `json:"arrayFormat,omitempty"`
}

// Response defines the structure for API responses
type Response struct {
	Data    map[string]any `json:"data"`
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
}

// ErrorResponse defines the structure for error responses
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type server struct {
	options Options
}

func newServer(options Options) *server {
	return &server{options: options}
}

// ProcessHandler handles API requests to process JSON data
func (s *server) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	if request.Data == nil {
		s.sendError(w, "No data provided in request", http.StatusBadRequest)
		return
	}

	// Create options copies
	flattenOpts := s.options.FlattenOpts
	unflattenOpts := s.options.UnflattenOpts

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
			message = "Warning: Invalid array format specified, using default ('index')."
		}
	}

	// Process the data
	var result map[string]any
	if request.Reverse {
		result = fitter.UnflattenMapWithOptions(request.Data, unflattenOpts)
	} else {
		result = fitter.FlattenMapWithOptions(request.Data, "", flattenOpts)
	}

	// Send response
	response := Response{
		Data:    result,
		Success: true,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.sendError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// StartServer starts the API server on the specified port
func StartServer(port string) error {
	options := DefaultOptions()
	options.Port = port
	return StartServerWithOptions(options)
}

// StartServerWithOptions starts the API server with custom options
func StartServerWithOptions(options Options) error {
	// Ensure proper defaults
	if options.FlattenOpts.MaxDepth == 0 {
		options.FlattenOpts.MaxDepth = -1
	}
	if !options.FlattenOpts.IncludeArrayIndices {
		options.FlattenOpts.IncludeArrayIndices = true
	}

	s := newServer(options)

	// Register handlers
	http.HandleFunc("/process", s.ProcessHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Printf("API server running at http://localhost:%s/process\n", options.Port)
	fmt.Printf("Health check available at http://localhost:%s/health\n", options.Port)
	fmt.Printf("Using separator: '%s', array format: '%s'\n",
		options.FlattenOpts.Separator,
		options.FlattenOpts.ArrayFormatting)

	return http.ListenAndServe(":"+options.Port, nil)
}
