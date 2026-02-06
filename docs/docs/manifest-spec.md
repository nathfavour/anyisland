---
sidebar_position: 4
---

# Manifest Specification (`anyisland.json`)

The `anyisland.json` file is the primary configuration for Anyisland-aware tools. It provides explicit instructions to the Anyisland ingestion engine, allowing it to bypass AI-based build plan generation and ensure a predictable installation experience.

## Example Manifest

```json
{
  "name": "my-tool",
  "version": "1.2.0",
  "description": "A high-performance CLI utility for data processing.",
  "repository": "github.com/username/my-tool",
  "build": {
    "steps": [
      "make build",
      "mv bin/my-tool ."
    ],
    "bin": "my-tool",
    "install_dir": "/usr/local/bin"
  },
  "runtime": {
    "dependencies": ["curl", "git"],
    "daemon": false,
    "pulse": true
  }
}
```

## Field Reference

### Root Fields

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | `string` | **Yes** | The unique identifier for the tool. This determines the command name used in the registry. |
| `version` | `string` | **Yes** | Semantic version (e.g., `1.0.0`). Used for update tracking and rollbacks. |
| `description` | `string` | No | A short summary of what the tool does. |
| `repository` | `string` | No | The canonical URL of the source code repository. |
| `build` | `object` | **Yes** | Instructions on how to compile or prepare the tool's binary. |
| `runtime` | `object` | No | Configuration for how the tool behaves on the host system. |

---

### The `runtime` Object

| Field | Type | Description |
| :--- | :--- | :--- |
| `dependencies` | `array[string]` | System packages required for the tool to function. |
| `daemon` | `boolean` | If true, Anyisland will treat this as a long-running service. |
| `pulse` | `boolean` | If true, the tool is **Anyisland Pulse Aware** and can receive OTA update notifications from the local daemon. |

---

## Anyisland Pulse (OTA)

"The Pulse" is a low-energy OTA communication channel between `anyislandd` and installed tools. 

### How it Works
1. **Centralized Polling**: `anyislandd` fetches update metadata for all registered tools in the background.
2. **Local IPC**: Tools can query the local Unix Domain Socket at `~/.anyisland/anyisland.sock` to check for updates without touching the network.
3. **Push Notifications**: Tools with `pulse: true` can subscribe to the socket to receive immediate push notifications when an update is available.

### Why use Pulse?
- **Zero Bandwidth**: Your tool doesn't need its own update-checking logic.
- **Battery Efficient**: Consolidates network requests into a single background process.
- **Privacy**: No independent tracking of tool usage by external servers.

---

### The `build` Object

The `build` object defines the lifecycle of the tool from source to executable.

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `steps` | `array[string]` | **Yes** | A sequence of shell commands to run in the root of the source directory. |
| `bin` | `string` | **Yes** | The path to the produced binary **relative to the repository root**. Supports glob patterns (e.g., `dist/*_linux_amd64/mytool`). |
| `install_dir` | `string` | No | An optional override for the installation path. Defaults to `~/.anyisland/bin/`. |
| `toolchain` | `string` | No | The required toolchain (e.g., `go`, `rust`). Anyisland will verify this exists before building. |

---

## First-Class Go Support

Anyisland treats Go projects as primary citizens. When a `go.mod` file is detected:

1. **Automatic Detection**: Anyisland automatically identifies the project as a Go toolchain project.
2. **Build Optimization**: It defaults to optimized build steps like `go build -v`.
3. **GoReleaser Integration**: If a `.goreleaser.yaml` is found, Anyisland can use `goreleaser build --snapshot --single-target` to produce the binary, ensuring that all metadata and build flags defined by the author are respected.
4. **Glob Pattern Matching**: The `bin` field supports glob patterns to easily locate binaries in complex output structures like those produced by GoReleaser (e.g., `dist/mytool_*/mytool`).

---

#### How Anyisland Handles Installation
1. **Build**: Runs `steps` in the source directory.
2. **Locate**: Finds the binary at the path specified in `bin`.
3. **Deploy**: Copies the binary to `install_dir` (or the default system path).

---

## AI Ingestion & Overrides

When Anyisland ingests a repository without an `anyisland.json` file, it uses the **AI Synthesizer** to generate a build plan by analyzing:
1.  The folder structure (e.g., presence of `go.mod`, `package.json`, `Makefile`).
2.  The `README.md` content.
3.  The source code file names.

**By providing an `anyisland.json`, you explicitly override this behavior.** This is highly recommended for complex projects or those requiring specific build flags.

## Best Practices

1.  **Statically Link Binaries:** Whenever possible, ensure your build steps produce a statically linked binary (e.g., `CGO_ENABLED=0 go build`) to maximize portability across user systems.
2.  **Relative Binary Paths:** The `bin` field should always be relative to the source root.
3.  **Minimal Steps:** Keep build steps focused and minimal. Use existing build tools (like `make` or `go build`) rather than complex shell scripts.
4.  **Version Consistency:** Ensure the `version` field in the manifest matches your Git tags to enable seamless updates via the `anyisland update` command.
