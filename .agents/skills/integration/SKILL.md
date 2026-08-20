---
name: anyisland-integration
description: Complete guide and reference for integrating any tool, binary, or repository with Anyisland. Covers single-command bootstrap hooks, manifest specification (anyisland.json), install adapter scripts, and live pulse/daemon auto-registration.
---

# Anyisland Integration Skill

Use this skill when integrating any repository, service, or CLI tool with the **Anyisland** decentralized ecosystem.

---

## 1. Core Architecture Pattern

Anyisland enables a frictionless **one-command distribution model** for tools:

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
```

When an end-user runs this command:
1. `install.sh` downloads/builds Anyisland in `~/.local/bin` and registers it in `~/.anyisland/island.db` if not already present.
2. `install.sh` immediately calls `anyisland install <owner/repo>`.
3. Anyisland reads the tool's `anyisland.json` manifest, prepares dependencies, executes the build pipeline, and symlinks the binary into the user's `PATH`.

---

## 2. Manifest Schema (`anyisland.json`)

Every integrating tool must include an `anyisland.json` at its repository root:

```json
{
  "name": "your-tool-name",
  "version": "1.0.0",
  "description": "Short summary of what the tool does.",
  "repository": "github.com/owner/your-tool-name",
  "source_dir": "~/.your-tool-name/source",
  "install_dir": "~/.local/bin",
  "build": {
    "steps": [
      "bash install.sh"
    ],
    "bin": "your-tool-name",
    "toolchain": "go"
  },
  "runtime": {
    "dependencies": [
      "libtesseract-dev",
      "libleptonica-dev"
    ],
    "daemon": false,
    "pulse": true
  }
}
```

### Manifest Fields
- **`name`**: Tool binary and registry name (must match binary output).
- **`version`**: Semantic version string (e.g. `1.0.0`).
- **`build.steps`**: Sequential shell commands to compile or bundle the application.
- **`build.bin`**: Name or relative path of the produced executable.
- **`build.toolchain`**: Optional hint (`go`, `rust`, `node`, `python`, `flutter`, `c`).
- **`runtime.dependencies`**: OS packages or libraries required at runtime.
- **`runtime.pulse`**: Set `true` if the tool participates in Anyisland health/heartbeat pulses.

---

## 3. Tool `install.sh` Adapter Script

Include an `install.sh` in your repository root that handles both direct standalone execution and Anyisland invocation:

```bash
#!/bin/bash
set -e

# 1. Fallback: If run directly without Anyisland, delegate to Anyisland bootstrap
if ! command -v anyisland &> /dev/null; then
    echo "🏝️ Bootstrapping via Anyisland..."
    curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
    exit 0
fi

# 2. Host Dependency Verification (Linux/macOS)
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

if [ "$OS" = "linux" ]; then
    if command -v apt-get >/dev/null 2>&1; then
        # Check and install system packages if needed
        echo "Verifying Linux libraries..."
    fi
elif [ "$OS" = "darwin" ]; then
    echo "Verifying macOS Homebrew dependencies..."
fi

# 3. Build Binary
echo "Building binary..."
go build -o your-tool-name ./cmd/your-tool-name
```

---

## 4. Live Daemon Auto-Registration (UDP :1995)

Tools can broadcast their presence to the Anyisland background daemon on startup:

### Go Example
```go
package main

import (
    "encoding/json"
    "net"
)

func RegisterWithIsland(name, version, repo string) {
    packet := map[string]string{
        "op":      "REGISTER",
        "name":    name,
        "source":  repo,
        "version": version,
        "type":    "binary",
    }
    data, err := json.Marshal(packet)
    if err != nil {
        return
    }
    conn, err := net.Dial("udp", "127.0.0.1:1995")
    if err == nil {
        defer conn.Close()
        _, _ = conn.Write(data)
    }
}
```

---

## 5. Standard README Installation Section

Document installation in your tool's `README.md` as follows:

```markdown
## 📦 Installation

Install instantly with a single command (powered by [Anyisland](https://github.com/nathfavour/anyisland)):

\`\`\`bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
\`\`\`

*Or if you already have Anyisland installed:*
\`\`\`bash
anyisland install <owner/repo>
\`\`\`
```
