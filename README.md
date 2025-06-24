# gmacs-diff-plugin

A buffer comparison plugin for gmacs that provides diff functionality between buffers.

## Features

- **buffer-diff**: Compare two buffers by name and show differences
- **buffer-diff-current**: Compare current buffer with another buffer

Both commands use sequential argument prompts to collect buffer names, demonstrating gmacs' multi-argument command system.

## Commands

### `buffer-diff`
Compares two buffers and displays the differences in a new buffer.

**Usage:** 
1. Run `M-x buffer-diff`
2. Enter first buffer name when prompted: "Compare buffer: "
3. Enter second buffer name when prompted: "With buffer: "

### `buffer-diff-current`
Compares the current buffer with another buffer.

**Usage:**
1. Run `M-x buffer-diff-current`
2. Enter buffer name when prompted: "Compare current buffer with: "

## Output

The plugin creates a dedicated buffer named `*Diff: buffer1 <-> buffer2*` showing:
- Lines prefixed with `-` indicate content only in the first buffer
- Lines prefixed with `+` indicate content only in the second buffer  
- Lines with no prefix are identical in both buffers

## Installation

```bash
# Build the plugin
go build -o gmacs-diff-plugin .

# Install using gmacs
gmacs plugin install /path/to/gmacs-diff-plugin/
```

## Development

This plugin demonstrates:
- Multi-argument command registration with `ArgPrompts`
- RPC communication between plugin and host
- Buffer manipulation through the gmacs plugin SDK
- Simple line-by-line diff algorithm

## Requirements

- gmacs with multi-argument command support
- gmacs-plugin-sdk

## License

MIT License