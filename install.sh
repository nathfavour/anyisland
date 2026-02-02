#!/bin/bash
set -e

# Anyisland Bootstrap Script
# This script builds Anyisland from source and installs it to ~/.local/bin

ISLAND_DIR="$HOME/.anyisland"
LOCAL_BIN="$HOME/.local/bin"

echo "🏝️  Installing Anyisland to $LOCAL_BIN..."

# 1. Check dependencies
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go to build Anyisland."
    exit 1
fi

# 2. Build binaries
echo "🔨 Building Anyisland..."
go build -o anyisland ./cmd/anyisland
go build -o anyislandd ./cmd/anyislandd

# 3. Setup directories
mkdir -p "$LOCAL_BIN"
mkdir -p "$ISLAND_DIR"

# 4. Install binaries
echo "🚚 Installing binaries to $LOCAL_BIN..."
mv anyisland "$LOCAL_BIN/"
mv anyislandd "$LOCAL_BIN/"

# 5. Initialize environment
echo "⚙️  Initializing Anyisland..."
"$LOCAL_BIN/anyisland" init

echo ""
echo "✅ Anyisland installation complete!"
echo "🚀 The 'anyisland' and 'anyislandd' binaries are now in $LOCAL_BIN"
echo "👉 Ensure $LOCAL_BIN is in your PATH."