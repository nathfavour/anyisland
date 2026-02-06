# Distribution & Integrity

Anyisland is distributed as a set of statically-linked Go binaries (`anyisland` and `anyislandd`), ensuring maximum portability across environments.

## CI/CD Pipeline

Our distribution pipeline uses GitHub Actions to automate the release process. On every tagged release (`v*`) or push to the `release` branch, the following occurs:

1.  **Cross-Compilation:** Binaries are built for:
    *   `linux/amd64`, `linux/arm64`
    *   `darwin/amd64`, `darwin/arm64` (macOS)
    *   `windows/amd64`, `windows/arm64`
2.  **Metadata Injection:** Version, Git Commit, and Build Time are baked into the binaries using `-ldflags`.
3.  **Checksum Generation:** A `checksums.txt` file is generated using SHA-256 for all artifacts.
4.  **Decentralized Sync:** (Planned) Hashes and binaries are announced to the Anyisland registry for decentralized discovery.

## Strict Integrity Policy

To protect against tampered binaries, Anyisland enforces a strict integrity check:

*   **Verification:** Before any update or installation, the tool downloads the signed `checksums.txt`.
*   **Validation:** The downloaded binary's SHA-256 hash must match the entry in `checksums.txt`.
*   **Fail-Safe:** If verification fails, the update is immediately aborted, and the current stable binary is preserved.

## Discovery Mechanism

Anyisland uses a multi-layered discovery process to identify updates:

1.  **Registry Discovery:** Queries the decentralized Anyisland registry (or fallback GitHub repository) for the latest release metadata.
2.  **Commit Tracking:** Compares the embedded `Commit` hash of the running binary with the remote target to identify "Stable-Edge" updates that may not have a new semantic version.
3.  **Background Polling:** The `anyislandd` daemon periodically checks for updates in the background and notifies the user via the CLI.
