#!/bin/bash
# RepoNest install script for macOS and Linux.
# Installs the desktop app and the MCP server (reponest-mcp), which AI clients
# use to read the same local database.
#
# Release assets (see .github/workflows/release.yml):
#   Linux   reponest-linux-amd64.tar.gz   -> bare `reponest` binary
#   macOS   reponest-darwin-<arch>.dmg    -> RepoNest.app
#   MCP     reponest-mcp-<target>.tar.gz  -> bare `reponest-mcp` binary
set -e

BINARY_NAME="reponest"
MCP_BINARY_NAME="reponest-mcp"
REPO="sky-jiangcheng/RepoNest"
RELEASES="https://github.com/$REPO/releases/latest/download"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux)   TARGET="linux-amd64"; INSTALL_DIR="/usr/local/bin" ;;
  darwin)
    if [ "$ARCH" = "arm64" ]; then
      TARGET="darwin-arm64"
    else
      TARGET="darwin-amd64"
    fi
    INSTALL_DIR="/usr/local/bin"
    ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Run a command with sudo only when the install dir is not writable.
# Usage: as_root <cmd...>
as_root() {
  if [ -w "$INSTALL_DIR" ]; then
    "$@"
  else
    echo "Need sudo to write to $INSTALL_DIR"
    sudo "$@"
  fi
}

# Fetch $1 (an http(s) URL) into $2. Fails loudly on a non-2xx response so a
# wrong asset name surfaces as an error instead of a zero-byte binary.
fetch() {
  curl -fsSL --retry 3 -o "$2" "$1"
}

# Move the local file $1 to $2, preserving its executable bit.
install_binary() {
  as_root install -m 0755 "$1" "$2"
}

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# --- Desktop app --------------------------------------------------------------

if [ "$OS" = "darwin" ]; then
  DMG="$TMP_DIR/reponest.dmg"
  APP_NAME="RepoNest.app"
  echo "Downloading RepoNest for $TARGET (dmg)..."
  fetch "$RELEASES/reponest-$TARGET.dmg" "$DMG"

  MOUNT_DIR=$(hdiutil attach "$DMG" -nobrowse -readonly | tail -1 | sed -E 's|.*(/Volumes/.*)|\1|')
  trap 'rm -rf "$TMP_DIR"; [ -n "${MOUNT_DIR:-}" ] && hdiutil detach "$MOUNT_DIR" -quiet' EXIT
  if [ ! -d "$MOUNT_DIR/$APP_NAME" ]; then
    echo "ERROR: $APP_NAME not found inside the dmg" >&2
    exit 1
  fi
  as_root rm -rf "/Applications/$APP_NAME"
  as_root cp -R "$MOUNT_DIR/$APP_NAME" /Applications/
  hdiutil detach "$MOUNT_DIR" -quiet
  MOUNT_DIR=""

  echo ""
  echo "RepoNest installed to /Applications/$APP_NAME"
  echo "Run 'open -a \"$APP_NAME\"' to start!"
else
  TARBALL="$TMP_DIR/reponest.tar.gz"
  echo "Downloading RepoNest for $TARGET (tar.gz)..."
  fetch "$RELEASES/reponest-$TARGET.tar.gz" "$TARBALL"
  tar -xzf "$TARBALL" -C "$TMP_DIR" "$BINARY_NAME"
  install_binary "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

  echo ""
  echo "RepoNest installed to $INSTALL_DIR/$BINARY_NAME"
  echo "Run '$BINARY_NAME' to start!"
fi

# --- MCP server ---------------------------------------------------------------
# Ships as its own pure-Go archive, so AI clients can use RepoNest without the
# desktop app. Non-fatal: a release predating the split simply has no asset.
echo ""
echo "Downloading RepoNest MCP server for $TARGET..."

MCP_ASSET="reponest-mcp-$TARGET.tar.gz"
MCP_URL="$RELEASES/$MCP_ASSET"
if fetch "$MCP_URL" "$TMP_DIR/$MCP_ASSET" \
  && tar -xzf "$TMP_DIR/$MCP_ASSET" -C "$TMP_DIR" "$MCP_BINARY_NAME" \
  && install_binary "$TMP_DIR/$MCP_BINARY_NAME" "$INSTALL_DIR/$MCP_BINARY_NAME"; then
  echo "RepoNest MCP server installed to $INSTALL_DIR/$MCP_BINARY_NAME"
  echo "Register it with an AI client: claude mcp add reponest -- $INSTALL_DIR/$MCP_BINARY_NAME"
else
  echo "WARNING: could not install reponest-mcp (no asset for $TARGET, or download failed)." >&2
  echo "         The desktop app is installed and working; install the MCP server manually:" >&2
  echo "         $MCP_URL" >&2
  echo "         then: tar -xzf $MCP_ASSET && sudo install -m 0755 $MCP_BINARY_NAME $INSTALL_DIR/" >&2
fi
