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
    sha256 "8b65d256c7c75e41ff06ef7f3582456aa2e2b823263ef70da8011a345952b4a9"
  end

  on_intel do
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "72b106626373ad52ea7b6836c0f3223e9f097be1c30306f87eb077da1c406018"
  end

  # The tarball stages a single bare `reponest-mcp` at its root; `binary` links
  # it into $(brew --prefix)/bin so `which reponest-mcp` works for
  # `claude mcp add reponest -- $(which reponest-mcp)`.
  binary "reponest-mcp"
end
