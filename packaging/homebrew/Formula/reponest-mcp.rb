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
  version "1.9.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "70d9773f8b74820b30dbdfe86be8498632459dfa8e41f708981eb208d75c785e"
  elsif OS.mac? && Hardware::CPU.is_64_bit?
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "bbd2fdbb62be06a0987fae7c991a39827cba79959ca33b1292efdc1770c6560f"
  else
    url "https://github.com/sky-jiangcheng/RepoNest/releases/download/v#{version}/reponest-mcp-linux-amd64.tar.gz"
    sha256 "222bed58e6ee8406bd2544180c67a5f5261787fe9a676b06d623a949c30215c8"
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
