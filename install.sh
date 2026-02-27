#!/bin/sh
set -e

# git-share universal installer
# Usage: curl -sSf https://raw.githubusercontent.com/flawiddsouza/git-share/main/install.sh | sh

GITHUB_REPO="flawiddsouza/git-share"
BINARY_NAME="git-share"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux*) OS='linux' ;;
  darwin*) OS='darwin' ;;
  msys*|cygwin*|mingw*) OS='windows' ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
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

# Construct download URL
# Assumes naming convention: git-share-linux-amd64, git-share-darwin-arm64, etc.
FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
  FILENAME="${FILENAME}.exe"
fi

URL="https://github.com/$GITHUB_REPO/releases/download/${VERSION}/${FILENAME}"

echo "Installing ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}..."
echo "Downloading from ${URL}..."

# Download to temp file
TMP_BIN=$(mktemp)
curl -L -o "$TMP_BIN" "$URL"
chmod +x "$TMP_BIN"

# Move to /usr/local/bin
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Requesting sudo to install to ${INSTALL_DIR}..."
  sudo mv "$TMP_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
else
  mv "$TMP_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
