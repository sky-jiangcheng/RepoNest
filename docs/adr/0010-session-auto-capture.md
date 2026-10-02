# ADR-0010: 会话自动捕捉——零人工参与的 Claude Code 会话交接（M1）

- 状态：Proposed（**解析核心已落地** `internal/importers/claude/session.go`，实测对齐真机格式；端到端自动捕捉的默认开关/触发方式仍需产品拍板）
- 日期：2026-10-02
- 关联：[ADR-0007](0007-session-memory-protocol.md)（context/handoff 协议与 `handoff` 标签）、[ADR-0009](0009-ide-presence.md)（SessionEnd hook 一键接入）、[ADR-0006](0006-scope-freeze.md)（范围冻结）、TODO 会话记忆路线 M1

## 背景

今天 RepoNest 的「交接」是**由 agent 主动推送**的：Claude Code 等客户端在会话里调用 `reponest_handoff`（MCP），把结构化的 exit record 交给 `internal/service/handoff.go` 的 `CreateHandoffNote` 落成一条带 `handoff` 标签的笔记，`reponest_context` 再把它排在下次会话的最前（`internal/service/context.go`）。这条链路的前提是 **agent 记得、且愿意去调那个工具**。一旦用户/agent 忘了调，会话就无痕流失。

M1 要补的是「零人工参与」的下限：**不依赖 agent 主动调用**，由 RepoNest 直接读取 Claude Code 自己落盘的原始会话记录 `~/.claude/projects/<project-slug>/<session-uuid>.jsonl`（逐行 JSON 事件流：user/assistant 消息、tool_use、tool_result），从中提取「最后一段 assistant 结论 + 关键工具调用摘要」自动生成一条 handoff。

注意与既有 `internal/importers/claude` 的**边界**：那个 importer 读的是 `~/.claude/projects/*/memory/*.md`（Claude 手动维护的**记忆文档**），走 `plugin.KnowledgeImporter` 通道；M1 读的是 `*.jsonl` **原始逐字会话**，是另一份、更敏感、格式更不稳定的数据源。两者不可混为一谈。

## 决策（提案，均需拍板后方实现）

1. **默认关，显式开**。自动捕捉读取的是完整对话逐字稿，属敏感数据。能力默认关闭，在 设置 → AI 集成 下单独开关；未开启时任何代码路径都不得触碰 `~/.claude/projects/*.jsonl`。
2. **本地只读 + 路径白名单**。仅读 `~/.claude/projects/<slug>/*.jsonl`，不遍历其它目录、不外传、不进任何遥测。落盘 handoff 笔记复用 `CreateHandoffNote`（本地 SQLite），产物与 agent 推送的 handoff 同构，`reponest_context` 无差别消费。
3. **有界提取，控体积与隐私**：只解析尾部 N 条消息（默认建议 N≈最后一段 assistant + 其前一条 user + 其间 tool_use 名称），不整文件入库、不保留逐字全文；生成的是「摘要级」handoff 而非会话存档。原始 jsonl 不回写、不复制。
4. **触发方式二选一（待决）**：(A) 事件驱动——复用 ADR-0009 的 SessionEnd hook，会话结束时调一个只读端点/CLI 触发解析；(B) 按需扫描——设置页「导入最近会话」按钮 + 可选手动扫描。两者都不引入后台轮询常驻。推荐先做 (A)（真正零人工），(B) 作兜底。
5. **格式漂移防护**：Claude Code 的 jsonl schema 非公开契约、随版本变。解析器必须「宽松失败」——字段缺失/结构不认识时跳过并记日志，绝不因单条记录解析失败中断扫描；以 golden-file 测试锁定当前已知的几种事件形状。

## 理由

- 「零人工参与」是产品级卖点（会话记忆环自动闭合），但其代价是直接读取全量对话逐字稿——这是隐私与合规的实质扩张，必须由用户显式知情并选择开启，不能默认生效。
- 有界提取（尾部 + 摘要）同时压低体积与敏感面：handoff 的价值在「下一次会话能接上」，不在「回放整段历史」。
- 复用既有 `handoff` 标签与 `CreateHandoffNote`/`reponest_context` 链路，意味着自动捕捉只是**多一个 handoff 生产者**，不新增消费端、不碰 ADR-0006 冻结面。

## 后果

- 正面：不依赖 agent 自觉也能沉淀会话交接；记忆环对「忘了调工具」的用户自动兜底。
- 负面/风险：绑定 Claude Code 未公开的 jsonl 格式，需持续跟随其变更；引入一个敏感只读路径，隐私声明与文案必须同步。
- **待决（实现前需回答）**：① 默认触发用 (A) hook、(B) 按需，还是都上；② 尾部 N 的取值与是否需要「去工具结果」红白名单；③ 生成的 handoff 是否标注「自动捕捉」来源以与 agent 推送区分；④ 非 Claude 客户端（见 ADR-0011）是否纳入同一机制。
