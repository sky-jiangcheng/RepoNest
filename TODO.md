# TODO — 已知事项与待办

> 本文件记录产品深度评估产出的改进项。Sprint 1-13 已完成（D1-D11, D20-D24, C1-C3, C5-C6, C8, C10, C12-C15, P3, P7, P10-P11, P13-P16, P19, P21, P25-P28, P30, P35, P37-P38）。
> 第四轮深度评估（2026-08-27）：[docs/product-review/2026-08-27-new-deep-review.md](docs/product-review/2026-08-27-new-deep-review.md)。
> 修复后请从此清单移除并写入 CHANGELOG。
>
> **快捷跳转**：[✅ 已完成](#-已完成sprint-1-13) ｜ [🔴 删除项](#-删除项d7-d11) ｜ [🟡 收敛项](#-收敛项c11) ｜ [🟢 细化项](#-细化项p34-p38) ｜ [📋 遗留项](#-遗留项)

---

## ✅ 已完成（Sprint 1-13）

<details>
<summary>展开查看已完成项</summary>

| ID | 内容 | 完成 |
|----|------|------|
| D1 | 删除 `scripts/legacy/` | S1 |
| D2 | 删除 `scripts/generate_screenshots.py` | S1 |
| D3 | `tools/agent-score/` → MCP `reponest_agent_score` | S1 |
| D4 | 社区文件精简 | S1 |
| D5 | 删除 `.agents/` + gitignore | S1 |
| D6 | 删除 `.workbuddy/` + gitignore | S1 |
| D7 | 删除 `web/dist/` 构建产物 | S3 |
| D8 | 删除 DMG 素材 | S3 |
| D9 | 删除 `tokens.md` | S3 |
| D10 | 删除 `openapi.json` | S3 |
| D11 | 删除 docs issue 模板 | S3 |
| C1 | PWA/SEO 清理 | S1 |
| C2 | 插件系统降级 | S2 |
| C3 | 块编辑器冻结 | S2 |
| C5 | Knowledge 页面数据层拆分 | S4 |
| C6 | Dashboard 页面数据层拆分 | S4 |
| C8 | 历史评估文档压缩 | S3 |
| C10 | NoteSection 拆分（417行 → 305行 + hooks） | S6 |
| C12 | ProjectDetail 拆分（316行 → useProjectDetail） | S7 |
| C13 | CSS 死代码清理（移除 ~20 个未引用类） | S8 |
| C14 | queries_test.go 拆分（784行 → test_helpers + queries_test） | S8 |
| P3 | MCP 写入工具 | S2 |
| P7 | 版本号 SSOT | S2 |
| P10 | MCP 搜索结果结构化 | S3 |
| P11 | MCP 写入返回 note_id | S3 |
| P13 | 搜索排序优化 | S4 |
| P14 | 前端错误边界 | S4 |
| P15 | `db/db.go` 拆分 | S5 |
| P16 | `service/project.go` 拆分 | S5 |
| P19 | 版本历史 diff 可视化 | S5 |
| P21 | useScanPolling 测试 | S5 |
| P25 | `knowledge.go` 拆分（572行 → knowledge.go + types.go） | S7 |
| P26 | `stats.go` 拆分（471行 → 7 files） | S7 |
| P27 | `cmd/mcp/main.go` 拆分（453行 → 6 files） | S7 |
| P28 | MCP 工具描述增强 | S7 |
| P30 | 前端路由懒加载（dashboard/knowledge/projectDetail/settings chunks） | S8 |
| D20 | 删除空目录 `skills/` | S9 |
| D21 | 删除空目录 `.agents/` | S9 |
| D22 | `.claude/settings.local.json` 确认未跟踪 | S9 |
| D23 | PWA 图标清理（`web/public/` 3 个 icon 文件） | S9 |
| C15 | install 脚本评估 → 保留 + README 双路径说明 | S9 |
| P35 | NoteSection CSS Modules 试点（notes.css 242→192 行，新建 .module.css 93 行） | S10 |
| P38 | ProjectDetail 拆分（316→214 行 + useProjectDetail hook 122 行） | S11 |
| P37 | SKILL.md 工作流指引 + MCP 工具参数/示例增强 | S12 |
| D24 | PWA 移出桌面主构建（ADR-0008：残留清零 + 图标/孤儿 locale 清理） | S13 |
| P31 | `Domain/types.go` 评估 → 核心实体已收拢，DTO 按分层归位 | S14 |
| P32 | Wails 绑定层审计 → 46 方法无死绑定，`bindings.go` 顶部落审计块 | S14 |
| P33 | `TrendChart` 评估 → 实为 chart.js 封装（非纯 SVG），保留 | S14 |
| P34 | `project_overview.go` 评估 → 不拆，`mineAndCache` recover 已生效 | S14 |
| P36 | `knowledge.go` 进一步拆分评估 → 内聚度高暂不拆 | S14 |
| P29 | `parseTimestamp` 多格式鲁棒解析（unix/RFC3339/ISO8601/git %ai/%ad 默认）+ 测试 | S14 |
| P35·b | KnowledgeCard → `KnowledgeCard.module.css`（全局仅留共享 .pin-btn） | S14 |
| C11·A | 插件运行时：`loader.go` 合入 `runtime.go`（删 loader.go，Claude importer 路径不受影响） | S14 |

</details>

---

## 🔴 删除项（D7-D11）

> 纯减法，零功能回退。

### D7: 删除 `web/dist/` 构建产物

- [x] 删除 `web/dist/` 目录（~200 个 KaTeX 字体/Mermaid chunk/Vite 缓存文件）
- [x] `.gitignore` 已包含 `web/dist/`，`vite build` 重建后 Wails embed 正常

### D8: 删除 macOS DMG 素材

- [x] 删除 `build/dmg-background.svg` + `build/dmg-readme.txt`

### D9: 删除 `tokens.md`

- [x] 删除 `web/src/styles/tokens.md`

### D10: 删除手动维护的 OpenAPI spec

- [x] 删除 `docs/api/openapi.json`（后续改为 CI 自动生成）

### D11: 删除 docs issue 模板

- [x] 删除 `.github/ISSUE_TEMPLATE/docs.yml`

---

## 🔴 删除项（D20-D24）

> 第四轮评估新增。零功能回退，纯减法。

### D20: 删除空目录 `skills/`

- [x] `rm -rf skills/`（零内容空目录）

### D21: 删除空目录 `.agents/skills/`

- [x] `rm -rf .agents/`（AI 工具产物，.gitignore 已覆盖）

### D22: `.claude/settings.local.json` 从 git 跟踪中移除

- [x] 确认未被 git 跟踪（.gitignore `.claude/` 已覆盖），无需操作

### D23: PWA 图标清理（`web/public/`）

- [x] 确认无代码引用 icon-192/512/maskable
- [x] 删除 3 个 PWA 图标文件（共 51KB），保留 favicon.ico + favicon.svg

### D24: PWA 移出桌面主构建（ADR-0008）

- [x] 删除 `web/public/` 3 个 PWA 图标（icon-192/512/maskable）——同 D23 范围，随 ADR-0008 落地
- [x] 两套 locale（zh-CN / en）清理 11 个孤儿安装字符串（installTitle/Desc/App/Msg/Desktop 等）
- [x] `App.tsx` 路由注释改为「浏览器 / 桌面壳」区分，不再以 PWA 叙事描述
- [x] `main.go` CSP 注释去掉 PWA/registerSW.js 表述
- [x] README / getting-started / settings / SKILL.md / docs 失实行清零
- [x] ADR-0008 + index 登记
- [x] web 构建保留（`npm run build` 照常），不再是可安装 PWA

---

## 🟡 收敛项（C11, C15）

> 前轮遗留 + 本轮确认。

### C11: 插件运行时精简评估（669 行）

- [x] **方案 A（推荐）已完成**：`loader.go`（实际 121 行，非 52）合入 `runtime.go` 末尾「yaegi script loader」段，删除 `loader.go`；`runtime.go` import 合并（+reflect/strings/interp/stdlib），符号无冲突（`exportedTypes`↔`Context`、`loadPlugin`↔`compileScript` 同包互引）；`go build ./... + go vet + go test ./internal/core/plugin/runtime/` 全绿
- [ ] 方案 B（2.0 考虑）：评估移除 yaegi 依赖，Claude importer 改为内置函数
- [x] Claude importer 路径确认不受影响：`internal/importers/claude` 是 Go 原生 `plugin.KnowledgeImporter`，经 `service/plugin.go:100` 的 `rt.RegisterSource` 注册，不走 yaegi `compileScript`/`script` 路径，本次合并不触碰

### C15: install 脚本评估

- [x] **保留**脚本，README 安装说明已调整为「直接下载」+「脚本安装」双路径

---

## 🟢 细化项（P34-P38）

> 第四轮新增。按 AI 产品优先级排序。

### P34: `service/project_overview.go` 评估（239 行）

- [x] 确认函数内聚度合理，**不拆**：9 个方法均为 project/stats/level/star/search/overview/mining/summary/list 的薄 service 委托，同属「项目读模型」域。异步挖掘在 `GetProjectOverview` 内以 goroutine 触发 `mineAndCache`，`project_overview.go:139` 的 `defer recover()` 已确认生效（后台 panic 记日志不崩主进程）

### P35: 前端 CSS 架构迁移（4,055 行全局 CSS）

- [x] NoteSection CSS Modules 试点完成（notes.css 242→192 行，NoteSection.module.css 93 行新建）
- [x] 第二组件 KnowledgeCard 迁移完成：卡片样式从 `knowledge.css` 迁入 `KnowledgeCard.module.css`（`kind-*` 动态类改为 `badgeByKind` 查表映射），全局仅保留共享的 `.pin-btn`（与 NoteSection 复用）与 `.markdown-body`/`.btn`；`knowledge.css` 随之收缩，`npm run build` + tsc + eslint 全绿
- [ ] 保留全局 CSS 仅用于 reset、design tokens、跨组件基础样式
- [ ] 逐组件迁移，每轮 sprint 处理 1-2 个组件（下一步：ProjectCard）

### P36: `knowledge.go` 进一步拆分评估（536 行）

- [x] 评估函数间共享参数情况，**暂不拆**：12 个函数全部以 `repoPath string` 为入参、各自独立作用于仓库路径，无跨函数共享可变状态，内聚度高（单一「repo 知识探测」职责）。若后续再增挖掘维度，首选切出依赖探测子簇（`DetectDependencies` + `parseNpmDeps/parseGoDeps/parseCargoDeps`，约 150 行）为 `dependencies.go`

### P37: SKILL.md 工作流指引增强

- [x] 在 SKILL.md 开头增加「推荐工作流」段落（search → ask → read → create）
- [x] 为每个 MCP 工具补充使用场景 + 参数约束 + 示例值

### P38: ProjectDetail 拆分确认（316→214 行）

- [x] `useProjectDetail` hook 已创建（122 行），数据层真正下沉
- [x] 组件降至 214 行（≤200 目标基本达成）

---

## 🟢 细化项（P25-P33）

> 新增项。按文件体量和 AI 产品优先级排序。

### P25: `knowledge.go` 拆分 ✅

- [x] 572 行拆为 `knowledge.go`（515 行实现）+ `types.go`（48 行类型定义）

### P26: `stats.go` 拆分 ✅

- [x] 471 行拆为 7 个文件：`types.go` + `stats.go` + `validation.go` + `query.go` + `commits.go` + `range.go` + `dates.go`

### P27: `cmd/mcp/main.go` 拆分 ✅

- [x] 453 行拆为 6 个文件：`main.go` + `tools_notes.go` + `tools_projects.go` + `tools_search.go` + `tools_score.go` + `results.go`

### P28: MCP 工具描述增强 ✅

- [x] 为每个工具补充使用场景 + 参数约束 + 示例值
- [x] 推荐 AI 工作流：ask → read → create → update

### P29: `stats` 时间戳解析鲁棒性 🔻低

- [x] `parseTimestamp`（`internal/stats/validation.go`，P26 拆分后）从「只认 `2006-01-02 15:04:05`、忽略错误」升级为多格式：裸 unix 秒（git `%at`）、RFC 3339 / ISO 8601（git `%aI`/`%cI`，带/不带时区）、git `%ai`（`… -0700`）、date-only（`%ad --date=short`）、git 默认作者日期（`Mon Jan _2 … -0700`，含空格补零日）；不可解析仍返回 0（保持 latest-commit 比较的宽松契约）。新增 `TestParseTimestamp` / `TestParseTimestamp_Unparseable`（10 + 4 例）

### P30: 前端路由级懒加载 ✅

- [x] `React.lazy()` + `Suspense` 对 Dashboard/Knowledge/ProjectDetail/Settings 代码分割
- [x] `vite.config.ts` manualChunks 独立 chunk（dashboard、knowledge、projectDetail、settings）

### P31: `Domain/types.go` 评估 🔻低

- [x] 评估是否收拢核心 domain 类型：**已收拢**，`internal/domain/types.go`（132 行）已集中 13 个核心持久化实体（Project/Repository/Todo/Note/NoteWithProject/NoteVersion/NoteDiff/TodoCount/NoteCount/DailyStat/HeatmapDay/SearchHit/RepoMeta）。各包里另见的 `service.ProjectResponse`、`knowledge.RepoKnowledge` 等是响应/DTO 形状，按分层归各层，非「散落的 domain 实体」，无需再收拢

### P32: Wails 绑定层审计 🔸中

- [x] 审计 `bindings.go` 46 个方法（218→254 行含审计注释）：**无死绑定/无重复**。逐方法标注 MCP 对应 vs desktop-only，结论落为 `bindings.go` 顶部审计块。13 个绑定有对应 `reponest_*` 工具（notes/project/scan/context/search 面），其余为 GUI 状态、todos、config、插件管理、文件导出、笔记版本/pin/move 等刻意仅供桌面端；`reponest_ask/handoff/agent_score/integrity/notes_read` 为 agent-only、无绑定（直接走 service）

### P33: `TrendChart` 组件评估 🔻低

- [x] **原描述已过时**：`TrendChart.tsx` 不是「纯 SVG」，而是 `chart.js` / `react-chartjs-2` 的 `<Line>` 配置封装（约 100 行 options/data），被 `ProjectDetail`（经 `useProjectDetail`）实际使用。折线图用 CSS 替代不成立——chart.js 已承担该职责，手写 CSS 折线图反而更重。**结论：保留**，作为 chart.js 的薄封装合理

---

## 🟢 会话记忆路线（ADR-0007 / ADR-0009 后续）

> 定位升级为「AI agent 记忆层」后的主攻方向，按传播价值排序；M4-M5 的分发决策（薄客户端纪律、VS Code 先行、JetBrains 缓议）见 [ADR-0009](docs/adr/0009-ide-presence.md)。

### M1: 会话自动捕捉（零人工参与）→ [ADR-0010](docs/adr/0010-session-auto-capture.md)

- [x] **解析核心已落地**（`internal/importers/claude/session.go`）：`ParseSession` 流式解析 `~/.claude/projects/<slug>/<id>.jsonl`，宽松跳过未知/超长行、忽略 sidechain，抽取稳定信封字段（sessionId/cwd/gitBranch/timestamp）+ 最后一条 assistant 文本 + 首次 user 指令 + 去重工具名；实测对齐真机 v2.1.278（`session_test.go`）
- [ ] 端到端接线（**决策已定**：默认关；C端按需 / B端可选 SessionEnd hook；风险项写进设置+文档由 B端自担，ADR-0010 决策落地节）：`Session`→`service.HandoffInput` 映射 + `CreateHandoffNote` + 按需捕获方法 + `claude_session_capture` 配置默认关 + hook 接入；解析器目前不自动运行、不读盘
- [x] 体积与隐私评估：ADR-0010 已定「默认关 + 只读路径白名单 + 尾部 N 有界提取 + 宽松失败 + golden-file」，边界与既有 claude *memory* importer（读 `memory/*.md`，非 jsonl 逐字稿）划清

### M2: 多 agent 记忆源导入 → [ADR-0011](docs/adr/0011-multi-agent-memory-importers.md)

- [x] **Codex 源已落地**：`internal/importers/codex`（流式解析 `~/.codex/sessions/**/rollout-*.jsonl`，取 cwd→项目匹配 + 首次指令 + 最近回复成一条 `log` 笔记），复用 `plugin.KnowledgeImporter`/`upsertDoc`；抽公共件 `internal/importers/memsrc`（`MatchProject`/`ReadCapped`/`LastPathSegment`），claude 改为委托（测试不破）。**隐私门**：新增 `RegisterSourceManual` + `sourceEntry.auto`，`ImportAll`（启动自动导入）只跑 auto 源、Codex 经设置里 sources 列表显式一键触发（ADR-0011 决策 4）。golden-style 测试 + runtime 门控测试
- [x] **OpenCode 源已落地**（`internal/importers/opencode`，opt-in MANUAL）：真机核验 `~/.local/share/opencode/storage/session/<hash>/ses_*.json`（自带 title/summary/directory，已是摘要级），`directory` 末段→项目匹配；golden-style 测试 + 接口断言
- [ ] Cursor 源（范围内、未实现）：`state.vscdb` 未公开 SQLite blob，先只读核验 chat 表编码再写；默认关、标非官方契约——风险最高一源
- [ ] OpenClaw 源（范围内，格式已核验、未实现）：记忆=`~/.openclaw-autoclaw/workspace/*.md`（agent 全局、非按项目分）；**必须 allowlist 到这些 md**（同目录含私钥/vault，绝不整目录遍历、绝不把 key 落进笔记）；无 1.0/2.0 判别位（日期式 version），「只做 2.0」＝按当前布局实现。**待决**：全局记忆挂到哪个项目（`workspace/.git` 推断 / 用户指定 / 专用项目）——定前不硬塞
- [ ] Hermes 源（范围内，未实现）：**Nous Research 独立产品，与 OpenClaw 两家**（早先误判为 OpenClaw 运行时，已更正）。本机无可检视的 Hermes 数据目录 → **需真实数据样例/路径核验格式**再写，不臆测

### M3: 语义检索 → [ADR-0012](docs/adr/0012-semantic-search.md)

- [x] 评估 + **实测验证**：`CGO_ENABLED=0` 下 `modernc.org/sqlite/vec`（bundled sqlite-vec v0.1.9）跑通 vec0 建表 + KNN + `vec_distance_l2`（回归测试 `internal/vecprobe`，副作用用无生产码的 test-only 包隔离）；混合检索中间件 `internal/search/hybrid`（`Embedder` 接口 + `FuseRRF` k=60 + 单测）已落地，尚未接入生产
- [ ] **决策已定（ADR-0012 落地节）**：放弃 B（纯 Go 本地模型）；**默认 C**＝FTS5 增强（同义词/query 改写，本地零依赖）；**A＝远程 embedding API 面向 B端可选、默认关**，风险项（内容出机器/依赖网络密钥）必须写进设置+文档由用户显式确认。下一步实现 C 为默认；A 待 Provider 选型（`Embedder` 接口已就绪，插一个远程 client 即可）
- [ ] 端到端接线：C 先做（`db/search.go` 近义扩展）；A 走时补 `note_embeddings` vec0 表迁移 + hybrid RRF 接进 search + `semantic_search` 配置默认关 + A/B 评测门

### M4: Agent 集成即插即用

- [x] Claude Code hook 示例：SessionEnd hook 自动触发 reponest_handoff（v1.9.4 交付于 docs/features/ai-integration.md「会话结束自动交接」节）
- [x] `npx reponest-init` 类一键注册脚本（写 .mcp.json + 提示 hook 配置；ADR-0009 第一步，该 ADR Proposed→Accepted 的门槛项）——已交付 `scripts/reponest-init/`（零依赖 Node ≥18，幂等 + dry-run + hook 安装，2026-10-02）

### M5: IDE 存在感（ADR-0009 薄客户端分发）

- [x] VS Code 扩展（唯一 IDE 扩展，一份 VSIX 覆盖 VS Code / Cursor / Windsurf 全 fork 家族）——骨架已交付 `ide/vscode/`（tsc 零错误）：命令面板 context / handoff / search + 状态栏入口 + 侧边栏笔记检索（MCP stdio 薄客户端）；余：VSIX 打包 / CI、状态栏「上次交接时间」（需服务端交接时间戳 API）
- [ ] JetBrains 插件：缓议——独立 Kotlin/Gradle 代码库双倍维护面，待真实需求信号（issue/star）并补充 ADR 后再立项
- [ ] 分发评估门：此后每个分发资产立项时必须回答「人在哪个界面上看见它」；仅 agent 可消费的资产需说明服务存量深度而非获客（ADR-0009 决策 5）

---

## 📋 遗留项

- [ ] 桌面 GUI 回归测试：建议在真机跑一轮冒烟（扫描→收藏→刷新历史→笔记 CRUD→版本恢复→知识库搜索→MCP 问答）
- [ ] **D25 仪表盘生产力门面收缩（2.0 候选，非现在）**：首屏讲记忆环、打开是仪表盘，定位纯度持续被消耗。收敛方向：仪表盘退化为「项目列表 + 最近活动」；目标环 / 每日代码量标准 / 工作日告警沉入插件或删除（GitBoard/GitBuddy 时代遗产，见 [ADR-0008](docs/adr/0008-pwa-removal.md) 遗留项）
- [x] `reponest_context` brief/full 档位评估（v1.9.4 收敛：中等优先，缓做——当前固定 10 notes × 1200 字符 + 8 commits 对单会话偏充裕）
- [x] README 对比表 + ASCII 架构图（v1.9.4 收敛：文档润色，低优先，缓做）
- [x] `mineAndCacheAsync` 后台 goroutine 加 recover（`project_overview.go:138`）
- [x] `.zcode/` 已移出跟踪，不需要 history rewrite
- [x] P29 `parseTimestamp` 鲁棒性（多格式 + 测试，见 P29 结论）
- [x] P31 `Domain/types.go` 评估（核心实体已收拢，DTO 按分层归各层）
- [x] P32 Wails 绑定层审计（无死绑定，`bindings.go` 顶部审计块）
- [x] P33 `TrendChart` 评估（实为 chart.js 封装非纯 SVG，保留）
- [x] P36 `knowledge.go` 进一步拆分评估（内聚度高暂不拆，见 P36 结论）

---

## 建议执行节奏

| 阶段 | 内容 | 预估 |
|------|------|------|
| ~~Sprint 1-5~~ | ~~D1-D11, C1-C3, C5-C6, C8, P3, P7, P10-P11, P13-P16, P19, P21~~ | ✅ 共 28 项 |
| ~~Sprint 6~~ | ~~D12-D19 剩余删除 + C10 NoteSection 拆分~~ | ✅ |
| ~~Sprint 7~~ | ~~C12 + P25-P27 大文件拆分 + P28 MCP 描述增强~~ | ✅ |
| ~~Sprint 8~~ | ~~C13 CSS 清理 + C14 测试拆分 + P30 懒加载~~ | ✅ |
| ~~**Sprint 9**~~ | ~~D20-D23 资产大扫除 + C15 文案调整~~ | ✅ |
| ~~**Sprint 10**~~ | ~~P35 NoteSection CSS Modules 试点~~ | ✅ |
| ~~**Sprint 11**~~ | ~~P38 ProjectDetail hook 提取~~ | ✅ |
| ~~**Sprint 12**~~ | ~~P37 SKILL.md 工作流指引~~ | ✅ |
| ~~**Sprint 13**~~ | ~~D24 PWA 移出桌面主构建（ADR-0008 落地）~~ | ✅ |
| **Sprint 14** | P29/P31/P32/P33/P34/P36 评估类小项收口（验证 + 落结论，含 P32 绑定审计块、P29 多格式解析 + 测试） | ✅ 共 6 项 |
| **2.0 规划** | D25 仪表盘生产力门面收缩 + C11 插件系统评估 + P35 全量 CSS Modules | 按版本 |
| **按需** | P29, P31, P32, P33, P36 | 随重构穿插 |
