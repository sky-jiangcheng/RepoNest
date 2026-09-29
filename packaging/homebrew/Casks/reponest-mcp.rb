# RepoNest MCP server (macOS)
#
# The MCP stdio server is the AI execution interface for RepoNest — the desktop
# app is not required to use it, it reads the same local SQLite database.
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
cask "reponest-mcp" do
  version "1.9.4"
  name "RepoNest MCP Server"
  desc "MCP stdio server exposing the RepoNest local knowledge base to AI agents"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "7f2f1f4fc8c62eef7ecf57ef3ec666bd5233051f444670e8d4604f93dfac26f9"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "145bb9cf597837c57200c8596b21650c9e66c930ba08535ff50c7cd1390f937a"
  end

  # The tarball stages a single bare `reponest-mcp` at its root; `binary` links
  # it into $(brew --prefix)/bin so `which reponest-mcp` works for
  # `claude mcp add reponest -- $(which reponest-mcp)`.
  binary "reponest-mcp"
end
