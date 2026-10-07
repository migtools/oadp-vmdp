# AGENTS.md — AI Agent Instructions for migtools/oadp-vmdp

## Project Overview
OADP VM Data Protection (oadp-vmdp) is a specialized fork of Kopia that adds VM-specific data protection capabilities for the OADP ecosystem. It extends Kopia's backup/restore functionality with features tailored for virtual machine data protection in OpenShift/KubeVirt environments, including custom binaries published to Quay.io.

- **Primary Language**: Go
- **Module**: `github.com/kopia/kopia`
- **Default Branch**: `oadp-dev`

## Build Instructions
```bash
# Build the kopia binary (with OADP VM extensions)
make install

# Build without HTML UI
make install-noui

# Build with race detector
make install-race
```

## Test Instructions
```bash
# Run all tests
go test ./...

# Run a specific test
go test ./path/to/package -run TestName

# Vet code
make vet
# or directly:
go vet ./...
```

## Linting
```bash
# Run linter (requires golangci-lint)
make lint

# Run linter with auto-fix
make lint-fix

# Platform-specific linting
make lint-linux
make lint-all
```

Configuration: `.golangci.yml`

## Code Conventions
- Based on Kopia's codebase with OADP-specific extensions
- Standard Go project layout with `internal/` for private packages
- CLI commands in `cli/`
- Additional commands in `cmd/` (OADP-specific binaries)
- Repository backends in `repo/`
- Snapshot management in `snapshot/`
- File system abstractions in `fs/`

## Project Structure
```
cli/           - CLI command definitions
cmd/           - OADP-specific command binaries
fs/            - Filesystem abstraction layer
internal/      - Private packages (crypto, cache, compression, etc.)
notification/  - Notification system
repo/          - Repository and storage backends
site/          - Documentation website
snapshot/      - Snapshot management and policies
tests/         - Integration and end-to-end tests
tools/         - Build tooling
app/           - Desktop application (Electron-based)
```

## CI/CD
- GitHub Actions workflows in `.github/workflows/`:
  - `lint.yml` — Linting checks
  - `tests.yml` — Test suite
  - `quay_binaries_push.yml` — Build and push binaries to Quay.io
- Reproduce CI locally:
  ```bash
  make ci-setup  # Install tools and dependencies
  make lint      # Run linters
  go test ./...  # Run tests
  ```

## Common Tasks

### Adding a new OADP-specific command
1. Create a new Go file in `cmd/`
2. Implement the command using Kopia's internal libraries
3. Add build targets if needed
4. Test with `go test ./cmd/...`

### Modifying snapshot policies
1. Policy code is in `snapshot/policy/`
2. Follow existing patterns for policy options
3. Update CLI flags in `cli/` if adding new policy knobs

### Updating from upstream Kopia
- OADP patches live on the `oadp-dev` branch
- Rebase carefully against upstream `kopia/kopia`
- The `cmd/` directory contains OADP-specific additions not in upstream
