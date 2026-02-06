---
sidebar_position: 1
---

# Introduction

**Anyisland** is an AI-powered, platform-agnostic, and decentralized package manager. It is designed to transform any operating system into a sovereign development environment by intelligently ingesting tools from source or binaries, managing them through an AI-driven daemon, and synchronizing state via Git.

Unlike traditional package managers that rely on centralized repositories, Anyisland treats the entire GitHub ecosystem as its repository, using AI to understand and provision tools on the fly.

## Why Anyisland?

- **🏝️ Sovereign Environment:** Your tools and configurations live in a user-space "Island" (`~/.anyisland`), completely isolated from system-level dependencies.
- **🤖 AI-Driven Provisioning:** Point Anyisland at a GitHub repository, and its AI Synthesizer will analyze the source code, README, and project structure to generate a custom build and installation plan for your specific platform.
- **🌍 Platform Agnostic:** Built on a robust Platform Abstraction Layer (PAL), Anyisland provides a unified experience across Linux, macOS, and Windows.
- **🔐 Privacy First:** Features an E2EE (End-to-End Encrypted) shell history sync with an AI-powered "Privacy Firewall" that redacts secrets and PII before they ever leave your machine.
- **🔄 Decentralized Sync:** Your tool registry, configurations, and encrypted history are managed via Git, allowing you to synchronize your entire development environment across machines effortlessly.

## Core Concepts

### The Island (`~/.anyisland`)
The "Island" is your personal workspace. Everything Anyisland manages—from binaries and source code to encrypted history—is stored here. This ensures that your development environment is portable and doesn't clutter your host system.

### Ingestion vs. Installation
- **Installation:** Refers to deploying pre-defined "Official Packages" or tools that already contain an `anyisland.json` manifest.
- **Ingestion:** The process of taking a generic source repository, analyzing it via AI, and creating a specialized build plan to turn it into an installed tool.

## Quick Start

### 1. Installation
Install Anyisland using our official bootstrap script:

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash
```

### 2. Setup
Initialize your local Island and configure your PATH:

```bash
anyisland setup
```

### 3. Ingest your first tool
Transform a repository into a local tool:

```bash
anyisland ingest https://github.com/nathfavour/anyisland
```

### 4. Manage your environment
List installed tools and check their status:

```bash
anyisland list
```
