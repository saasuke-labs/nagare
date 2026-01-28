# Nagare Refactoring Summary

## Changes Made

This refactoring enables nagare to be used in three modes:

### 1. HTTP Server Mode (existing, enhanced)

- Added `-port` flag to configure server port
- Default behavior when no CLI args provided
- Start with: `nagare` or `nagare -port 3000`

### 2. CLI Mode (new)

- Pass input and output file paths
- Usage: `nagare -input diagram.nagare -output diagram.svg`

### 3. Library Mode (new)

- Import as a Go library in other projects
- Three exported functions in `pkg/nagare` package:
  - `RenderToSVG(code string) (string, error)` - Main library entry point
  - `RenderToSVGWithDebug(code string) (string, error)` - With debug output
  - `RenderFileToFile(inputPath, outputPath string) error` - File-based rendering

## New Files

### `pkg/nagare/nagare.go`

New public API package containing the three main functions for library usage.

### `docs/LIBRARY_USAGE.md`

Comprehensive documentation with examples of using nagare as a library.

## Modified Files

### `cmd/main.go`

- Added CLI flag parsing (`-input`, `-output`, `-port`)
- Added mode detection logic (CLI vs HTTP server)
- Added usage help text

### `pkg/diagram/diagram.go`

- Refactored to use the new `pkg/nagare` library internally
- Maintains backward compatibility for HTTP server

### `README.md`

- Updated with all three usage modes
- Added installation and usage examples

## Usage Examples

### HTTP Server

```bash
# Default port 8080
nagare

# Custom port
nagare -port 3000
```

### CLI

```bash
nagare -input diagram.nagare -output diagram.svg
```

### Library

```go
import "github.com/saasuke-labs/nagare/pkg/nagare"

svg, err := nagare.RenderToSVG(diagramCode)
```

## Testing

- All existing tests pass (pkg/diagram has a pre-existing failure unrelated to this refactor)
- Manually tested CLI mode with example files
- Binary builds successfully

## Backward Compatibility

- HTTP server mode maintains full backward compatibility
- Existing HTTP endpoints unchanged
- No breaking changes to existing functionality
