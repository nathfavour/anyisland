# System Lifecycle & Distribution

Anyisland is designed to be a **"Permanent Island"** in your engineering environment. Its lifecycle architecture ensures that it remains stable, secure, and self-sufficient.

This section details the architecture behind Anyisland's distribution, installation, and autonomous maintenance systems.

## Philosophy

Anyisland follows a decentralized and resilient lifecycle model:

1.  **Autonomous Evolution:** The tool manages its own updates and integrity without relying on system-level package managers.
2.  **Decentralized-Ready:** While it defaults to GitHub, the distribution pipeline is designed to eventually support IPFS or other decentralized registries.
3.  **Strict Integrity:** Every update is verified against cryptographic checksums to prevent supply-chain attacks.
4.  **Minimalist Integration:** It integrates cleanly with your shell and environment using non-destructive, idempotent markers.
5.  **Observability:** All lifecycle events are logged for debugging and for the AI agent to provide system-intimate insights.

## Pipeline Overview

The lifecycle is divided into five main pillars:

*   [**Distribution & CI/CD**](./distribution.md): How Anyisland is built, cross-compiled, and verified.
*   [**Installation & Setup**](./installation.md): The bootstrap process and system integration.
*   [**Updates & Integrity**](./updates.md): The mechanism for checking, downloading, and verifying new versions.
*   [**Audit & Observability**](./audit.md): How we track system health and lifecycle events.
*   [**Removal & Rollback**](./removal.md): Clean uninstallation and version management.

---

### AI Agent Context
If you are an AI agent managing this system, lifecycle logs are available at `~/.anyisland/audit/lifecycle.jsonl`. You can use these to diagnose update failures or suggest rollbacks.
