# Bitbucket Runner

A CLI tool for running Bitbucket pipelines locally by parsing bitbucket-pipelines.yml files and executing the defined steps.

## Features

- ✅ Parse complex Bitbucket Pipeline YAML configurations
- ✅ Execute pipeline steps in Docker containers
- ✅ Support for environment variables and step dependencies
- ✅ Structured logging and error handling
- ✅ Cross-platform support (Linux, macOS, Windows)

## Installation

### Global Installation
```bash
go install github.com/vek-servicos/bitbucket-runner@latest
```

### Docker
```bash
docker run --rm -v $(pwd):/workspace ghcr.io/vek-servicos/bitbucket-runner:latest
```

### Local Build
```bash
make build
```

## Usage

### List available pipelines
```bash
bitbucket-runner list
```

### Run default pipeline
```bash
bitbucket-runner run
```

### Run specific pipeline
```bash
bitbucket-runner run --pipeline=custom-build
```

## Development

### Prerequisites
- Go 1.21.5+
- Docker (for step execution)

### Building
```bash
make build
```

### Testing
```bash
make test
```

### Integration Testing
```bash
make test-integration
```

## Project Structure

```
bitbucket-runner/
├── cmd/                 # CLI commands
├── internal/           # Private application code
│   ├── models/        # Data structures
│   └── parser/        # YAML parsing logic
├── docs/              # Documentation
├── scripts/           # Build and deployment scripts
└── testdata/          # Test configurations
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[Add your license here]

## Supported Projects

This tool is used by:
- vks-jee-auth-api (Java/Spring Boot)
- vks-jss-upix-api (Java/Quarkus)
- [Add other projects as they adopt it]