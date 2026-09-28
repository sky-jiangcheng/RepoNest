# RepoNest desktop app (macOS)
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
#
# sha256 must be filled after each release is published; Homebrew hard-fails on
# a mismatch, which is exactly what we want from a placeholder.
cask "reponest" do
  version "1.9.0"
  name "RepoNest"
  desc "Local-first code project context base"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-arm64.dmg"
    sha256 "e742904d2c26da0f08698551c56548f913c6e7725d7a583726e44eadd28ed55d"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-darwin-amd64.dmg"
    sha256 "6e3030d046dc657d07e7b1edbad5e878b7c9f1d6dcf8d9eed96143f8bf327ca0"
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
