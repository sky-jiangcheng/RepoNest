# ADR-0011: 多 agent 记忆源导入——可复用的 importer 框架与各源可行性（M2）

- 状态：Proposed（框架已存在；**Codex、OpenCode 已落地**为 opt-in 手动源；Cursor、OpenClaw、Hermes 已纳入范围，各有待决/待验证门）
- 日期：2026-10-02
- 关联：[ADR-0002](0002-c-end-repositioning.md)（进程内插件运行时）、[ADR-0007](0007-session-memory-protocol.md)、[ADR-0010](0010-session-auto-capture.md)、TODO 会话记忆路线 M2

## 背景

RepoNest 已有一条干净的「外部记忆 → 知识库笔记」管线：`plugin.KnowledgeImporter` 产出 `[]plugin.ImportDoc`（`ProjectID/Title/Content/Tags/Kind/Source`，见 `internal/core/plugin/plugin.go`），经 `Runtime.RegisterSource`（`internal/core/plugin/runtime/runtime.go`）注册为内置源，导入时由 `TriggerImport → upsertDoc` 以 `(project, source, title)` 幂等 upsert 成笔记（`internal/service/plugin.go:100` 把 Claude 源接进去）。目前只有**一个**具体源：`internal/importers/claude`（读 `~/.claude/projects/*/memory/*.md`）。

M2 要把同一管线扩展到 Cursor / Codex / OpenCode 等。但 2026-10 现状盘点显示这些源的落盘形态**差异极大且多数非公开契约**：

- **Codex CLI**：会话以逐行 JSON 事件落盘（与 Claude Code 的 `*.jsonl` 同类，尾部 assistant 消息 + tool 调用），相对可解析，但路径/字段随版本漂移。
- **OpenCode**：会话/历史结构有第三方解析项目公开映射（claude-replay、obsidian-aisync、cli-continues 等），可按其已知形状适配。
- **Cursor**：聊天/Agent 历史主要存在 **`state.vscdb`（一个未公开的 SQLite 库里的序列化 blob）**，非文档化、跨 workspace、随版本重构——是四者中最脆、最不官方的一种。

结论：M2 的真正风险不是「写解析器」，而是**给未公开、易漂移的外部格式背长期维护债**。

## 决策

1. **不新建机制，复用现有 importer 管线**。每个新源都是一个实现 `KnowledgeImporter` 的 Go 原生 importer，走 `RegisterSource` + `upsertDoc`，与 Claude 同构（幂等、`Source` 字段区分来源、匹配到已有 project）。`ImportClaudeMemory` 已是范式样板。
2. **抽一层薄公共件，而非抽象大框架**：把「按 project 匹配目标 ProjectID」「ImportDoc 构造 + kind/tag 约定」「逐文件宽松失败」从 claude.go 里提出为 `internal/importers` 下的小 helper，供各源复用。**只在第 2、3 个源真的落地时才抽**，避免为想象中的通用性提前抽象。
3. **按「格式稳定性 × 获客价值」排序推进**，一次只立一个源，每个源立项前先做一次「拿真机样本验证格式」的门禁：
   - 先 **Codex**（与 Claude jsonl 最近、价值高）；
   - 再 **OpenCode**（有公开形状可循）；
   - **Cursor 缓行**：`state.vscdb` 属逆向、易碎，除非有明确需求信号，单独立项并在文档显著标注「非官方、随时可能失效」。
4. **每个源默认关、显式开 + 可 dry-run**：设置里逐源开关，导入前可预览将产生的笔记条数/标题；失败源在 `SourceStatus`（`runtime.go`）里如实报错，不静默吞。
5. **golden-file 回归锁格式**：每个源带一组「真实样本 → 期望 ImportDoc」的测试（Claude 已有 `claude_test.go`），上游改格式时测试先红，避免线上静默产出脏笔记。

## 理由

