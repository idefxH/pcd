# pcdp-lint

A command-line tool for validating specification files written in the Post-Coding Development Paradigm format.

## Features

- Validates structural requirements of specification files
- Checks META field presence and format
- Validates SPDX license identifiers
- Ensures deployment template compliance
- Supports strict mode (treats warnings as errors)
- Lists available deployment templates

## Installation

### From Package Manager

The tool is available as RPM and DEB packages via the OpenSUSE Build Service (OBS):

```bash
# For openSUSE/SLES
zypper install pcdp-lint

# For Debian/Ubuntu  
apt install pcdp-lint

# For Fedora
dnf install pcdp-lint
```

### From Source

```bash
make build
sudo make install
```

## Usage

### Lint a specification file

```bash
pcdp-lint myspec.md
```

### Strict mode (treat warnings as errors)

```bash
pcdp-lint strict=true myspec.md
```

### List available deployment templates

```bash
pcdp-lint list-templates
```

## Output

The tool writes diagnostics to stderr and a summary to stdout:

- Exit code 0: Valid (no errors, or warnings in non-strict mode)
- Exit code 1: Invalid (errors present, or warnings in strict mode)  
- Exit code 2: Invocation error (bad arguments, file not found)

## Requirements

- Linux (primary platform)
- No runtime dependencies (static binary)
- Input files must have `.md` extension

## License

Apache-2.0