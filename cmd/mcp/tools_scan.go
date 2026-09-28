package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"reponest/internal/service"
)

// registerScanTool wires reponest_scan: the headless cold-start tool. Without
// it a pure-MCP install (no desktop app) could never populate the database,
// so the advertised "the desktop app is not required" path led to an empty
// knowledge base. Calling it once seeds the default scan roots and runs the
// scan synchronously, reducing the first-run funnel to install → scan →
// context instead of install → scan → star → refresh → use.
func registerScanTool(mcpServer *server.MCPServer, svc *service.Service) {
	mcpServer.AddTool(mcp.Tool{
		Name: "reponest_scan",
		Description: "Discover local Git repositories and populate the knowledge base. Call this once after install (or whenever reponest_context reports no projects) before anything else: it seeds the platform default scan roots on first run, then scans them synchronously and reports how many repositories/projects were found. " +
			"After it returns, call reponest_context to load a project's full working context.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc.EnsureDefaultScanRoots()
		res, err := svc.ScanNow()
		if err != nil {
			return makeTextResult("error: " + err.Error()), nil
		}
		return makeJSONResult("scan", map[string]any{
			"success":     res.Success,
			"repos_found": res.ReposFound,
			"projects":    res.Projects,
			"next_step":   "Call reponest_context to load a project's full working context.",
		})
	})
}
