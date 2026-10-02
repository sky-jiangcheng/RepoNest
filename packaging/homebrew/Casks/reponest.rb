# RepoNest desktop app (macOS)
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
#
# sha256 must be filled after each release is published; Homebrew hard-fails on
# a mismatch, which is exactly what we want from a placeholder.
cask "reponest" do
  version "1.12.0"
  name "RepoNest"
  desc "Local-first code project context base"
  homepage "https://github.com/sky-jiangcheng/repo-nest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/repo-nest/releases/download/v#{version}/reponest-darwin-arm64.dmg"
    sha256 "41a57ab7ab0df17036489b6362a665fa12c9ec82c2a890fbde5b83f07e625246"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/repo-nest/releases/download/v#{version}/reponest-darwin-amd64.dmg"
    sha256 "ac770bfdfd69d703951cad734dc8bfe3c6f141a8ab2ac8dfaeffbe7af84df5bb"
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
