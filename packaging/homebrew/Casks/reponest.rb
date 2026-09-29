# RepoNest desktop app (macOS)
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
#
# sha256 must be filled after each release is published; Homebrew hard-fails on
# a mismatch, which is exactly what we want from a placeholder.
cask "reponest" do
  version "1.9.4"
  name "RepoNest"
  desc "Local-first code project context base"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-arm64.dmg"
    sha256 "2f576e056e0ec7858d70a93834b1963bcb16d1029162192ce39180dfa792255e"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-amd64.dmg"
    sha256 "e37b42253d7b9489ef5c7bcb2eee5c2f59422c878f05dbe862174d01579fe3a2"
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
