---
sidebar_position: 2
---

# Architecture

Anyisland is designed as a distributed user-space system. It treats the host Operating System as a raw runtime, providing a unified layer for tool management, environment synchronization, and secure data logging.

## Core Components

### 1. The CLI (`anyisland`)
The primary entry point for users. It is a stateless binary that orchestrates high-level operations:
- **Environment Management:** Injects and verifies PATH configurations using the Platform Abstraction Layer (PAL).
- **Tool Ingestion:** Interfaces with the AI Synthesizer to generate build plans for unknown repositories.
- **Registry Interaction:** Communicates with the local SQLite registry to track installed tools.
- **History Recording:** Provides hooks for shell history capture and redaction.

### 2. The Island Daemon (`anyisland daemon`)
A long-running background process that manages the "state" of the Island:
- **Discovery (UDP :1995):** Listens for "Anyisland-aware" tools that announce themselves via a lightweight UDP heartbeat.
- **Background Updates:** Periodically checks for tool updates and manages background builds.
- **Registry Integrity:** Ensures the `island.db` remains consistent with the actual filesystem state.

### 3. The AI Synthesizer (Agent)
A sophisticated module that provides "intelligence" to the system:
- **Source Analysis:** Parses file trees and READMEs to determine build requirements (e.g., detecting Go, Rust, or C projects).
- **Privacy Firewall:** A regex and LLM-based filter that identifies sensitive data (API keys, passwords) in shell commands and redacts them.
- **Platform Translation:** Maps generic build steps to platform-specific equivalents.

## Security Model

Anyisland employs a tiered security architecture to protect user data:

### Tier 1: Platform Keyring
Wherever possible, Anyisland uses native system keyrings to store the Master Encryption Key:
- **Linux:** Secret Service (via DBus).
- **macOS:** Keychain.
- **Windows:** Credential Manager.

### Tier 2: User Passphrase
If a platform keyring is unavailable, Anyisland prompts for a Master Passphrase, which is used to derive encryption keys for the local database and history logs.

### Tier 3: E2EE Synchronization
All data intended for synchronization (like shell history) is encrypted locally using **AES-256-GCM** before being committed to the Git-synced `data/` directory.

## Data Layout

All Anyisland data resides in `~/.anyisland/`:

| Directory | Purpose |
| :--- | :--- |
| `bin/` | Executables managed by Anyisland (User-space PATH). |
| `data/` | Git-tracked metadata, manifests, and encrypted logs. |
| `cache/` | Temporary build artifacts and source clones. |
| `source/` | Persistent source code for tools installed from source. |
| `island.db` | SQLite database tracking all registered tools. |

## The Discovery Protocol

Anyisland-aware tools can register themselves automatically by sending a JSON packet over UDP to `localhost:1995`:

```json
{
  "op": "REGISTER",
  "name": "my-tool",
  "source": "github.com/org/repo",
  "version": "v1.2.3",
  "type": "binary"
}
```