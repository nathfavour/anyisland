# Installation & System Integration

Anyisland is designed to be installed without root permissions, living entirely within the user's home directory. It supports standalone installation as well as unified single-command installation for target tools.

## The Universal Bootstrap Process

### 1. Standalone Anyisland Installation
```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash
```

### 2. Single-Command Tool Installation (Target Hook)
Any tool can use Anyisland as its zero-friction installer by passing its GitHub repository slug as an argument to the bootstrap script:

```bash
curl -fsSL https://raw.githubusercontent.com/nathfavour/anyisland/master/install.sh | bash -s -- <owner/repo>
```

When a target repo is passed:
1. The script checks if Anyisland is installed/up-to-date and bootstraps it if necessary.
2. The script immediately executes `anyisland install <owner/repo>` to resolve dependencies and build the tool.

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

During the first installation, Anyisland generates a unique **Master Key** used for E2EE shell history and local vault encryption. This key is stored securely using the platform's native keyring (via the Platform Abstraction Layer).
