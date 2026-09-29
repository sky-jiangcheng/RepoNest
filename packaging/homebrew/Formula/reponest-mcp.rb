# RepoNest MCP server (Linux)
#
# The desktop app has no Linux installer artifact (it ships as a bare tarball),
# but the MCP server is a pure Go / zero-CGO binary and installs cleanly, so the
# Formula covers MCP on Linux rather than the GUI app.
#
# Version is stamped by scripts/update-manifests.sh from wails.json
# (info.productVersion) — never edit it by hand.
class ReponestMcp < Formula
  desc "MCP stdio server exposing the RepoNest local knowledge base to AI agents"
  homepage "https://github.com/sky-jiangcheng/RepoNest"
  version "1.9.4"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "7f2f1f4fc8c62eef7ecf57ef3ec666bd5233051f444670e8d4604f93dfac26f9"
  elsif OS.mac? && Hardware::CPU.is_64_bit?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "145bb9cf597837c57200c8596b21650c9e66c930ba08535ff50c7cd1390f937a"
  else
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-linux-amd64.tar.gz"
    sha256 "1ec3b2aba0b71494d74d0eff25da006ee40c5f6e94ab572c1bb0fa64aa1cb3e5"
  end

  def install
    bin.install "reponest-mcp"
  end

  # The binary is a stdio MCP server: running it with no stdin makes it block
  # on the JSON-RPC loop, so it cannot be executed as a test. Assert the file
  # landed and is executable instead of pretending to run it.
  test do
    assert_path_exists bin/"reponest-mcp"
    assert_predicate bin/"reponest-mcp", :executable?
  end
end
