#!/bin/bash
set -euo pipefail

# build-deployer.sh
#   builds op-deployer binary for a given version and caches it in $HOME/.cache/binaries/$DEPLOYER_VERSION/op-deployer,
#   where $DEPLOYER_VERSION is the github tag from monorepo for op-deployer of the form op-deployer/v0.4.0.
#   $DEPLOYER_VERSION corresponds to a value in the ops/internal/deployer/versions.json file
DEPLOYER_VERSION=$1
CACHE_DIR="$HOME/.cache/binaries"

# Determine the binary path
BINARY_PATH="$CACHE_DIR/$DEPLOYER_VERSION/op-deployer"

# Check if binary already exists
if [[ -f "$BINARY_PATH" && -x "$BINARY_PATH" ]]; then
    echo "Binary already exists at $BINARY_PATH"
    exit 0
fi
echo "Binary does not exist at $BINARY_PATH"
echo "Building binary..."

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Efficient clone - only fetches the specific tag without history
echo "Cloning monorepo tag $DEPLOYER_VERSION..."
git clone --depth 1 --branch "$DEPLOYER_VERSION" https://github.com/ethereum-optimism/optimism.git "$TEMP_DIR"

# Build binary
echo "Building binary..."
cd "$TEMP_DIR/op-deployer"
just build

# Copy to destination, e.g. /opt/binaries/op-deployer/v0.4.0-rc.2/op-deployer
echo "Installing binary to $BINARY_PATH..."
mkdir -p "$(dirname "$BINARY_PATH")"
cp bin/op-deployer "$BINARY_PATH"
chmod +x "$BINARY_PATH"

echo "Build complete!"