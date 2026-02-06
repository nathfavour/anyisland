---
sidebar_position: 5
---

# CLI Reference

The `anyisland` command is the primary tool for managing your sovereign environment.

## General Commands

### `anyisland setup`
Initializes the Anyisland environment.
- Creates `~/.anyisland` directory structure.
- Initializes the local SQLite registry.
- Generates or retrieves the Master Encryption Key.
- Injects `~/.anyisland/bin` into your shell's `PATH`.

### `anyisland init`
Initializes the current directory as an Anyisland-compatible project by creating a template `anyisland.json`.

---

## Tool Management

### `anyisland install <url>`
Installs a tool that already has an `anyisland.json` manifest.
- **Example:** `anyisland install https://github.com/nathfavour/anyisland`

### `anyisland ingest <url>`
Ingests a tool from a repository *without* a manifest.
- Triggers AI analysis.
- Proposes a build plan for your approval.
- Builds and installs the tool.

### `anyisland list`
Lists all tools currently registered in your local Island.

### `anyisland update [tool]`
Updates a specific tool or all tools if no argument is provided.
- `anyisland update anyisland`: Updates the Anyisland CLI itself.

---

## History Management

### `anyisland history show`
Displays your synchronized and decrypted shell history.

### `anyisland history record "[command]"`
Manually records a command into the encrypted history. Usually called automatically by shell hooks.

---

## Global Flags
- `--source, -s`: Override the source URL or local path for install/ingest/update commands.
- `--help, -h`: Show help for any command.
- `--version, -v`: Display the current version of Anyisland.
