# fitobj

A lightweight Go tool for flattening and unflattening nested JSON objects, with i18n key management.

## Features

- **Flatten nested objects** with customizable separators and array notation
- **Unflatten objects** back into nested structures
- **Batch processing** for multiple JSON files
- **API mode** with RESTful endpoints
- **Parallel processing** for improved performance
- **i18n key management** for detecting missing or unused translation keys

## Installation

```bash
# Install using go
go install github.com/haiyon/fitobj@latest

# Or clone and build
git clone https://github.com/haiyon/fitobj.git
cd fitobj
go build -ldflags "-X github.com/haiyon/fitobj/cmd.version=0.2.0"
```

## Usage

### Command Line Interface

#### Flatten JSON files

```bash
fitobj flatten ./examples/nested ./examples/flattened
fitobj flatten ./nested ./flat --separator="__" --array-format=bracket --workers=8
```

#### Unflatten JSON files

```bash
fitobj unflatten ./examples/flattened ./examples/nested
fitobj unflatten ./flat ./nested --separator="__"
```

#### i18n Key Management

```bash
# Check for missing and unused i18n keys
fitobj i18n check ./src ./translations/en.json

# Automatically remove unused keys
fitobj i18n clean ./src ./translations
```

#### API Server

```bash
fitobj api --port=8080
```

### Configuration File

Create a `.fitobj.yaml` file in your home directory or project root:

```yaml
separator: "."
array-format: "index"
workers: 4
buffer: 16
api:
  port: "8080"
```

### API Usage

Example request:

```bash
curl -X POST http://localhost:8080/process \
  -H "Content-Type: application/json" \
  -d '{"data": {"user": {"name": "John", "address": {"city": "New York"}}}, "reverse": false}'
```

### Library Usage

```go
import (
    "github.com/haiyon/fitobj/fitter"
    "github.com/haiyon/fitobj/i18n"
)

// Flatten a nested object
flatObj := fitter.FlattenMap(nestedObj, "")

// With custom options
options := fitter.DefaultFlattenOptions()
options.Separator = "__"
options.ArrayFormatting = "bracket"
customFlatObj := fitter.FlattenMapWithOptions(nestedObj, "", options)

// Unflatten back to nested structure
nestedAgain := fitter.UnflattenMap(flatObj)

// i18n key management
sourceKeys, _ := i18n.ExtractKeysFromDir("./src")
jsonKeys, _ := i18n.ExtractKeysFromJSONDir("./translations")
missingInJSON, unusedInSource := i18n.CompareKeys(sourceKeys, jsonKeys)
```

## Examples

### Nested object

```json
{
  "person": {
    "name": "John Doe",
    "addresses": [
      {
        "type": "home",
        "street": "123 Main St"
      }
    ]
  }
}
```

### Flattened with dot notation (default)

```json
{
  "person.name": "John Doe",
  "person.addresses.0.type": "home",
  "person.addresses.0.street": "123 Main St"
}
```

### Flattened with bracket notation

```json
{
  "person.name": "John Doe",
  "person.addresses[0].type": "home",
  "person.addresses[0].street": "123 Main St"
}
```

### Commands Reference

```bash
# Global flags (available for all commands)
--separator string      separator character for flattened keys (default ".")
--array-format string   array format: 'index' or 'bracket' (default "index")
--workers int          number of workers for parallel processing (default: CPU count)
--buffer int           initial buffer size for maps (default 16)
--config string        config file (default is $HOME/.fitobj.yaml)

# Available commands
fitobj flatten [input-dir] [output-dir]    # Flatten nested JSON objects
fitobj unflatten [input-dir] [output-dir]  # Unflatten JSON objects
fitobj api [--port=8080]                   # Start API server
fitobj i18n check [source-dir] [json-path] # Check i18n keys
fitobj i18n clean [source-dir] [json-path] # Clean unused i18n keys
fitobj help [command]                      # Help about any command
fitobj version                            # Show version information
```

## Changes from v0.1.0

- **Breaking**: Replaced flags with subcommands for better UX
- **New**: Added Cobra CLI framework with better help and structure
- **New**: Added configuration file support
- **New**: Added health check endpoint for API mode
- **Fixed**: Improved nested array handling in flatten/unflatten
- **Fixed**: Better error handling and validation
- **Fixed**: Enhanced bracket notation support

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
