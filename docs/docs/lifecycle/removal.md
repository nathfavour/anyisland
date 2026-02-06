# Removal & Rollback

We respect your system's hygiene. Anyisland provides clear paths for reverting updates or removing the tool entirely.

## Rollbacks

If an update introduces issues, you can roll back to the previous version:

```bash
anyisland rollback
```

This command swaps the current binary with the `anyisland.old` backup created during the last update.

## Uninstallation

To remove Anyisland from your system, use the `uninstall` command:

```bash
anyisland uninstall
```

### Standard Uninstallation
1.  **Binary Removal:** Deletes `anyisland` and `anyisland daemon` from `~/.local/bin/`.
2.  **Shell Cleanup:** Removes the `# >>> anyisland initialize >>>` blocks from your shell profiles.
3.  **Data Preservation:** Keeps `~/.anyisland` (configs, registry, history) intact so you can reinstall without losing data.

### Complete Wipe
To remove all traces, including encrypted history and configurations:

```bash
anyisland uninstall --clean
```

## Manual Cleanup
If the binary is already removed, you can manually clean up by:
1. Deleting `~/.anyisland`.
2. Removing the initialization blocks from `.bashrc`, `.zshrc`, or `.config/fish/config.fish`.