- 复用现成管线使每个新源是「一个 importer 文件 + 一份测试」的增量，不碰 service/app/domain 分层与冻结面（ADR-0006）。
- 反爬式解析未公开格式是纯维护负债；用「先验证样本再立项、一次一个、默认关」把风险面切到最小，Cursor 用最保守处置。
- golden-file 把「格式漂移」从线上事故降级为测试失败——对易变外部源这是唯一可持续的工程手段。

## 后果

- 正面：多 agent 记忆可汇入同一知识库与 `reponest_context`；新增源成本随公共件下沉递减。
- 负面：+N 个外部格式适配器，每个都绑定上游版本；需持续维护真机样本测试。
- **待决**：① project 匹配规则是否对所有源统一（Claude 现按路径/slug 匹配目标项目，Codex/Cursor 的项目边界未必同名）；② Cursor 是否值得做，或改走「让用户导出为 markdown 再导入」的离线兜底；③ 是否需要一个「记忆源」设置页的统一开关面板。

## 落地进度与各源真机核验（2026-10-02）

产品决策：M2 **全部源都做**（覆盖 claude/codex/opencode/cursor/openclaw）。逐源按 ADR-0011 决策 3 的真机格式核验结果：

- **claude**（已有）：读 `~/.claude/projects/*/memory/*.md`，AUTO 源（人工整理的记忆文档，可启动自动导入）。
- **codex**（已落地，opt-in MANUAL）：`~/.codex/sessions/**/rollout-*.jsonl` 流式解析，每会话→一条 log 笔记。真机核验：session_meta.payload.cwd + response_item message content 块。
- **opencode**（已落地，opt-in MANUAL）：`~/.local/share/opencode/storage/session/<hash>/ses_*.json`，字段已自带 `{id,projectID,directory,title,summary,time}`（已是摘要级，无需解析逐条消息），`directory` 末段驱动项目匹配。真机核验通过（v1.1.36 样本）。
- **cursor**（范围内，未实现）：聊天/Agent 历史在未公开的 `state.vscdb`（本机 7.4MB SQLite，序列化了 workspace 键控的 blob），逆向、易碎。实现前须先只读核验 `itemTable`/`cursorDiskKV` 之类表里 chat 记录的真实编码，且**默认关**、明确标注非官方契约。风险最高的一源。
- **openclaw**（范围内，格式已核验，未实现）：记忆 = `~/.openclaw-autoclaw/workspace/{IDENTITY,SOUL,USER,AGENTS,HEARTBEAT,TOOLS}.md`（与 RepoNest 自身 awareness 文件同名同形）。版本是日期式 `lastTouchedVersion: 2026.6.8`，**没有 1.0/2.0 判别位**——产品决定「只做 2.0」= 直接按当前布局实现、不写 1.0 兼容，无需额外识别。**安全红线**：`~/.openclaw-autoclaw` 含**私钥**（`office-plugin-tls/*.pem`）、`vault-roots.json`、`identity/`、`client-sign.json` 等——importer 必须**严格 allowlist 到 `workspace/*.md`**，绝不遍历整目录、绝不把任何 key/token/vault 落进笔记。**待决**：这些是 **agent 全局记忆、非按项目分**（不同于 codex/opencode 每会话带 cwd），落到 (project, source, title) 键上需先定「全局记忆挂到哪个项目」——按 `workspace/.git` 仓库推断单一项目？还是让用户指定？或建一个专用「OpenClaw 记忆」项目？未定前不硬塞（ProjectID 0 会被管线当 skipped 丢掉）。
- **hermes**（范围内，未实现）：**Hermes 是 Nous Research 的独立产品**（bundle `com.nousresearch.hermes`），与 OpenClaw 是**两家**、各自成源——先前把 `.hermes-runtime-receipts`（OpenClaw 目录里的一个子项）误判成「Hermes 是 OpenClaw 运行时」，已更正。**本机无 Hermes 可检视数据目录**（仅一个 WebKit 安装偏好桩），按「先核验真机落盘格式再写 importer」的纪律，**需产品/用户给一份真实数据目录或样例路径**方可实现，不臆测其 schema。
