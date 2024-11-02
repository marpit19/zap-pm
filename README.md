# Zap Package Manager

A lightweight, fast JavaScript package manager built with Go. Zap focuses on speed and simplicity, making it perfect for quick projects and learning about package management.

## Current Status (2 Days Progress)

Zap is currently in early development (Day 2 complete). It implements core package management features with a focus on speed and reliability.

### Features Comparison

| Feature | Zap (Current) | npm | yarn |
|---------|--------------|-----|------|
| Initialize Project | ✅ | ✅ | ✅ |
| Package Info | ✅ | ✅ | ✅ |
| Package Download | ✅ | ✅ | ✅ |
| Dependencies Download | ✅ | ✅ | ✅ |
| Cache System | ✅ | ✅ | ✅ |
| Progress Bar | ✅ | ✅ | ✅ |
| Version Resolution | ✅ | ✅ | ✅ |
| Lock File | 🚧 | ✅ | ✅ |
| Workspaces | ❌ | ✅ | ✅ |
| Scripts | ❌ | ✅ | ✅ |
| Plugins | ❌ | ✅ | ✅ |

✅ = Implemented, 🚧 = In Progress, ❌ = Not Yet Implemented

## Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Build from Source
```bash
# Clone the repository
git clone https://github.com/marpit19/zap-pm
cd zap-pm

# Get dependencies
go mod tidy

# Build
go build -o zap cmd/zap/main.go
```

## Usage

### Initialize a New Project
```bash
# Create a new directory
mkdir my-project
cd my-project

# Initialize package.json
./zap init
```

### Get Package Information
```bash
# View package details
./zap info express
```

### Download Packages
```bash
# Download latest version
./zap download express

# Download specific version
./zap download express@4.17.1

# Download with dependencies
./zap download express --with-dependencies
```

### Verify Package Cache
```bash
# Verify package integrity
./zap verify express@4.17.1
```

## Features in Detail

### 1. Package Initialization
- Creates and validates package.json
- Supports standard npm package.json format
- Interactive creation process

### 2. Registry Integration
- Fast package metadata fetching
- Smart version resolution
- Support for semver ranges
- Automatic retry on failures

### 3. Download System
- Concurrent package downloads
- Progress visualization
- Download speed tracking
- Checksum verification
- Intelligent caching

### 4. Cache Management
- Local cache in ~/.zap/cache
- Automatic cache validation
- Cache cleanup
- Checksum verification

## Project Structure
```
zap-pm/
├── cmd/
│   └── zap/
│       └── main.go           # Entry point
├── internal/
│   ├── cli/                  # CLI implementation
│   ├── registry/             # NPM registry client
│   ├── downloader/           # Download management
│   ├── parser/              # package.json parsing
│   ├── logger/              # Logging system
│   └── errors/              # Error handling
└── README.md
```

## Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Known Limitations
- No lock file support yet
- Single registry only (npm)
- No workspace support
- No script execution
- No authentication support
- Basic retry logic
- No proxy support

## Coming Soon
- Lock file implementation
- Dependency graph resolution
- Circular dependency detection
- Advanced version conflict resolution
- Node modules tree optimization

## Contributing
We welcome contributions! This is a fun weekend project to learn about package management. Feel free to submit issues and pull requests.

## License
MIT License - see LICENSE file for details.

## Author
[marpit19](https://github.com/marpit19)
