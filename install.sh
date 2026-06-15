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

# 3. Build/Download Binary
if command -v go &> /dev/null; then
    echo "🔨 Building Anyisland from source..."
    if ! go build -o anyisland ./cmd/anyisland; then
        echo "⚠️  Build failed, attempting to download pre-built binary..."
        curl -fsSL "https://github.com/nathfavour/anyisland/releases/latest/download/anyisland_${OS}_${ARCH}.tar.gz" -o anyisland.tar.gz || {
            echo "❌ Failed to download pre-built binary. Exiting."
            exit 1
        }
        tar -xzf anyisland.tar.gz anyisland || unzip anyisland.tar.gz anyisland.exe
        rm -f anyisland.tar.gz
    fi
else
    echo "⚠️  Go not found. Downloading pre-built binary for $OS-$ARCH..."
    curl -fsSL "https://github.com/nathfavour/anyisland/releases/latest/download/anyisland_${OS}_${ARCH}.tar.gz" -o anyisland.tar.gz || {
        echo "❌ Failed to download pre-built binary. Exiting."
        exit 1
    }
    tar -xzf anyisland.tar.gz anyisland || unzip anyisland.tar.gz anyisland.exe
    rm -f anyisland.tar.gz
fi

# 4. Hand-off to Anyisland for self-installation
echo "🚚 Handing off to Anyisland for system integration..."
if [ -f "anyisland.exe" ]; then
    mv anyisland.exe anyisland
fi
chmod +x anyisland
mkdir -p "$LOCAL_BIN"
mv anyisland "$LOCAL_BIN/"

"$LOCAL_BIN/anyisland" self-install

echo ""
echo "✅ Anyisland installation complete!"
echo "🚀 Run 'anyisland' to get started."
echo "👉 You may need to restart your shell to update your PATH."
