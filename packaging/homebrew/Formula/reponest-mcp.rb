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
  version "1.8.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "__FILL_SHA256_DARWIN_ARM64_MCP_TARBALL__"
  elsif OS.mac? && Hardware::CPU.is_64_bit?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "__FILL_SHA256_DARWIN_AMD64_MCP_TARBALL__"
  else
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-linux-amd64.tar.gz"
    sha256 "bea0a8a3c29ad293a23c47b097095ed72cd461ac3a05d0b229c36110d65a9af9"
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
