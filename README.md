# Anyisland

AI-powered, platform-agnostic, and decentralized package manager.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash
```

## Quick Start

### 1. Initialize
Set up the environment and inject Anyisland into your PATH.
```bash
go run ./cmd/anyisland/main.go init
```

### 2. Start the Daemon
The daemon listens for "Anyisland-aware" tools and manages the registry.
```bash
go run ./cmd/anyislandd/main.go
```

### 3. Ingest a Tool
Transform any GitHub repository into an installed tool via AI analysis.
```bash
go run ./cmd/anyisland/main.go ingest github.com/user/repo
```

### 4. List Tools
See what's installed and managed.
```bash
go run ./cmd/anyisland/main.go list
```

## Architecture
- **PAL (Platform Abstraction Layer):** Unified interface for Linux, macOS, and Windows.
- **Registry:** SQLite database (`~/.anyisland/island.db`) tracking all tools.
- **Daemon:** Background UDP server on port 1995 for tool discovery.
- **Agent:** AI Synthesizer that generates build plans from source code analysis.

## Development
```bash
# Build CLI
go build -o anyisland ./cmd/anyisland

# Build Daemon
go build -o anyislandd ./cmd/anyislandd
```
