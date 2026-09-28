#!/usr/bin/env bash
# Propagate wails.json's productVersion into the Homebrew / Scoop manifests.
#
# Called by scripts/bump-version.sh after the SSOT has moved, and runnable
# standalone (idempotent) to re-sync without a version change:
#
#   ./scripts/update-manifests.sh
#
# The sha256 fields are deliberately NOT touched: they are release-time facts
# that only exist once the release workflow has published the assets, and both
# Homebrew and Scoop hard-fail on a mismatch. They ship as __FILL_SHA256_*__
# placeholders and are filled in by hand after publishing.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VERSION="$(grep -oE '"productVersion"[[:space:]]*:[[:space:]]*"[^"]+"' wails.json | head -1 | sed -E 's/.*"([^"]+)"$/\1/')"
if [[ -z "$VERSION" ]]; then
  echo "ERROR: could not read productVersion from wails.json" >&2
  exit 1
fi
echo "Syncing package manifests to $VERSION"

# Portable in-place edit, same technique (and same reason) as bump-version.sh:
# `sed -i` is not portable across BSD/GNU/toybox, so write to a temp file and
# rename. The quotes live on the replacement side — writing "\1$VERSION" would
# expand to e.g. "\11.7.9", which sed reads as back-reference 11.
update_file() {
  local file="$1" pattern="$2" replacement="$3"
  if [[ ! -f "$file" ]]; then
    echo "  WARN: $file not found, skipping" >&2
    return 0
  fi
  # mktemp creates 0600, so reset the mode on the replacement — otherwise every
  # run silently tightens the manifest's permissions. 0644 explicitly rather
  # than "copy the old mode via ls -l": manifest files are plain data, and
  # parsing `ls` output is not portable (toybox/busybox ls formats it
  # differently, which is the same trap that makes `sed -i` unusable here).
  local tmp
  tmp="$(mktemp "${file}.XXXXXX")"
  if ! sed -E "s|$pattern|$replacement|" "$file" > "$tmp"; then
    echo "  ERROR: sed failed for $file" >&2
    rm -f "$tmp"
    return 1
  fi
  chmod 0644 "$tmp"
  mv "$tmp" "$file"
  echo "  updated $file"
}

for cask in packaging/homebrew/Casks/reponest.rb \
            packaging/homebrew/Casks/reponest-mcp.rb \
            packaging/homebrew/Formula/reponest-mcp.rb; do
  # Ruby: `  version "1.7.9"`
  update_file "$cask" \
    '(version[[:space:]]+)"[^"]+"' \
    "\1\"$VERSION\""
done

for manifest in packaging/scoop/reponest.json packaging/scoop/reponest-mcp.json; do
  # JSON: the version is pinned in two places, and Scoop has no interpolation
  # in the top-level `url` — so both must move together or `scoop install`
  # silently resolves a 404.
  update_file "$manifest" \
    '("version"[[:space:]]*:[[:space:]]*)"[^"]+"' \
    "\1\"$VERSION\""
  update_file "$manifest" \
    '(releases/download/)v[0-9]+\.[0-9]+\.[0-9]+([+-][A-Za-z0-9.-]+)?' \
    "\1v$VERSION"
done

echo
echo "Done. Verify with:"
echo "  grep -rn '$VERSION' packaging/"
echo "  ruby -c packaging/homebrew/Casks/reponest.rb"
echo "  python3 -m json.tool packaging/scoop/reponest-mcp.json > /dev/null"
echo
echo "sha256 fields still need filling once the release assets are published —"
echo "see packaging/README.md."

# --- sha256 filling -----------------------------------------------------------
#
# `--fill-sha256` reads assets/releases/v<version>/SHA256SUMS and replaces the
# __FILL_SHA256_*__ placeholders in packaging/ with the real digests. It only
# touches a manifest when the exact asset that manifest's URL points at is
# present in the checksum file, so a partial release leaves the remaining
# placeholders in place instead of writing a digest that does not correspond to
# the download.
#
# Run it *after* the release assets are published:
#   ./scripts/update-manifests.sh --fill-sha256
if [ "${1:-}" = "--fill-sha256" ]; then
  SUMS="$ROOT/assets/releases/v$VERSION/SHA256SUMS"
  if [ ! -f "$SUMS" ]; then
    echo "ERROR: $SUMS not found. Build the assets first:" >&2
    echo "         ./scripts/build-release-assets.sh" >&2
    exit 1
  fi

  sha_for() { awk -v want="$1" '$2 ~ ("/" want "$") { print $1 }' "$SUMS" | head -1; }

  fill() {
    local file="$1" asset="$2" placeholder="$3"
    local digest
    digest="$(sha_for "$asset")"
    if [ -z "$digest" ]; then
      echo "  skip $file: $asset not in SHA256SUMS (placeholder kept)"
      return 0
    fi
    if ! grep -q "$placeholder" "$file"; then
      return 0
    fi
    # Portable in-place edit; see the note above on why sed -i is avoided.
    # The trap removes the temp file if sed or chmod fails, so an interrupted
    # run cannot leave a `<manifest>.XXXXXX` turd in the tree.
    local tmp
    tmp="$(mktemp "${file}.XXXXXX")"
    trap 'rm -f "${tmp:-}"' RETURN
    sed -E "s|$placeholder|$digest|g" "$file" > "$tmp"
    chmod 0644 "$tmp"
    mv "$tmp" "$file"
    trap - RETURN
    echo "  filled $file <- $asset ($digest)"
  }

  echo "Filling sha256 from $SUMS"
  fill packaging/homebrew/Casks/reponest-mcp.rb       reponest-mcp-darwin-arm64.tar.gz  __FILL_SHA256_DARWIN_ARM64_MCP_TARBALL__
  fill packaging/homebrew/Formula/reponest-mcp.rb     reponest-mcp-linux-amd64.tar.gz   __FILL_SHA256_LINUX_AMD64_MCP_TARBALL__
  fill packaging/scoop/reponest-mcp.json              reponest-mcp-windows-amd64.zip   sha256:__FILL_SHA256_WINDOWS_AMD64_MCP_ZIP__
  fill packaging/homebrew/Casks/reponest.rb           reponest-darwin-arm64.dmg        __FILL_SHA256_DARWIN_ARM64_DMG__
  fill packaging/scoop/reponest.json                  reponest-windows-amd64.zip      sha256:__FILL_SHA256_WINDOWS_AMD64_ZIP__

  echo
  echo "Any manifest still holding a __FILL_SHA256_*__ placeholder needs a"
  echo "release artifact that was not built on this host - build it, or leave it."
  exit 0
fi
