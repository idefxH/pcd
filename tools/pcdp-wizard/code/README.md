# pcdp-wizard

Interactive wizard for creating Post-Coding Development Paradigm (PCDP) specifications.

## Installation

### Via OBS (Recommended)

Install from the openSUSE Build Service packages:

**openSUSE Leap / Tumbleweed:**
```bash
zypper install pcdp-tools
```

**Fedora:**
```bash
dnf install pcdp-tools
```

**Debian / Ubuntu:**
```bash
apt update
apt install pcdp-tools
```

## Usage

### Start a New Specification

```bash
pcdp-wizard
```

The wizard will interactively guide you through creating a complete PCDP specification.

### Specify Output Path

```bash
pcdp-wizard output=./specs/my-component.md
```

### List Resumable Sessions

```bash
pcdp-wizard list-sessions
```

Shows all incomplete wizard sessions that can be resumed.

### Resume a Session

Simply run `pcdp-wizard` again with the same component name to resume an incomplete session.

## Command Line Arguments

### Key=Value Options

- `output=<path>` - Write specification to this path (default: `./<component-name>.md`)

### Commands (Bare Words)

- `list-sessions` - List all resumable wizard sessions and exit

## Exit Codes

- `0` - Success (specification written and valid)
- `1` - Specification written but contains validation errors
- `2` - Invocation error (bad arguments, missing templates, unwritable path)

## Features

- **Interactive Guidance**: Step-by-step prompts for each specification section
- **Session Resume**: Automatically saves progress and allows resuming incomplete sessions
- **Template Support**: Automatically detects available deployment templates
- **Validation Integration**: Runs `pcdp-lint` automatically to validate generated specifications
- **Smart Defaults**: Suggests reasonable defaults based on context and user environment

## Specification Sections

The wizard guides you through creating all required PCDP sections:

1. **META** - Component metadata (name, version, author, license)
2. **TYPES** - Custom data type definitions
3. **BEHAVIOR** - Component behavior specifications
4. **PRECONDITIONS** - Required conditions before execution
5. **POSTCONDITIONS** - Guaranteed conditions after execution
6. **INVARIANTS** - Conditions that must always hold
7. **EXAMPLES** - Concrete usage examples with GIVEN/WHEN/THEN format
8. **DEPLOYMENT** - Runtime and deployment information

## Session Management

Sessions are stored in `~/.config/pcdp/wizard-state/` as JSON files. Each session:

- Has a unique UUID identifier
- Tracks completion status of each section
- Stores the target output path
- Can be resumed at any time
- Is automatically deleted upon successful completion

## Integration with pcdp-lint

The wizard automatically runs `pcdp-lint` on generated specifications to validate correctness. If validation fails:

- The specification file is still written
- Error details are displayed
- The session is retained for fixing and resuming
- Exit code 1 is returned

## Requirements

- **Required**: Go 1.19+ (for building from source)
- **Recommended**: `pcdp-lint` installed for post-generation validation
- **Platform**: Linux (primary), macOS (supported)

## Examples

### Creating a CLI Tool Specification

```bash
$ pcdp-wizard
Component name: file-processor
Available deployment templates:
1. cli-tool
2. library
3. service
Select deployment template (number): 1
Version [0.1.0]: 
Author [John Doe <john@example.com>]: 
Common licenses: Apache-2.0, MIT, GPL-2.0-only, GPL-3.0-only, CC-BY-4.0
License (SPDX identifier) [Apache-2.0]: MIT
...
✓ file-processor.md: written and valid
```

### Resuming a Session

```bash
$ pcdp-wizard
Resuming session for 'file-processor' (started 2024-01-15 14:30)
Completed: META, TYPES, BEHAVIOR
Continuing from: PRECONDITIONS
...
```

### Listing Sessions

```bash
$ pcdp-wizard list-sessions
Session: 550e8400-e29b-41d4-a716-446655440000
Component: file-processor
Started: 2024-01-15 14:30:25
Last Updated: 2024-01-15 14:45:12
Sections Completed: META, TYPES, BEHAVIOR
Partial Spec: ./file-processor.md
```

## Development

### Building from Source

```bash
git clone <repository>
cd pcdp-wizard
make build
```

### Running Tests

```bash
make test
```

### Installing Locally

```bash
make install
```

This installs to `/usr/local/bin/pcdp-wizard`.

## License

GPL-2.0-only

## Support

For issues, questions, or contributions, please refer to the project repository or contact the maintainer listed in the META section.