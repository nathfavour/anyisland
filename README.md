# Anyisland

AI-powered, platform-agnostic, and decentralized package manager.

## Documentation

Full documentation is available at [nathfavour.github.io/anyisland/](https://nathfavour.github.io/anyisland/).

- **[Architecture](https://nathfavour.github.io/anyisland/docs/architecture)**
- **[System Lifecycle](https://nathfavour.github.io/anyisland/docs/lifecycle)**
- **[CLI Reference](https://nathfavour.github.io/anyisland/docs/cli)**
- **[Security & Privacy](https://nathfavour.github.io/anyisland/docs/security)**
- **[Distribution Guide](https://nathfavour.github.io/anyisland/docs/distribution)**

## Quick Start

### 1. Install, Update & Uninstall
```bash
# Install Anyisland
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash

# Update (Anyisland updates automatically on every run, but you can force it)
anyisland update anyisland

# Uninstall Anyisland
anyisland uninstall
```

### 2. Setup
Initialize your local Island and configure your PATH.
```bash
anyisland setup
```

### 3. Ingest a Tool
Transform any GitHub repository into an installed tool via AI analysis.
```bash
anyisland ingest github.com/user/repo
```

### 4. List Tools
See what's installed and managed.
```bash
anyisland list
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
```
