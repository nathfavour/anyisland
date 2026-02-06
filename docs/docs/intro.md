---
sidebar_position: 1
---

# Introduction

Anyisland is an AI-powered, platform-agnostic, and decentralized package manager. It is designed to transform any operating system into a sovereign development environment by intelligently ingesting tools from source or binaries, managing them through an AI-driven daemon, and synchronizing state via Git.

## Key Features

- **AI-Powered Ingestion:** Transform any GitHub repository into an installed tool via source analysis.
- **Platform Abstraction Layer (PAL):** Unified interface for Linux, macOS, and Windows.
- **Decentralized Registry:** SQLite database tracking all tools and configurations.
- **Secure Secret Management:** Tiered approach using platform keyrings and user passphrases.
- **Shell History Sync:** Encrypted and redacted shell history across machines.

## Installation

To install Anyisland, run the following command in your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash
```

## Quick Start

### 1. Initialize
Set up the environment and inject Anyisland into your PATH.
```bash
anyisland init
```

### 2. Start the Daemon
The daemon listens for "Anyisland-aware" tools and manages the registry.
```bash
anyislandd
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