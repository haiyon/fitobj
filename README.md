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
go install github.com/haiyon/fitobj

# Or clone and build
git clone https://github.com/haiyon/fitobj.git
cd fitobj
go build -ldflags "-X main.version=0.1.0"
```

## Usage

### Command Line

#### Flatten JSON files

```bash
./fitobj -input=./examples/nested -output=./examples/flattened
```

#### Unflatten JSON files

```bash
./fitobj -input=./examples/flattened -output=./examples/nested -reverse
```

#### Custom options

```bash
./fitobj -input=./nested -output=./flat -separator="__" -array-format=bracket -workers=8
```

#### i18n Key Management

```bash
# Extract and compare i18n keys
./fitobj -i18n -source-dir=./src -json-path=./translations/en.json
```

### API Server

```bash
./fitobj -api -port=8080
```

Example request:

```bash
curl -X POST http://localhost:8080/process \
  -H "Content-Type: application/json" \
  -d '{"data": {"user": {"name": "John", "address": {"city": "New York"}}}, "reverse": false}'
```

### Library Usage

```go
import "github.com/haiyon/fitobj/fitter"

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
import "github.com/haiyon/fitobj/i18n"

// Extract keys from source and JSON
sourceKeys, _ := i18n.ExtractKeysFromDir("./src")
jsonKeys, _ := i18n.ExtractKeysFromJSONDir("./translations")

// Compare keys to find missing or unused ones
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

### i18n Key Management

Extract and compare translation keys between source code and JSON files:

```bash
./fitobj -i18n -source-dir=./examples/i18n/source -json-path=./examples/i18n/translations

🔍 Total keys in source: 15
📚 Total keys in JSON: 16

❌ Missing in JSON (2):
buttons.getStarted
footer.terms

🟡 Unused in Source (3):
nav.products
buttons.cancel
footer.privacy
```

> Note: This project has been optimized with the assistance of Claude.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
