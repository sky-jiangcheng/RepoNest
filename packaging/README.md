# packaging/

Homebrew and Scoop manifests for RepoNest release artifacts.

The release workflow (`.github/workflows/release.yml`) publishes these assets per
tag, and the manifests below point at them:

| Asset | Platform | Consumers |
|-------|----------|-----------|
| `reponest-darwin-arm64.dmg` / `reponest-darwin-amd64.dmg` | macOS desktop app | Homebrew Cask |
| `reponest-linux-amd64.tar.gz` | Linux desktop app | manual |
| `reponest-windows-amd64.zip` | Windows desktop app | Scoop |
| `reponest-mcp-darwin-arm64.tar.gz` / `reponest-mcp-darwin-amd64.tar.gz` | macOS MCP server | Homebrew Cask |
| `reponest-mcp-linux-amd64.tar.gz` | Linux MCP server | Homebrew Formula |
| `reponest-mcp-windows-amd64.zip` | Windows MCP server | Scoop |

The desktop app is macOS/Windows/Linux, but only macOS has a real installer
artifact (`.dmg`); Linux ships a bare tarball and Windows a zip. The MCP server
is a pure Go / zero-CGO binary, so it installs cleanly everywhere via a Formula
or Cask `binary` stanza.

## Version is never edited by hand

`wails.json` → `info.productVersion` is the single source of truth. After
`scripts/bump-version.sh` bumps it, run:

```bash
./scripts/update-manifests.sh
```

which stamps the new version into every manifest below. `scripts/bump-version.sh`
should call it automatically — see the note in that script.

## sha256 fields are release-time

No release artifact exists for a given tag until the release workflow runs, so
the `sha256` fields ship as explicit `__FILL_SHA256_*__` placeholders. Homebrew
and Scoop both hard-fail on a mismatch, which is the behaviour we want: a
placeholder must never be mistaken for a verified hash.

After publishing a release, fill them in:

```bash
# for each asset you ship
curl -sL https://github.com/sky-jiangcheng/repo-nest/releases/download/v1.7.9/reponest-mcp-linux-amd64.tar.gz | shasum -a 256
```

## Where these are published

| | Repository | User command |
|---|---|---|
| Homebrew | [sky-jiangcheng/homebrew-repo](https://github.com/sky-jiangcheng/homebrew-repo) | `brew tap sky-jiangcheng/repo` |
| Scoop | [sky-jiangcheng/scoop-repo](https://github.com/sky-jiangcheng/scoop-repo) | `scoop bucket add repo https://github.com/sky-jiangcheng/scoop-repo` |

The tap keeps its own CI (`brew tap` → `brew install` → verify the installed
layout) so a broken manifest fails there rather than in a user's terminal. The
bucket validates each manifest's `hash` against the digest GitHub reports for
the release asset. Neither repository should be edited by hand: these files are
copied here on every release.

## Trying a manifest before publishing

```bash
# Ruby syntax only (no Homebrew needed)
ruby -c homebrew/Casks/reponest.rb

# JSON syntax only
python3 -m json.tool scoop/reponest-mcp.json > /dev/null
```

Both are safe to run in CI on every change.

## Cutting a release

```bash
# 1. Bump the version (SSOT is wails.json; this syncs package.json,
#    package-lock.json, version.go and every manifest below).
./scripts/bump-version.sh 1.8.0

# 2. Build what this host can build — the MCP server for all four targets.
#    The Wails desktop shell is skipped with a printed reason: it needs GTK on
#    Linux, Xcode on macOS and MSVC on Windows, so it comes from CI instead.
./scripts/build-release-assets.sh

# 3. Fill in the digests of whatever was built.
./scripts/update-manifests.sh --fill-sha256

# 4. Commit, tag, push. The tag is what makes release.yml run on the native
#    runners and produce the desktop artifacts.
git add -A && git commit -m "chore: release 1.8.0"
git tag -a v1.8.0 -m "Release v1.8.0"
git push origin master --tags

# 5. After CI publishes the desktop artifacts, download them into
#    assets/releases/v1.8.0/ and re-run step 3 so the two remaining
#    __FILL_SHA256_*__ placeholders are filled.
```

For v1.8.0 all eight digests are filled in and match what the release
actually serves. A manifest still holding `__FILL_SHA256_*__` is one Homebrew
and Scoop will refuse to install from — deliberate, since a missing digest
should fail loudly rather than silently install an unverified binary.

Digests must come from the **published release**, not from a local build of
the same version: a binary compiled on a different host (or with different
flags) is a different file, and a manifest carrying its digest would reject
every real download. Use the `digest` field the GitHub API returns for each
release asset, or download the assets and run `sha256sum` on them.

## Pruning old releases

`scripts/prune-releases.sh` removes historical releases and their assets
(~40 releases, >1 GB as of 1.8.0). It is dry-run by default, always keeps the
newest `--keep` releases, and leaves git tags alone unless `--delete-tags` is
passed. It needs `gh` and an authenticated session.

```bash
./scripts/prune-releases.sh --list              # inventory
./scripts/prune-releases.sh --keep 3            # show what would go
./scripts/prune-releases.sh --keep 3 --yes      # do it
```
