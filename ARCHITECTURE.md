ARCHITECTURE.md: Anyisland
Anyisland is an AI-powered, platform-agnostic, and decentralized package manager. It is designed to transform any operating system into a sovereign development environment by intelligently ingesting tools from source or binaries, managing them through an AI-driven daemon, and synchronizing state via Git.
1. System Overview
Anyisland operates as a distributed system composed of a CLI, a background Daemon, and an AI Engine. It treats the host OS as a generic runtime, installing all tools and configurations into a user-space "Island" directory.
2. Core Components
A. The CLI (anyisland)
The primary interface for user interaction. It is a stateless binary responsible for:
 * Ingestion: Accepting GitHub URLs and triggering the AI analysis flow.
 * Manual Management: Standard install, update, and remove commands.
 * Environment Injection: Managing the user's PATH and shell aliases across different platforms (Windows, macOS, Linux).
B. The Island Daemon (anyislandd)
A low-resource background process that acts as the "brain" of the machine.
 * Discovery Server: Listens for UDP heartbeats on port 1995 from "Anyisland-aware" tools.
 * Registry Manager: Maintains a local database (island.db) of all registered tools, their versions, and their GitHub sources.
 * Auto-Updater: Periodically checks for updates via the AI Engine and performs background fetches/builds.
C. The AI Synthesizer (Agent)
A module that interfaces with LLMs (local via Ollama or remote via Gemini/OpenAI).
 * Documentation Parser: Reads READMEs/Manifests to determine build steps for unknown repositories.
 * Platform Translator: Converts generic build instructions into platform-specific commands (e.g., mapping libssl to OpenSSL on Windows).
 * Update Summarizer: Generates human-readable changelogs from git commits.
 * Privacy Firewall: Scans shell history for secrets (API keys, passwords, PII) and redacts them before synchronization.

3. Data Architecture
Directory Structure
~/.anyisland/
├── bin/                # Platform-specific binaries (added to PATH)
├── data/               # Git-synced configurations and YAML manifests
├── cache/              # Temporary source code clones and build artifacts
├── history/            # Encrypted and redacted shell history logs
├── island.db           # Local registry of registered/discovered tools
└── secrets.enc         # Encrypted environment variables (Age/SOPS)

4. E2EE Shell History Sync
Anyisland captures shell history across sessions.
 * Capture: A shell hook sends commands to Anyisland.
 * Redaction: The AI Synthesizer identifies sensitive data (e.g., export AWS_SECRET_ACCESS_KEY=...) and replaces it with placeholders.
 * Encryption: Commands are encrypted using the master key in `secrets.enc`.
 * Sync: Encrypted files are committed to the `data/` directory for Git-based synchronization.
 * Recovery: Users can search history across machines, with sensitive data remaining redacted or locally decrypted.

5. Execution Logic (The Ingestion Flow)

 * Input: User provides a GitHub URL.
 * Inspection: Anyisland fetches the file tree.
   * Path A: If anyisland.yaml exists, follow explicit instructions.
   * Path B: If not, pass README.md and file list to AI Synthesizer.
 * Plan Generation: AI returns a JSON "Build Plan" (e.g., go build -o anyisland).
 * Verification: User approves the AI-generated plan.
 * Provisioning: Anyisland executes the build/download, moves the binary to ~/.anyisland/bin, and registers it in island.db.
5. Platform Abstraction Layer (PAL)
To remain platform-independent, Anyisland uses a System interface:
 * FileProvider: Handles symlinks and directory creation (ln -s vs. mklink).
 * EnvProvider: Handles persistent environment variable injection (.bashrc vs. setx).
 * BuildProvider: Detects local compilers (Go, Rust, CC) to determine if "Source Ingestion" is possible.
6. Security Model
 * User-Space Isolation: No sudo or admin privileges required.
 * Plan Review: AI-generated scripts are never executed without an explicit user "Yes."
 * Secret Management: Master-key-based decryption of sensitive data within the Git-synced directory.
