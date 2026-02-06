# Installation & System Integration

Anyisland is designed to be installed without root permissions, living entirely within the user's home directory.

## The Bootstrap Process

The standard installation uses a universal bootstrap script:

```bash
curl -fsSL https://anyisland.io/install.sh | bash
```

### 1. Platform Detection
The script detects the host OS and architecture (e.g., `linux-arm64`) to select the correct binary package.

### 2. Binary Hand-off
The script downloads the specific binary and executes it with the `self-install` flag. This transfers the installation logic from the ephemeral shell script to the robust Anyisland binary itself.

## System Integration (`self-install`)

When Anyisland runs its installation routine, it performs the following:

*   **Folder Structure:** Creates `~/.anyisland` and its subdirectories (`bin`, `data`, `cache`, `source`).
*   **Self-Migration:** Copies the `anyisland` and `anyisland daemon` binaries to `~/.local/bin/`.
*   **Shell Injection:** Detects the active shell (Bash, Zsh, or Fish) and adds `~/.local/bin` to the `PATH` using idempotent markers:

```bash
# >>> anyisland initialize >>>
export PATH="$PATH:$HOME/.local/bin"
# <<< anyisland initialize <<<
```

## Security Initialization

During the first installation, Anyisland generates a unique **Master Key** used for E2EE shell history and local vault encryption. This key is stored securely using the platform's native keyring (via the [PAL](../internal/pal/README.md)).
