//go:build ollamalive

// Opt-in integration test: runs my OpenAI-compatible RemoteEmbedder against a
// real Ollama server (default http://localhost:11434/v1), verifying Ollama's
// /v1/embeddings envelope matches the client. Run:
//
//	ollama serve &           # then: ollama pull nomic-embed-text
//	go test -tags ollamalive ./internal/search/hybrid/ -run Live -v
package hybrid

import (
	"os"
	"testing"
)

func TestOllamaEmbedLive(t *testing.T) {
	base := os.Getenv("OLLAMA_URL")
	if base == "" {
		base = "http://localhost:11434/v1"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "nomic-embed-text"
	}
	e := &RemoteEmbedder{BaseURL: base, Model: model} // Dim 0 = learn from response
	out, err := e.Embed([]string{"hello world", "another sentence"})
	if err != nil {
		t.Fatalf("embed against %s (%s): %v (Ollama running? model pulled?)", base, model, err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d vectors, want 2", len(out))
	}
	if len(out[0]) == 0 {
		t.Fatal("empty vector returned")
	}
	t.Logf("live Ollama OK: model=%s dim=%d", model, len(out[0]))
}
