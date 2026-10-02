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
  homepage "https://github.com/sky-jiangcheng/repo-nest"
  version "1.11.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/sky-jiangcheng/repo-nest/releases/download/v#{version}/reponest-mcp-darwin-arm64.tar.gz"
    sha256 "c849f07d870b1e99895c7eac533715e548fbf95d48d2371bd0cc37529abe055e"
  elsif OS.mac? && Hardware::CPU.is_64_bit?
    url "https://github.com/sky-jiangcheng/repo-nest/releases/download/v#{version}/reponest-mcp-darwin-amd64.tar.gz"
    sha256 "19bc379852a8db607d1e61d3ca808e3d62bc7bfb4a8d59421681f54be1bd0491"
  else
    url "https://github.com/sky-jiangcheng/repo-nest/releases/download/v#{version}/reponest-mcp-linux-amd64.tar.gz"
    sha256 "daa1c36ea9061968a41e929efe040579a17b782718fdc48882d77a8967f2194f"
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
