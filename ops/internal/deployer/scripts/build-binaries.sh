#!/bin/bash
set -euo pipefail

# build-deployer.sh
#   Downloads pre-built op-deployer binaries for specified versions from GitHub releases
#   and caches them in $HOME/.cache/deployer/$VERSION/op-deployer
#   where $VERSION is the version number (e.g., v0.4.0) without the op-deployer/ prefix.
#   If a pre-built binary isn't available, falls back to building from source.

CACHE_DIR="$HOME/.cache/op-deployer"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSIONS_JSON="$SCRIPT_DIR/../versions.json"
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Determine OS type and architecture
detect_os() {
  case "$(uname -s)" in
    Linux*)     OS="linux";;
    Darwin*)    OS="darwin";;
    MINGW*)     OS="windows";;
    *)          echo "Unsupported OS: $(uname -s)" && exit 1;;
  esac

  case "$(uname -m)" in
    x86_64)     ARCH="amd64";;
    arm64|aarch64) ARCH="arm64";;
    *)          echo "Unsupported architecture: $(uname -m)" && exit 1;;
  esac

  echo "Detected OS: $OS, Architecture: $ARCH"
}

# Get available versions from versions.json
get_versions() {
  if [ ! -f "$VERSIONS_JSON" ]; then
    echo "Error: versions.json not found at $VERSIONS_JSON"
    exit 1
  fi

  # Extract op-deployer versions using jq
  VERSIONS=$(jq -r '.[]' "$VERSIONS_JSON" | sort -u)

  if [ -z "$VERSIONS" ]; then
    echo "Error: No versions found in versions.json"
    exit 1
  fi
}

# Download and install a specific version
download_and_install() {
  local full_version=$1
  # Strip the op-deployer/ prefix for the cache directory
  local stripped_version=$(echo "$full_version" | grep -o 'v[0-9][^"]*')
  local binary_path="$CACHE_DIR/$stripped_version/op-deployer"

  # Check if binary already exists
  if [[ -f "$binary_path" && -x "$binary_path" ]]; then
    echo "Binary already exists at $binary_path"
    return 0
  fi

  echo "Binary does not exist at $binary_path"

  # Try downloading from GitHub releases first
  if try_download_release "$full_version" "$stripped_version" "$binary_path"; then
    echo "Successfully installed $stripped_version from GitHub releases"
    return 0
  fi

  # If download fails, fall back to building from source
  echo "Pre-built binary not available. Falling back to building from source..."
  if build_from_source "$full_version" "$binary_path"; then
    echo "Successfully built and installed $stripped_version from source"
    return 0
  fi

  echo "Failed to install $full_version"
  return 1
}

try_download_release() {
  local full_version=$1
  local stripped_version=$2
  local binary_path=$3

  echo "Attempting to download $full_version for $OS-$ARCH from GitHub releases..."

  # Construct GitHub release URL
  local github_release_url="https://github.com/ethereum-optimism/optimism/releases/tag/$full_version"
  echo "Checking release at $github_release_url"

  # Extract version number without prefix for asset filename
  local version_number=$(echo "$stripped_version" | sed 's/^v//')

  # Construct asset filename based on OS and architecture with version number
  local asset_name="op-deployer-${version_number}-${OS}-${ARCH}.tar.gz"
  if [ "$OS" = "windows" ]; then
    asset_name="op-deployer-${version_number}-${OS}-${ARCH}.zip"
  fi

  # Use the correct download URL format (keep the slash, don't URL encode it)
  local download_url="https://github.com/ethereum-optimism/optimism/releases/download/$full_version/$asset_name"

  # Download the archive file (silent failure)
  echo "Downloading from $download_url"
  if ! curl --fail -s -L -o "$TEMP_DIR/$asset_name" "$download_url"; then
    echo "Release binary not found at $download_url"
    return 1
  fi

  # Create extraction directory
  local extract_dir="$TEMP_DIR/extract"
  mkdir -p "$extract_dir"

  # Extract the archive file
  echo "Extracting archive..."
  if [ "$OS" = "windows" ]; then
    # Handle zip files for Windows
    if ! unzip -q "$TEMP_DIR/$asset_name" -d "$extract_dir"; then
      echo "Error: Failed to extract $TEMP_DIR/$asset_name"
      return 1
    fi
  else
    # Handle tar.gz for Linux and macOS
    if ! tar -xzf "$TEMP_DIR/$asset_name" -C "$extract_dir"; then
      echo "Error: Failed to extract $TEMP_DIR/$asset_name"
      return 1
    fi
  fi

  # Find the op-deployer binary in the extracted content
  local binary_name="op-deployer"
  if [ "$OS" = "windows" ]; then
    binary_name="op-deployer.exe"
  fi

  # Check if binary exists in extract directory
  if [ -f "$extract_dir/$binary_name" ]; then
    local extracted_binary="$extract_dir/$binary_name"
  else
    # Try to find it recursively
    local extracted_binary=$(find "$extract_dir" -name "$binary_name" -type f | head -n 1)
    if [ -z "$extracted_binary" ]; then
      echo "Error: Could not find $binary_name in extracted archive"
      return 1
    fi
  fi

  # Move to destination and make executable
  echo "Installing binary to $binary_path"
  mkdir -p "$(dirname "$binary_path")"
  mv "$extracted_binary" "$binary_path"
  chmod +x "$binary_path"

  return 0
}

build_from_source() {
  local full_version=$1
  local binary_path=$2

  echo "Building $full_version from source..."

  # Create temp directory for source code
  local source_dir="$TEMP_DIR/source"
  mkdir -p "$source_dir"

  # Clone the repository at the specific tag
  echo "Cloning repository at tag $full_version..."
  if ! git clone --depth 1 --branch "$full_version" https://github.com/ethereum-optimism/optimism.git "$source_dir"; then
    echo "Failed to clone repository at tag $full_version"
    return 1
  fi

  # Navigate to the op-deployer directory
  cd "$source_dir/op-deployer"

  # Build the binary
  echo "Building op-deployer..."
  if command -v just &>/dev/null; then
    # Use 'just' if available
    if ! just build; then
      echo "Failed to build using 'just build'"
      return 1
    fi
  else
    # Fall back to Go build directly
    echo "'just' command not found, using 'go build' directly..."
    if ! go build -o bin/op-deployer ./cmd/main.go; then
      echo "Failed to build using 'go build'"
      return 1
    fi
  fi

  # Copy the binary to the destination
  echo "Installing binary to $binary_path"
  mkdir -p "$(dirname "$binary_path")"
  cp bin/op-deployer "$binary_path"
  chmod +x "$binary_path"

  return 0
}

main() {
  detect_os
  get_versions

  local failed=0

  # Process each version
  for version in $VERSIONS; do
    echo "Processing $version..."
    if ! download_and_install "$version"; then
      echo "Failed to process $version"
      failed=1
    fi
  done

  if [ $failed -eq 1 ]; then
    echo "One or more binaries could not be downloaded or installed"
    exit 1
  fi

  echo "All binaries are successfully installed!"
  exit 0
}

main