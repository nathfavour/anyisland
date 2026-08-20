#!/bin/bash
set -e

# Anyisland Universal Bootstrap & Tool Installer Hook
# Prepares the host, bootstraps Anyisland, and optionally installs a target repository.

ISLAND_DIR="$HOME/.anyisland"
LOCAL_BIN="$HOME/.local/bin"
VERSION="latest"
TARGET_REPO="${1:-$ANYISLAND_TARGET}"

if [ -n "$TARGET_REPO" ]; then
    echo "🏝️  Anyisland Universal Bootstrap (Installing: $TARGET_REPO)"
else
    echo "🏝️  Anyisland Bootstrap"
fi

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
ALREADY_INSTALLED=false
if command -v anyisland &> /dev/null; then
    CURRENT_COMMIT=$(anyisland version 2>/dev/null | grep "Commit:" | awk '{print $2}' | cut -d'(' -f1 | xargs || true)
    REPO_URL="https://github.com/nathfavour/anyisland"
    
    if command -v git &> /dev/null; then
        REMOTE_COMMIT=$(git ls-remote "$REPO_URL" HEAD 2>/dev/null | awk '{print $1}' || true)
        if [ -n "$CURRENT_COMMIT" ] && [ "$CURRENT_COMMIT" = "$REMOTE_COMMIT" ]; then
            echo "✅ Anyisland is already at the latest version ($CURRENT_COMMIT)."
            ALREADY_INSTALLED=true
        fi
    else
        ALREADY_INSTALLED=true
    fi
fi

# 3. Build/Download Binary if needed
if [ "$ALREADY_INSTALLED" = false ]; then
    BUILD_SUCCESS=false

    # Check if we can build from local source
    if [ -d "cmd/anyisland" ] && command -v go &> /dev/null; then
        echo "🔨 Building Anyisland from local source..."
        if go build -o anyisland ./cmd/anyisland 2>/dev/null; then
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
        DOWNLOAD_URL="https://github.com/nathfavour/anyisland/releases/latest/download/${BINARY_NAME}"
        
        if curl -fsSL "$DOWNLOAD_URL" -o anyisland_archive 2>/dev/null; then
            echo "✅ Downloaded pre-built binary archive."
            if [ "$OS" = "windows" ]; then
                if command -v unzip &> /dev/null; then
                    unzip -o anyisland_archive anyisland.exe
                fi
                [ -f anyisland.exe ] && mv anyisland.exe anyisland
            else
                tar -xzf anyisland_archive anyisland 2>/dev/null || true
            fi
            rm -f anyisland_archive
            [ -f anyisland ] && BUILD_SUCCESS=true
        fi
    fi

    if [ "$BUILD_SUCCESS" != "true" ]; then
        if command -v go &> /dev/null; then
            echo "🔨 Go is installed. Attempting to compile from remote source..."
            if GOBIN="$(pwd)" go install github.com/nathfavour/anyisland/cmd/anyisland@master 2>/dev/null; then
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
    chmod +x anyisland 2>/dev/null || true
    mkdir -p "$LOCAL_BIN"
    mv anyisland "$LOCAL_BIN/"

    "$LOCAL_BIN/anyisland" self-install
fi

# 5. Handle Target Repo Hook if provided
if [ -n "$TARGET_REPO" ]; then
    echo "📦 Installing target package: $TARGET_REPO via Anyisland..."
    ANYISLAND_BIN="$LOCAL_BIN/anyisland"
    if ! command -v anyisland &> /dev/null; then
        export PATH="$LOCAL_BIN:$PATH"
    fi
    "$ANYISLAND_BIN" install "$TARGET_REPO"
    echo ""
    echo "🎉 Successfully installed $TARGET_REPO!"
else
    echo ""
    echo "✅ Anyisland installation complete!"
    echo "🚀 Run 'anyisland' to get started."
    echo "👉 You may need to restart your shell to update your PATH."
fi
