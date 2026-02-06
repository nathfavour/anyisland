---
sidebar_position: 4
---

# Security & Privacy

Anyisland is built on the principle of **User Sovereignty**. We believe your development environment should be private, secure, and under your total control.

## AI Privacy Firewall

One of Anyisland's most unique features is the **Privacy Firewall**. When you enable shell history synchronization, Anyisland uses its AI agent to scan every command before it is saved.

### How it Works:
1. **Detection:** The agent scans for patterns matching API keys, passwords, bearer tokens, and PII (Personally Identifiable Information).
2. **Redaction:** Sensitive parts of the command are replaced with `[REDACTED]` placeholders.
3. **Local processing:** Redaction happens *locally* before any data is encrypted or moved to the sync directory.

## E2EE Shell History

Anyisland synchronizes your shell history across all your "Islands" using End-to-End Encryption (E2EE).

- **Key Management:** Encryption keys are derived from your Master Key, which is stored in your platform's secure keyring (e.g., macOS Keychain).
- **Zero-Knowledge Sync:** Because data is encrypted locally before being committed to Git, your Git provider (e.g., GitHub) never sees your raw shell commands or configurations.
- **Searchable History:** Even though history is encrypted on disk and in sync, the `anyisland history show` command allows you to search and view your history across all machines by decrypting it on demand using your local Master Key.

## User-Space Isolation

Anyisland strictly operates within user-space.
- **No `sudo` required:** All tools are installed into `~/.anyisland`.
- **System Integrity:** Anyisland never modifies system-level binaries or configuration files (with the exception of appending to your shell RC file to manage `PATH`).
- **Sandboxed Execution (Planned):** Future versions will support running ingested tools within lightweight containers or sandboxes to prevent malicious code from accessing your host system.

## Master Key Recovery

Your Master Key is the heart of your Anyisland security.
- **Backup:** You are encouraged to back up your `secrets.enc` file and your Master Passphrase.
- **Loss of Key:** If you lose your Master Key and haven't backed up your passphrase, encrypted data (like history logs) will be permanently unrecoverable. This is by design to ensure total privacy.
