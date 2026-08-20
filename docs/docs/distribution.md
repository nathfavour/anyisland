---
sidebar_position: 3
---

# Distribution & Integration

Anyisland provides a zero-friction, platform-agnostic distribution mechanism for modern CLI tools and sovereign software. Tools integrating with Anyisland can provide users with an instant single-command install experience without requiring separate package manager setup.

---

## 1. Universal One-Command Bootstrap (Recommended Default)

The default and cleanest distribution method is the **Anyisland Universal Bootstrap Hook**. This allows end users to install your tool directly using a single curl command. The script bootstraps Anyisland (if not already present) and immediately delegates to building/installing your tool.

### User Install Snippet

Add this to your project's `README.md`:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
```

*Example for `nathfavour/threader`:*
```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- nathfavour/threader
```

If the user already has Anyisland installed, they can also install your tool via:
```bash
anyisland install <owner/repo>
```

---

## 2. Defining the Manifest (`anyisland.json`)

Placing an `anyisland.json` in your repository's root tells Anyisland how to resolve runtime dependencies and build your project deterministically.

See the [**Manifest Specification**](./manifest-spec.md) for full field details.

### Standard Go / Rust / C / Native Schema

```json
{
  "name": "mytool",
  "version": "1.0.0",
  "description": "High performance marketing engine.",
  "repository": "github.com/owner/mytool",
  "source_dir": "~/.mytool/source",
  "install_dir": "~/.local/bin",
  "build": {
    "steps": [
      "bash install.sh"
    ],
    "bin": "mytool",
    "toolchain": "go"
  },
  "runtime": {
    "dependencies": [
      "libtesseract-dev",
      "libleptonica-dev"
    ],
    "pulse": true
  }
}
```

---

## 3. Tool `install.sh` Adapter Pattern

To give users maximum flexibility, your repository's local `install.sh` can act as an adapter that bootstraps through Anyisland when run standalone, and builds the binary when called inside an Anyisland cycle:

```bash
#!/bin/bash
set -e

# If run directly outside Anyisland, bootstrap via Anyisland
if ! command -v anyisland &> /dev/null; then
    echo "🏝️ Bootstrapping via Anyisland..."
    curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
    exit 0
fi

# Build logic when invoked by Anyisland
echo "Building binary..."
go build -o mytool ./cmd/mytool
```

---

## 4. Anyisland-Aware Tools (Pulse & Auto-Registration)

Tools can integrate with the local Anyisland daemon (`anyisland daemon`) to enable automatic discovery, telemetry, and live registration.

### Auto-Registration via UDP
A tool can register itself with the local Anyisland daemon by sending a UDP packet to port `1995`:

```go
import (
    "encoding/json"
    "net"
)

func RegisterWithIsland() {
    packet := map[string]string{
        "op":      "REGISTER",
        "name":    "mytool",
        "source":  "github.com/owner/mytool",
        "version": "v1.0.0",
        "type":    "binary",
    }
    data, _ := json.Marshal(packet)
    conn, err := net.Dial("udp", "127.0.0.1:1995")
    if err == nil {
        defer conn.Close()
        conn.Write(data)
    }
}
```

---

## 5. Contributing Official Packages

Anyisland maintains a set of "Official Packages" for common utilities in `packages/official/`.

To contribute:
1. Fork the Anyisland repository.
2. Create `packages/official/<name>/anyisland.json`.
3. Submit a Pull Request.
