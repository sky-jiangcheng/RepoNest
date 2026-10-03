# ADR-0013: 本地向量存储选型（sqlite-vec）与安装引导

- 状态：Accepted（本地 sqlite-vec 默认 + `VectorStore` 接缝已落地；`cmd/vector-init` 引导 + 远程 Qdrant 可选、不可达自动退回本地，均已实现）
- 日期：2026-10-03
- 关联：[ADR-0003](0003-fts5-search.md)（FTS5）、[ADR-0012](0012-semantic-search.md)（语义检索：Embedder + RRF 融合）、TODO M3

## 背景

M3 语义检索需要「把向量存起来 + 做相似度检索」的存储层，与 FTS5 结果经 RRF 融合。本 ADR 由用户选型文档引入，但据本仓 2026-10-03 的实际进展做了两点**事实校正**（原选型文档的过时/错误处见下）。

**先厘清两个独立的「本地 vs 远程」轴**（原稿把二者混在一起）：
- **轴 A · Embedding provider**（文字→向量怎么算）：本地（自托管 Ollama）/ 远程（OpenAI 兼容 API）。→ 已由 `hybrid.Embedder` + `hybrid.RemoteEmbedder` 支持（ADR-0012），base_url 指向 Ollama 即本地、指向云 API 即远程。
- **轴 B · 向量存储**（向量存哪、怎么 KNN）：本地嵌入库 / 远程向量数据库服务。→ 本 ADR 决策 = **本地**。

## 事实校正（相对用户选型原稿）

1. **sqlite-vec 不需要 CGO。** 本仓用 `modernc.org/sqlite/vec`——sqlite-vec 的**纯 Go 移植**，经 `sqlite3_auto_extension` 装载。已在 `CGO_ENABLED=0` 下 `go build ./...` 与 `internal/vecprobe` 的 vec0 建表/KNN/`vec_distance_l2` 实测通过。因此**不需要** `//go:build vec` 标签、不需要静态链接 C 扩展。
2. **向量存储层已落地、非「尚未」。** `internal/db/vecindex.go` 已实现 vec0 派生索引（`EnsureVectorIndex(dim)` 按 dim 自管重建、`Put/Delete/KnnNoteIDs/Clear/Drop`），并已接进默认关的 A 检索融合（ADR-0012）。
3. 因第 1 点，**`chromem-go`（原为「避免 CGO」的纯 Go 备选）的动机消失**——modernc sqlite-vec 已是纯 Go；且它能与结构化数据同库同事务，优于另起一套 chromem 存储。故降为不采用。

## 决策

**轴 B 选：sqlite-vec（`modernc.org/sqlite/vec`，纯 Go）作为唯一默认向量存储。**

1. **同库同文件**：向量存进既有 `dashboard.db` 的 `vec0` 虚拟表 `note_embeddings`（rowid=note id），与 notes / FTS5 同库、同事务、同备份。
2. **派生缓存**：向量是可从 `project_notes` 全量重算的缓存（`RebuildEmbeddings`），换 embedding 模型/维度即 drop+重建；SQLite 主库仍是唯一事实源。dim 记在自管 `note_embeddings_meta`，不读 sqlite-vec `_info` 影子表列名（跨版本不稳）。
3. **默认关**：语义检索仅 `semantic_search=1` 且向量索引就绪且配好 embedding provider 才生效；任何未配/失败一律优雅退回纯 FTS5（ADR-0012）。
4. **远程向量库经 `VectorStore` 接缝以 Qdrant 实现、默认本地 + 自动退回**：`internal/search/vectordb` 定义 `Store` 接口（`Ensure/Clear/Upsert/Search`），两实现——`Local`（sqlite-vec，默认）与 `Qdrant`（REST：建集合 / 写点 / `points/search`，`api-key` 可选）。`vectordb.Open` 按 `vector_store`(local|qdrant)/`vector_store_url`/`vector_store_api_key`/`vector_store_collection` 选择；**选了 qdrant 但未配/不可达则静默退回本地**。默认仍本地（个人库够用），远程面向 >百万向量/多端共享的 B 端；换 Weaviate 等只需再加一个 `Store` 实现。REST 契约带 build-tag 门控的**真实服务冒烟测试**（`ollamalive`/`qdrantlive`/`aelive`，CI 默认不跑），已在本地真 Ollama + 真 Qdrant 容器跑通（含全链路语义召回）。云端 Qdrant / 其它厂商未测。

## 安装引导（新增，落地「引导用户到设置」）

`cmd/vector-init`（对外的 `reponest vector-init`）——**默认本地、引导式、远程为显式可选**：

1. 环境自检：SQLite 版本 + `vec_version()`（确认 vec0 扩展已装载）。
2. 存储层健康检查：建 vec0 → 写测试向量 → KNN 往返 → 清理，逐条打勾（就是本 ADR 决策 1/2 的最小验证）。
3. Embedding provider 选择（轴 A）：
   - `[1]` **本地 Ollama（默认推荐，离线）**：base_url `http://localhost:11434/v1`，`nomic-embed-text`/`bge-m3`；
   - `[2]` **远程 OpenAI 兼容 API**（联网、内容出机器，ADR-0012 风险项）；
   - `[3]` 跳过，稍后在设置里配。
   选定后写入 `embedding_*` 配置（**api key 走脱敏存储，不回传前端**）。
4. 收尾**指向设置页**：打印「去 设置 → 插件 复核/调整 provider、打开 `语义检索` 开关、点『重建索引』」——即用户要的「引导到向量库设置」。`semantic_search` **保持默认关**，由用户在设置里显式开。
5. 远程向量库：`cmd/vector-init -store qdrant -store-url <url> [-store-api-key <k>]` 写入 `vector_store*` 配置并探测；不可达自动退回本地并如实打印。契约正确性由 `qdrantlive`/`aelive` 冒烟测试背书（本地真服务已跑）。

## 理由

- 单二进制 / 同库事务 / 已有 FTS5 / 零 CGO / 规模够用——sqlite-vec 是唯一同时满足全部产品约束的选项；原稿列的独立服务全都要额外进程与网络，违背本地优先定位。
- 「默认本地 + 引导到设置 + 远程确认后可选」把易用性（开箱即用）与灵活性（B 端确有远程诉求时可控切换）分开，且把语义检索的**风险项（内容出机器）放在显式开关之后**，符合 ADR-0010/0012 一贯的「敏感能力默认关、文档化」纪律。

## 后果

- 正面：向量能力零新依赖、零 CGO、与主库同生命周期；安装引导让本地默认开箱可用、设置可发现。
- 负面：单库向量规模有上限（个人库远未触及）；真实服务冒烟测试靠 build-tag 门控、CI 默认不跑（云端 Qdrant 未覆盖，接云端时补验）。
- 待决：更多远程后端（Weaviate/Pinecone）按同一 `Store` 接缝再加；embedding 与 key 在纯云场景的更严格管理（如系统 keychain）视需求。
