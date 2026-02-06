---
sidebar_position: 3
---

# Distribution & Integration

Anyisland allows developers to distribute their tools effortlessly and integrate them deeply into the user's sovereign environment.

## 1. Creating a Manifest (`anyisland.json`)

The manifest is the source of truth for Anyisland. Placing this file in your repository's root allows Anyisland to skip AI analysis and follow your explicit instructions.

See the [**Manifest Specification**](./manifest-spec.md) for a detailed breakdown of all available fields and best practices.

### Basic Schema

```json
{
  "name": "tool-name",
  "version": "1.0.0",
  "build": {
    "steps": [
      "go build -o mytool ."
    ],
    "bin": "mytool",
    "install_dir": "/custom/path"
  }
}
```

#### Field Reference
- **`name`**: The unique identifier for your tool.
- **`version`**: The current semantic version.
- **`build.steps`**: An array of shell commands executed sequentially in the source root.
- **`build.bin`**: The relative path to the resulting executable after build steps finish.
- **`build.install_dir`** *(Optional)*: An absolute path if you want to override the default `~/.anyisland/bin` location.

---

## 2. Anyisland-Aware Tools

Tools can integrate with the Anyisland ecosystem to enable features like auto-discovery and state management.

### Auto-Registration
A tool can register itself with the local Anyisland daemon (`anyislandd`) by sending a UDP packet to port `1995`.

**Example (Go):**
```go
import (
    "net"
    "encoding/json"
)

func Register() {
    packet := map[string]string{
        "op":      "REGISTER",
        "name":    "my-tool",
        "source":  "github.com/me/my-tool",
        "version": "v1.0.0",
        "type":    "binary",
    }
    data, _ := json.Marshal(packet)
    conn, _ := net.Dial("udp", "localhost:1995")
    conn.Write(data)
}
```

---

## 3. Contributing Official Packages

Anyisland maintains a set of "Official Packages" for common utilities. These manifests are stored in the core repository under `packages/official/`.

To contribute:
1. Fork the Anyisland repository.
2. Create `packages/official/<name>/anyisland.json`.
3. Submit a Pull Request.

Official packages are prioritized by the `anyisland install` command and are verified for platform compatibility.