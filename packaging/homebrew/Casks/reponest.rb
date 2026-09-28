# RepoNest desktop app (macOS)
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
#
# sha256 must be filled after each release is published; Homebrew hard-fails on
# a mismatch, which is exactly what we want from a placeholder.
cask "reponest" do
  version "1.9.3"
  name "RepoNest"
  desc "Local-first code project context base"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-arm64.dmg"
    sha256 "67fef545b84b29311563371318896903fba0fde9e153fffce0d02f2fa54b5265"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-amd64.dmg"
    sha256 "5fbdfea153a63f46c3205a457e36e631bdd7eb712b99a9e714d5dbcfea2d0e03"
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
