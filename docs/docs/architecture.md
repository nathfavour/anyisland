---
sidebar_position: 2
---

# Architecture

Anyisland operates as a distributed system composed of a CLI, a background Daemon, and an AI Engine. It treats the host OS as a generic runtime, installing all tools and configurations into a user-space "Island" directory.

## Core Components

### The CLI (`anyisland`)
The primary interface for user interaction. It is a stateless binary responsible for:
- **Ingestion:** Accepting GitHub URLs and triggering the AI analysis flow.
- **Manual Management:** Standard install, update, and remove commands.
- **Environment Injection:** Managing the user's PATH and shell aliases across different platforms (Windows, macOS, Linux).

### The Island Daemon (`anyislandd`)
A low-resource background process that acts as the "brain" of the machine.
- **Discovery Server:** Listens for UDP heartbeats on port 1995 from "Anyisland-aware" tools.
- **Registry Manager:** Maintains a local database (`island.db`) of all registered tools, their versions, and their GitHub sources.
- **Auto-Updater:** Periodically checks for updates via the AI Engine and performs background fetches/builds.

### The AI Synthesizer (Agent)
A module that interfaces with LLMs (local via Ollama or remote via Gemini/OpenAI).
- **Documentation Parser:** Reads READMEs/Manifests to determine build steps for unknown repositories.
- **Platform Translator:** Converts generic build instructions into platform-specific commands.
- **Update Summarizer:** Generates human-readable changelogs from git commits.
- **Privacy Firewall:** Scans shell history for secrets and redacts them before synchronization.

## Data Architecture

### Directory Structure
- `~/.local/bin/`: User-local binaries (system PATH).
- `~/.anyisland/`:
    - `data/`: Git-synced configurations and JSON manifests.
    - `cache/`: Temporary source code clones and build artifacts.
    - `history/`: Encrypted and redacted shell history logs.
    - `island.db`: Local registry of registered/discovered tools.
    - `secrets.enc`: Encrypted environment variables.

## Execution Logic (The Ingestion Flow)

1. **Input:** User provides a GitHub URL.
2. **Inspection:** Anyisland fetches the file tree.
    - **Path A:** If `anyisland.json` exists, follow explicit instructions.
    - **Path B:** If not, pass README.md and file list to AI Synthesizer.
3. **Plan Generation:** AI returns a JSON "Build Plan" (e.g., `go build -o anyisland`).
4. **Verification:** User approves the AI-generated plan.
5. **Provisioning:** Anyisland executes the build/download, moves the binary to `~/.anyisland/bin`, and registers it in `island.db`.
