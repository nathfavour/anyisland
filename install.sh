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
if [ "$OS" = "linux" ] && ( [ -d "/system/bin" ] || [ -n "$TERMUX_VERSION" ] ); then
    OS="android"
fi
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
BUILD_SUCCESS=false

# Check if we can build from local source (meaning we are running `./install.sh` inside the repo)
if [ -d "cmd/anyisland" ] && command -v go &> /dev/null; then
    echo "🔨 Building Anyisland from local source..."
    if go build -o anyisland ./cmd/anyisland; then
        BUILD_SUCCESS=true
    fi
fi

if [ "$BUILD_SUCCESS" != "true" ]; then
    echo "📥 Attempting to download pre-built binary for $OS-$ARCH..."
    EXT="tar.gz"
    if [ "$OS" = "windows" ]; then
        EXT="zip"
    fi
    BINARY_NAME="anyisland_${OS}_${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/nathfavour/anyisland/releases/download/v0.0.0/${BINARY_NAME}"
    
    if ! curl -fsSL "$DOWNLOAD_URL" -o anyisland_archive; then
        echo "⚠️  v0.0.0 release not found. Falling back to latest release..."
        DOWNLOAD_URL="https://github.com/nathfavour/anyisland/releases/latest/download/${BINARY_NAME}"
        curl -fsSL "$DOWNLOAD_URL" -o anyisland_archive
    fi
    
    if [ -f anyisland_archive ]; then
        echo "✅ Downloaded pre-built binary archive."
        if [ "$OS" = "windows" ]; then
            if command -v unzip &> /dev/null; then
                unzip -o anyisland_archive anyisland.exe
            else
                echo "⚠️  unzip command not found. Cannot extract archive."
                rm -f anyisland_archive
                exit 1
            fi
            mv anyisland.exe anyisland
        else
            tar -xzf anyisland_archive anyisland
        fi
        rm -f anyisland_archive
        BUILD_SUCCESS=true
    else
        echo "⚠️  Failed to download pre-built binary."
    fi
fi

if [ "$BUILD_SUCCESS" != "true" ]; then
    if command -v go &> /dev/null; then
        echo "🔨 Go is installed. Attempting to compile from remote source..."
        if GOBIN="$(pwd)" go install github.com/nathfavour/anyisland/cmd/anyisland@master; then
            echo "✅ Built successfully from remote source."
            BUILD_SUCCESS=true
        fi
    fi
fi

if [ "$BUILD_SUCCESS" != "true" ]; then
    echo "❌ Failed to download or build Anyisland binary. Exiting."
    exit 1
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
