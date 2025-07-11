# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool for running Bitbucket pipelines locally by parsing bitbucket-pipelines.yml files and executing the defined steps. The tool provides a foundation for local CI/CD testing and automation.

## Build and Development Commands

### Core Commands
- **Build project**: `make build`
- **Run tests**: `make test`
- **Run with race detection**: `make test-race`
- **Run integration tests**: `make test-integration` (requires Docker)
- **Test coverage**: `make test-coverage`
- **Clean build artifacts**: `make clean`
- **Install locally**: `make install`

### Cross-Platform Build
- **Build for all platforms**: `make build-all`
- **Create release artifacts**: `make release`

### Development Workflow
- **Development mode**: `make dev`
- **Watch and rebuild**: `make watch` (requires air)
- **Download dependencies**: `make deps`
- **Run linter**: `make lint` (requires golangci-lint)

### Docker Commands
- **Build Docker image**: `make docker-build`
- **Run in Docker**: `make docker-run`
- **Using Docker directly**: `docker run --rm -v $(pwd):/workspace ghcr.io/vek-servicos/bitbucket-runner:latest`

### CLI Usage
- **List pipelines**: `./bin/bitbucket-runner list`
- **Run default pipeline**: `./bin/bitbucket-runner run`
- **Run specific pipeline**: `./bin/bitbucket-runner run --pipeline=custom-build`
- **Help**: `./bin/bitbucket-runner --help`

## Architecture

### Project Structure
```
bitbucket-runner/
├── cmd/                 # CLI commands (Cobra-based)
│   ├── root.go         # Root command definition
│   ├── list.go         # List pipelines command
│   ├── run.go          # Run pipeline command
│   └── *_test.go       # Command tests
├── internal/           # Private application code
│   ├── models/         # Data structures
│   │   ├── pipeline.go # Pipeline configuration models
│   │   ├── execution.go # Execution context
│   │   └── config.go   # Runner configuration
│   └── parser/         # YAML parsing logic
│       ├── parser.go   # Basic YAML parser
│       └── pipeline.go # Advanced pipeline parser
├── testdata/           # Test configurations
├── docs/               # Documentation and stories
├── bin/                # Built binaries
└── main.go            # Application entry point
```

### Core Components

#### CLI Framework (Cobra)
- **Root Command**: Base command with help and version information
- **List Command**: Displays available pipelines from bitbucket-pipelines.yml
- **Run Command**: Executes pipeline steps (currently parsing only)
- **Command Pattern**: Each command is a separate file in cmd/ directory

#### Data Models
- **PipelineConfig**: Represents the complete bitbucket-pipelines.yml structure
- **Pipeline**: Array of StepWrapper containing individual steps
- **Step**: Individual pipeline step with script, image, services, etc.
- **Definitions**: Services and caches definitions
- **ExecutionContext**: Tracks execution state and runtime information
- **RunnerConfig**: Tool configuration and step type mappings

#### YAML Parser
- **ParsePipelineConfig**: Main parsing function using gopkg.in/yaml.v3
- **Custom Unmarshaling**: Handles complex YAML structures like string/object caches
- **Validation**: Comprehensive validation of pipeline structure and steps
- **Error Handling**: Detailed error messages for parsing failures

### Key Design Patterns

#### Configuration-Driven Design
- All behavior configurable via YAML parsing
- Flexible step definitions supporting various Bitbucket pipeline features
- Environment variable support and service definitions

#### Layered Architecture
- CLI layer handles user interaction and command parsing
- Core service layer processes pipeline logic
- Infrastructure layer manages YAML parsing and validation

#### Interface-Based Design
- Structured for testability with clear separation of concerns
- Custom unmarshaling interfaces for complex YAML structures
- Validation interfaces for pipeline components

## Development Guidelines

### Go Standards
- **Version**: Go 1.21.5+
- **Module**: `bitbucket-runner` (local module name)
- **Dependencies**: Cobra for CLI, yaml.v3 for parsing, testify for testing

### Testing Requirements
- **Framework**: Testify for assertions and test suites
- **Coverage**: Aim for comprehensive unit test coverage
- **Test Structure**: Mirror source code structure with _test.go suffix
- **Test Data**: Use testdata/ directory for sample pipeline files
- **Integration Tests**: Docker-based integration tests with `-tags=integration`

### Code Quality
- **Linting**: Use golangci-lint for code quality checks
- **Error Handling**: Comprehensive error handling with detailed messages
- **Validation**: All input validation with clear error messages
- **Documentation**: Godoc comments for all public functions and types

### Build System
- **Makefile**: Comprehensive build automation with help target
- **Cross-Platform**: Support for Linux, macOS, Windows (amd64/arm64)
- **Docker**: Multi-stage builds with alpine base
- **Version Info**: Build-time version injection via ldflags

## Pipeline Configuration Support

### Supported Features
- **Images**: Global and step-specific Docker images
- **Scripts**: Multi-line shell scripts for step execution
- **Services**: Docker services for databases, caches, etc.
- **Artifacts**: File artifacts configuration
- **Caches**: Both simple string and complex object caches
- **Environment**: Environment variable definitions
- **Conditions**: Changeset-based conditional execution
- **Pipeline Types**: Default, branches, pull-requests, custom, tags

### YAML Structure Support
- **Anchors and References**: Full YAML anchor/reference support
- **Complex Configurations**: Nested structures and advanced features
- **Validation**: Comprehensive validation of all configuration elements
- **Error Reporting**: Detailed error messages for invalid configurations

## Distribution and Installation

### Installation Methods
- **Global Install**: `go install github.com/vek-servicos/bitbucket-runner@latest`
- **Docker**: `docker run --rm -v $(pwd):/workspace ghcr.io/vek-servicos/bitbucket-runner:latest`
- **Local Build**: `make build` then use `./bin/bitbucket-runner`

### CI/CD Integration
- **GitHub Actions**: Automated testing and releases
- **Multi-Platform**: Builds for all major platforms
- **Container Registry**: Images available via GitHub Container Registry
- **Release Automation**: GoReleaser for automated releases

### Development Status
- **Foundation**: Complete with full YAML parsing and CLI framework
- **Current Phase**: Basic parsing and listing functionality implemented
- **Next Phase**: Execution engine for actually running pipeline steps
- **Integration**: Designed for use with vks-jee-auth-api and other VEK projects

## Testing Strategy

### Unit Tests
- **Models**: Test all data structures and validation logic
- **Parser**: Test YAML parsing with various configurations
- **Commands**: Test CLI commands with mock data
- **Coverage**: Use `make test-coverage` to generate coverage reports

### Integration Tests
- **Docker Required**: Integration tests require Docker for step execution
- **Tag-based**: Use `-tags=integration` for integration-specific tests
- **End-to-End**: Test complete pipeline parsing and execution flow

### Test Data
- **Location**: testdata/ directory contains sample pipeline files
- **Structure**: Mirror real-world Bitbucket pipeline configurations
- **Validation**: Test files validate parser robustness and error handling