#!/bin/sh
set -e

# git-share universal installer for Linux and macOS
# Usage: curl -sSf https://raw.githubusercontent.com/flawiddsouza/git-share/main/install.sh | sh

GITHUB_REPO="flawiddsouza/git-share"
BINARY_NAME="git-share"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux*) OS='linux' ;;
  darwin*) OS='darwin' ;;
  *) echo "Unsupported OS: $OS. This installer supports Linux and macOS only."; exit 1 ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH='amd64' ;;
  arm64|aarch64) ARCH='arm64' ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release version
VERSION=$(curl -sI https://github.com/$GITHUB_REPO/releases/latest | grep -i "location:" | awk -F "/" '{print $NF}' | tr -d '\r')

if [ -z "$VERSION" ]; then
  # Fallback if no releases found yet
  VERSION="v1.0.0"
fi

# Strip 'v' prefix for filename (releases use v1.0.0 in URL but 1.0.0 in filename)
VERSION_NO_V=$(echo "$VERSION" | sed 's/^v//')

# Construct download URL
# Naming convention: git-share-1.0.0-linux-amd64.tar.gz (version without 'v')
FILENAME="${BINARY_NAME}-${VERSION_NO_V}-${OS}-${ARCH}.tar.gz"
# URL path uses version with 'v', filename uses version without 'v'
URL="https://github.com/$GITHUB_REPO/releases/download/${VERSION}/${FILENAME}"

echo "Installing ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}..."
echo "Downloading from ${URL}..."

# Create temp directory
TMP_DIR=$(mktemp -d)
ARCHIVE_PATH="${TMP_DIR}/${FILENAME}"

# Download archive
curl -L -o "$ARCHIVE_PATH" "$URL"

# Extract archive
echo "Extracting..."
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

# Find the binary
BINARY_PATH="${TMP_DIR}/${BINARY_NAME}"

# Make executable
chmod +x "$BINARY_PATH"

# Move to /usr/local/bin
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Requesting sudo to install to ${INSTALL_DIR}..."
  sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
else
  mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
