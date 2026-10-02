# ADR-0012: 语义检索评估——向量召回补 FTS5 字面盲区，守住零 CGO（M3）

- 状态：Proposed（评估结论：向量侧已可零 CGO 落地，但本地 embedding 生成是唯一硬门；未实现）
- 日期：2026-10-02
- 关联：[ADR-0003](0003-fts5-search.md)（FTS5 trigram 全文检索）、[ADR-0006](0006-scope-freeze.md)、[ADR-0007](0007-session-memory-protocol.md)、TODO 会话记忆路线 M3

## 背景

现有检索是 **FTS5 trigram + bm25**（`internal/db/search.go`、`internal/db/migrate.go` 建 `project_notes_fts`/`project_todos_fts`），补 LIKE 兜底。trigram 对**子串/拼写/部分词**很稳，但结构性地缺**语义**：搜「内存泄漏」召回不到只写了「heap 持续增长、GC 压力大」的笔记——词面不重叠。M3 想补这层召回。

关键约束来自产品形态：RepoNest 是本地优先的 Wails 桌面 App，跨 mac/win 分发，**数据库驱动是 `modernc.org/sqlite`（纯 Go、零 CGO）**（`go.mod`、`internal/db/db.go`）。任何检索增强都不得把 CGO 重新引入构建（会直接炸掉跨平台打包与体积假设）。

## 决策 / 评估结论

1. **向量存储与检索：零 CGO 这条路现在已成立（2026-09 更新的事实）**。`modernc.org/sqlite` 现已提供 **`modernc.org/sqlite/vec`**——纯 Go 移植、sqlite-vec 兼容的向量虚拟表/索引（KNN、`vec_distance_*`）。即「ANN 索引 + 相似度查询」这一半**不需要**换驱动、不需要 CGO。结论：向量侧技术可行性从「存疑」上调为「可行」，且与现有 FTS5 同库共存（一张 `vec0` 虚拟表 + 外键回连 `project_notes.id`）。
2. **真正的硬门是本地 embedding 生成，且它才是默认关的理由**。要把笔记/query 变成向量，需要一个推理运行时；而主流本地方案（ONNX Runtime、llama.cpp、sentence-transformers 后端）普遍要 CGO 或附带大体积模型——这与「零 CGO + 桌面轻分发」正面冲突。M3 是否值得做，**几乎完全取决于能否找到零 CGO 的 embedding 路径**，需在下列选项中拍板：
   - A：**可选远程 embedding API**（零 CGO、零模型捆绑，但违背「纯本地」叙事且引入网络依赖/隐私面）；
   - B：**纯 Go 小模型推理**（若有满足质量阈的小 embedding 模型可在纯 Go 下跑；需先做技术验证，未知数最大）；
   - C：**放弃向量、只增强 FTS**（如 synonym 词表 / 查询改写，零新依赖，但天花板低）。
3. **落地形态（若通过）：混合检索 + RRF 融合 + 默认关**。FTS5 与向量各出一路排序，用 Reciprocal Rank Fusion 合并，向量路默认不启用、由设置开关控制。理由：字面匹配（trigram）在多数日常查询里已够且更快，向量只在「换词搜」场景增益；混合保证不劣化基线。
4. **先评测后开关（A/B 门，实现前置）**：建一个小型标注评测集（真实笔记 + 人工写的「应命中」query 对），脚本化比较 FTS-only vs FTS+vec 的 recall@k / nDCG；**达不到明确质量增益阈值就不改默认**、不引入 embedding 依赖。评测产物与结论回写本 ADR。
5. **数据可重建、非真相源**：embedding 与向量索引视为派生缓存，SQLite 主库仍是唯一事实源；向量表可随时 drop + 依据笔记重算，允许换模型/升版本，不参与备份真相。

## 理由

- 向量侧的可行性问题已被 `modernc.org/sqlite/vec` 这一外部事实解决，值得如实记录，避免团队继续按「sqlite-vec 需要 CGO」的旧认知否决它。
- 但把「能不能存向量」和「能不能零 CGO 地**产出**向量」分开看，才是 M3 的真问题：后者才是与产品定位（本地、轻量、跨平台）冲突的点，因此默认关、先 A/B 是诚实且低风险的推进方式。
- 混合 + RRF + 派生缓存：对现有 FTS5 只增不改、可随时回退，符合 ADR-0006 的克制。

## 后果

- 正面：语义召回可补上 trigram 的换词盲区，且**不**以牺牲零 CGO 为代价换取向量存储。
- 负面/风险：本地 embedding 的模型体积/推理时延/质量三者权衡尚无定论；引入向量表带来重算与版本管理成本。
- **待决（实现前需回答）**：① embedding 生成走 A/B/C 中哪条路（决定 M3 成败，先做 B 的可行性 spike）；② A/B 评测的质量增益阈值与评测集来源；③ 默认关下的首个目标用户场景；④ 模型/维度升级时向量表的重算策略与迁移成本。
