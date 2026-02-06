# Updates & Integrity

Anyisland maintains its own health through a resilient update system.

## Update Tracks

Users can choose between different update tracks:

1.  **Stable (Default):** Tracks tagged releases.
2.  **Edge:** Tracks the `release` branch for the latest verified features.
3.  **Source:** Automatically builds from source using the local Go toolchain (requires `go` to be installed).

## The Update Flow

When `anyisland update` is executed:

1.  **Check:** The tool identifies the latest version for the current track.
2.  **Verify:** It downloads the `checksums.txt` and verifies the target binary's integrity.
3.  **Shadow Download:** The new binary is downloaded to `~/.anyisland/cache/updates/`.
4.  **Swap:** The current binary in `~/.local/bin/` is renamed to `*.old` (providing a manual rollback path), and the new binary is moved into place.

## Daemon Hot-Swapping

The `anyislandd` daemon supports hot-swapping to ensure continuous background operations:

1.  **State Save:** The daemon serializes its current registry state and active tasks.
2.  **Exec:** It uses `syscall.Exec` to replace its process image with the new version.
3.  **Restore:** The new daemon process loads the serialized state and resumes operations seamlessly.

## Self-Healing

If Anyisland detects that its binary is corrupted (checksum mismatch on startup) or if the daemon fails to start repeatedly, it enters **Self-Healing Mode**:
*   It attempts to restore the `*.old` binary.
*   If that fails, it re-runs the bootstrap logic to fetch a fresh binary from the official distribution point.
