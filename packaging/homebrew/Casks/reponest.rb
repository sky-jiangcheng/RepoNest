# RepoNest desktop app (macOS)
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
#
# sha256 must be filled after each release is published; Homebrew hard-fails on
# a mismatch, which is exactly what we want from a placeholder.
cask "reponest" do
  version "1.8.1"
  name "RepoNest"
  desc "Local-first code project context base"
  homepage "https://github.com/sky-jiangcheng/reponest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/reponest/releases/download/v#{version}/reponest-darwin-arm64.dmg"
    sha256 "da2b09c034ba9660d685185bf1355a20f0960fb8573082069610980c2c43910f"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/reponest/releases/download/v#{version}/reponest-darwin-amd64.dmg"
    sha256 "cab81e34f3c71f2301f95ffa4c57707fa682a7a6f31d37895d3019abfd0d89f1"
  end

  app "RepoNest.app"

  # Application Support and Logs paths come from internal/platform; they are
  # where GetDbPath() and GetLogPath() point on macOS. A bundle-identifier
  # preference is deliberately not listed here: wails.json does not set one, so
  # listing a guess would send `brew uninstall --zap` at a file that may not
  # exist. Add it only once the identifier is pinned in wails.json.
  zap trash: [
    "~/Library/Application Support/reponest",
    "~/Library/Logs/reponest.log",
  ]
end
