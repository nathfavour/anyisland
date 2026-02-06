#!/bin/bash
set -e

# Anyisland Universal Bootstrap
# This script prepares the system and hands off to the Anyisland binary.

ISLAND_DIR="$HOME/.anyisland"
LOCAL_BIN="$HOME/.local/bin"
VERSION="latest"

echo "🏝️  Anyisland Bootstrap"

# 1. Detect Platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "📍 Platform: $OS-$ARCH"

# 2. Check if already installed and up-to-date
if command -v anyisland &> /dev/null; then
    CURRENT_COMMIT=$(anyisland version | grep "Commit:" | awk '{print $2}' | cut -d'(' -f1 | xargs)
    REPO_URL="https://github.com/nathfavour/anyisland"
    
    if command -v git &> /dev/null; then
        REMOTE_COMMIT=$(git ls-remote "$REPO_URL" HEAD | awk '{print $1}')
        if [ "$CURRENT_COMMIT" = "$REMOTE_COMMIT" ]; then
            echo "✅ Anyisland is already at the latest version ($CURRENT_COMMIT)."
            exit 0
        fi
    fi
fi

# 3. Check for Go (Fallback)
if ! command -v go &> /dev/null; then
    echo "⚠️  Go not found. In a full release, I would download a pre-built binary for $OS-$ARCH."
    echo "🔨 For now, please ensure Go is installed if you want to build from source."
    # exit 1 # In real release, we would download here
fi

# 3. Build/Download Binary
if command -v go &> /dev/null; then
    echo "🔨 Building Anyisland from source..."
    go build -o anyisland ./cmd/anyisland
else
    # This is where we would curl the binary
    # curl -L https://github.com/nathfavour/anyisland/releases/latest/download/anyisland-$OS-$ARCH -o anyisland
    echo "Error: Binary download not yet implemented in this prototype. Please install Go."
    exit 1
fi

# 4. Hand-off to Anyisland for self-installation
echo "🚚 Handing off to Anyisland for system integration..."
chmod +x anyisland
mkdir -p "$LOCAL_BIN"
mv anyisland "$LOCAL_BIN/"

"$LOCAL_BIN/anyisland" self-install

echo ""
echo "✅ Anyisland installation complete!"
echo "🚀 Run 'anyisland' to get started."
echo "👉 You may need to restart your shell to update your PATH."
