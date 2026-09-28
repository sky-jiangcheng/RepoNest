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
  version "1.9.1"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "d2c6ed0f9379be05725c4b3a6034618ac53e8fc61e9365b13dbf4ba35cdb69ec"
  elsif OS.mac? && Hardware::CPU.is_64_bit?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "66a93042d9dd776dd0e7c5edf6950545e0ee198016b9c3da699b56a8ec55738b"
  else
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-linux-amd64.tar.gz"
    sha256 "2bbc97f90237601f6ee2a3310895f4846ef16abca56703c90f3aa91e8b69cfdf"
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
