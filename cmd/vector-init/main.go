// Command vector-init is RepoNest's vector-store onboarding guide (ADR-0013).
// It defaults to the LOCAL sqlite-vec store (already the app's storage layer),
// verifies it works with a health check, records the chosen embedding provider,
// and then points the user at Settings → Plugins to review / enable / rebuild.
//
// Design: default local; remote is guided but explicitly opt-in.
//   - Embedding provider (axis A): [1] local Ollama (default, offline) |
//     [2] remote OpenAI-compatible API | [3] skip.
//   - Vector store (axis B): local sqlite-vec is the decision. A REMOTE vector
//     database (Qdrant/Weaviate/Pinecone) is a future VectorStore seam for
//     >1M-vector scale and is NOT implemented here — vector-init only notes it.
//
// It never enables semantic search: `semantic_search` stays OFF until the user
// flips it in Settings (ADR-0012: enable only after the cmd/abeval gate).
//
// Usage:
//
//	reponest vector-init                      # local Ollama default
//	reponest vector-init -provider openai -api-key $KEY
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"repo-nest/internal/db"
	"repo-nest/internal/platform"
	"repo-nest/internal/search/hybrid"
	"repo-nest/internal/service"
)

type providerDefaults struct {
	baseURL  string
	model    string
	dim      int
	needsKey bool
}

var providers = map[string]providerDefaults{
	"ollama": {"http://localhost:11434/v1", "nomic-embed-text", 768, false},
	"openai": {"https://api.openai.com/v1", "text-embedding-3-small", 1536, true},
}

func main() {
	log.SetFlags(0)
	dbPath := flag.String("db", "", "database path (default: the app's DB)")
	provider := flag.String("provider", "ollama", "embedding provider: ollama | openai | skip")
	baseURL := flag.String("base-url", "", "override provider base URL")
	model := flag.String("model", "", "override embedding model")
	dim := flag.Int("dim", 0, "override embedding dimension (0 = provider default)")
	apiKey := flag.String("api-key", "", "API key for a remote embedding provider (stored masked; never shown to the UI)")
	skipProbe := flag.Bool("skip-probe", false, "don't test the embedding endpoint reachability")
	flag.Parse()

	path := *dbPath
	if path == "" {
		path = platform.GetDbPath()
	}
	database, err := db.InitDB(path)
	if err != nil {
		log.Fatalf("vector-init: open db: %v", err)
	}
	defer database.Close()
	svc := service.New(database, "vector-init")

	// 1. Environment: SQLite + vec extension.
	var vecVersion string
	if err := database.QueryRow("SELECT vec_version()").Scan(&vecVersion); err != nil {
		log.Fatalf("vector-init: sqlite-vec not available (%v) — this build must link modernc.org/sqlite/vec", err)
	}
	fmt.Printf("✓ 向量扩展: sqlite-vec %s (纯 Go, 无 CGO)\n", vecVersion)

	switch *provider {
	case "skip":
		fmt.Println("• Embedding provider: 跳过（稍后在 设置 → 插件 配置）")
		healthOrDie(database, defaultDim("ollama")) // still verify the local store works
		printNextSteps()
		return
	case "ollama", "openai":
	default:
		log.Fatalf("vector-init: unknown -provider %q (want ollama|openai|skip)", *provider)
	}

	def := providers[*provider]
	bu := firstNonEmpty(*baseURL, def.baseURL)
	md := firstNonEmpty(*model, def.model)
	d := *dim
	if d <= 0 {
		d = def.dim
	}
	key := *apiKey
	if key == "" && *provider == "openai" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if def.needsKey && key == "" {
		fmt.Println("⚠ 远程 provider 未提供 -api-key/OPENAI_API_KEY；语义检索将无法生效，直到在设置里补上")
	}

	// 2. Store health check (creates the vec0 index at dim + verifies a KNN round-trip).
	healthOrDie(database, d)

	// 3. Embedding provider reachability (best-effort; warn, never hard-fail).
	if !*skipProbe {
		emb := &hybrid.RemoteEmbedder{BaseURL: bu, Model: md, APIKey: key, Dim: d}
		if _, err := emb.Embed([]string{"vector-init probe"}); err != nil {
			fmt.Printf("⚠ Embedding 服务暂不可达（%v）—— 安装后仍需 provider 就绪再『重建索引』\n", err)
		} else {
			fmt.Printf("✓ Embedding 服务: %s (%s, dim=%d)\n", *provider, md, d)
		}
	}

	// 4. Persist provider config (semantic_search deliberately left OFF).
	for k, v := range map[string]string{
		"embedding_base_url": bu,
		"embedding_model":    md,
		"embedding_dim":      fmt.Sprintf("%d", d),
	} {
		if err := svc.UpdateConfig(k, v); err != nil {
			log.Fatalf("vector-init: save %s: %v", k, err)
		}
	}
	if key != "" {
		if err := svc.UpdateConfig("embedding_api_key", key); err != nil {
			log.Fatalf("vector-init: save api key: %v", err)
		}
	}
	fmt.Println("✓ 向量索引: 已就绪（本地 sqlite-vec，存于 dashboard.db）")
	printNextSteps()
}

func healthOrDie(database *sql.DB, dim int) {
	if err := db.VectorStoreHealthCheck(database, dim); err != nil {
		log.Fatalf("vector-init: 向量存储健康检查失败: %v", err)
	}
	fmt.Printf("✓ 向量存储: 本地 sqlite-vec 健康检查通过（vec0 建表 + KNN 往返, dim=%d）\n", dim)
}

func printNextSteps() {
	fmt.Println(`
下一步（引导）：
  1. 打开 设置 → 插件：确认 embedding provider / 维度，必要时改。
  2. 打开「语义检索」开关（默认关，需你先跑 A/B 门：go run ./cmd/abeval -cases queries.jsonl）。
  3. 点「重建索引」把现有笔记写入向量索引。
远程向量库：当前不实现（本地 sqlite-vec 足够个人库）；确需 >百万向量时再按 ADR-0013 选厂商接 VectorStore 接缝。`)
}

func defaultDim(p string) int { return providers[p].dim }

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
