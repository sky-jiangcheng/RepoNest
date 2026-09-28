# RepoNest MCP server (macOS)
#
# The MCP stdio server is the AI execution interface for RepoNest — the desktop
# app is not required to use it, it reads the same local SQLite database.
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
cask "reponest-mcp" do
  version "1.9.3"
  name "RepoNest MCP Server"
  desc "MCP stdio server exposing the RepoNest local knowledge base to AI agents"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  license "MIT"

  on_arm do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "959b8711047ca994ae627d72a63de6cd3a42dafe19a8ed194f474eadd736112d"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "eab3cbca578a2054ca3ad9da7e6ee091cda681f247732ad04c961be7de5bb0e9"
  end

  # The tarball stages a single bare `reponest-mcp` at its root; `binary` links
  # it into $(brew --prefix)/bin so `which reponest-mcp` works for
  # `claude mcp add reponest -- $(which reponest-mcp)`.
  binary "reponest-mcp"
end
