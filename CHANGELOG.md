# Changelog

本项目遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 格式，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

版本号 SSOT 为 `wails.json` 的 `info.productVersion`，由 `scripts/bump-version.sh` 同步至 `web/package.json`、`internal/version/version.go` 与文档站徽章。

## [Unreleased]

### 新增

- **文档双语化，英文为默认语言**：手册 15 个核心页面全部提供英文版（`docs/en/`，中文源保留在 `docs/`），
  README 拆分为英文主文件 + `README.zh-CN.md`（互链切换）。文档站改为双 locale 构建：
  英文输出到站点根路径，中文在 `/zh/` 子路径，每页侧栏带语言切换器，英文站根页对中文浏览器
  一次性重定向到中文版。8 篇 ADR 详情页补齐英文版（含 ADR 状态规范化为
  Accepted / Superseded 标准用语）；产品评审与定位简报暂保持中文。

## [1.9.5] - 2026-09-30

### 新增

- **会话记忆协议（ADR-0007）**：MCP 工具从 10 个扩展到 12 个，补齐 agent 会话边界的记忆两端：
  - `reponest_context`（会话开始）：一次调用返回项目完整上下文 Markdown——技术栈 / README 摘要 /
    语言占比 / 依赖 / 贡献者 / 活跃度（`repo_meta` 缓存，未缓存时后台异步挖掘）、最近提交、开放待办、
    高相关知识笔记（`handoff` 标签笔记排序置顶）。项目解析支持 `project_id` 精确 → `project_name`
    模糊 → 单项目无参自动解析；多匹配返回项目目录供 agent 二次选择，绝不猜测注入。
    实现于 `internal/service/context.go`（`ResolveProject` / `BuildProjectContext`）。
  - `reponest_handoff`（会话结束）：结构化会话交接协议，`summary` 必填 + `changes` / `decisions` /
    `gotchas` / `next_steps` 至少一项非空，渲染为固定 Markdown 模板落库（`handoff` 标签自动附加），
    下一个会话经 `reponest_context` 自动读到。实现于 `internal/service/handoff.go`。
  - 工具注册拆分至 `cmd/mcp/tools_context.go`；测试覆盖 service 层（项目解析 / 上下文渲染 /
    交接排序 / 校验）与 MCP 工具层（经真实 server 调用），`SKILL.md` 工作流同步升级为
    「session start → context / session end → handoff」协议。

## [1.8.1] - 2026-09-28

### 修复

- **MCP `reponest_projects_list` 空库返回 `null` 而非 `[]`**：`ListProjects()` 在无项目时返回
  nil 切片，`makeJSONResult` 直接序列化成 `null`，与 httpapi 层已有的空集合回归测试
  （`server_test.go`）口径不一致。新注册的项目列表 handler 先归一化为空切片再编码。
  同时捕获了同病灶的检索路径（搜索空结果本就已归一化，此处仅补充断言）。

### 新增

- **MCP 工具测试**（`cmd/mcp/main_test.go`，13 个用例）：此前 566 行、作为 AI 唯一执行入口的
  `cmd/mcp` 零覆盖。测试经 `registerTools()` 构造真实 server、通过 `GetTool()` 调用 handler，
  与客户端同一路径。覆盖：工具注册与 schema、参数校验（空 query 返回提示而非协议错误）、
  笔记创建/读取/更新/搜索全回环（含 FTS5 索引写入后可检索——正是 1.8.0 修复的静默漏搜路径）、
  not-found 文案、`agent_score` 不再重复计分（DB 连通与“无笔记”是两个独立信号）、
  `integrity` 新库无误报、以及**注入 FTS 索引漂移后 integrity 必须报出 FTS 项**。
- **安装脚本冒烟测试**（`.github/workflows/install-smoke.yml`，Linux/macOS/Windows 三平台）：
  `install.sh` / `install.ps1` 此前从 `releases/latest/download` 下载，错误的资产名只会在
  用户侧失败且无从发现（1.8.0 之前三个平台的一键安装均从未成功）。现在脚本的下载根可被
  `RELEASES` / `REPO_NEST_RELEASE_BASE`、安装目录可被 `INSTALL_DIR` / `APP_INSTALL_DIR` 覆盖，
  CI 用本地 fixture 服务器跑真实下载→解压→安装路径，MCP 二进制是 `./cmd/mcp` 的真实构建
  并实际执行 initialize 握手验证。macOS 侧用 `hdiutil` 构造真实 dmg，覆盖挂载/卸载分支。
- **Dependabot 分组合并**（`.github/dependabot.yml`）：三个生态均由 weekly/10 PR 改为
  monthly + `groups` 全量合并，避免依赖升级淹没有真实议题（此前 14 个 open issue 里 13 个是 bump）。

### 变更

- 删除根目录残留的 `package-lock.json`（87 字节，`web/package-lock.json` 才是真锁文件，
  ci.yml 注释中已标注其为 stray）。

## [1.8.0] - 2026-09-28

### 修复

- **FTS5 索引回填是永久空操作（`migrate.go` v7）**：v7 用
  `WHERE id NOT IN (SELECT rowid FROM project_notes_fts)` 回填索引，但
  `project_notes_fts` 是 external-content 表，该子查询由**内容表**回答，条件恒为假。
  凡是 v7 执行时已有笔记的数据库，索引都是空的，且**没有任何机制能修复**。
  症状是静默的：`SearchNotes` 只在 FTS 查询**报错**时降级 LIKE，查询匹配不到时不报错也不返回结果，
  于是搜索一直只返回笔记的一个子集。新增迁移 **v12** 用 FTS5 官方 `'rebuild'` 命令重建两个索引，
  存量库下次启动自动修复。v7 保持原样（迁移不可变），已在代码中标注该缺陷并说明不可照抄。
  回归测试 `internal/db/migrate_fts_repair_test.go` 直接构造"索引空、内容表有数据"的损坏态；
  已验证移除 v12 后该测试必然失败。
- **`scripts/install.sh` 用 HTTP 下载函数拷贝本地文件**：提取出的 MCP 二进制是经
  `curl -fsSL <本地路径>` 拷过去的，真实 curl 会以 `Protocol not supported` 失败，
  结果装到用户机器上的是一个压缩包而不是可执行文件。拆出 `install_binary()` 走 `install -m 0755`。
- **Windows 一键安装从未成功过**：`install.ps1` 下载 `reponest-windows-amd64.exe`，
  而发布流程只产出 `.zip`。改为下载 zip → `Expand-Archive` → 校验 exe 存在。
- **macOS / Linux 一键安装从未成功过**：`install.sh` 下载无扩展名的 `reponest-$TARGET`，
  实际资产是 `.tar.gz`（Linux）和 `.dmg`（macOS）。Linux 改为解压 tar.gz，
  macOS 改为 `hdiutil attach` 后拷贝 `RepoNest.app` 到 `/Applications`。
- **`reponest_agent_score` 两项检查重复计分**：第 1、2 项判据都是 `noteCount > 0`，
  同一信号被数了两遍，抬高分数。第 1 项改为真正检查数据库连通性（`Health()`）。

### 新增

- **CI 测试门禁**（`.github/workflows/ci.yml`）：push 到 master 与所有 PR 触发，
  Go 与 Web 两个独立 job（`go build ./...` + `go test -race`；`npm ci` + `npm test` + `npm run build`）。
  此前 `release.yml` 只在打 tag 时跑、`pages.yml` 只在文档变更时跑，**仓库里 18 个 Go 测试文件
  和 10 个前端测试文件一个都没有进过 CI**。
  门禁同时解决了 `//go:embed all:web/dist` 在干净检出上无法解析的问题（`web/dist` 被 gitignore）。
- **数据可信度审计**（`internal/integrity` + MCP 工具 `reponest_integrity`）：6 项**只读**检查——
  FTS 索引漂移、孤儿行、schema 形状 vs 版本戳、扫描覆盖率、知识缓存新鲜度、版本快照孤儿，
  输出带人类可读证据与 0-100 可信度分数。动机：项目承诺"把知识建模成可被 AI 依赖的结构化数据"，
  此前没有任何机制能回答"这个库还能信吗"。
  与 `reponest_agent_score` 分工明确——后者答"配置好了吗"，前者答"数据还对吗"。
  该工具上线即复现了 v7 的索引漂移 bug。
- **`reponest-mcp` 独立分发**：此前 release.yml **从未构建过 MCP 二进制**，
  README 只能让用户自己 `go build`。现在四个平台的 MCP 产物随 release 一起发布。
- **包管理器支持**（`packaging/`）：Homebrew Cask（macOS 桌面 / macOS MCP）、
  Homebrew Formula（Linux MCP）、Scoop（Windows 桌面 / Windows MCP）。
- **`scripts/build-release-assets.sh`**：本地构建发布产物到 `assets/releases/v<version>/`。
  MCP 是纯 Go 零 CGO，任意宿主都能交叉编译四平台；桌面壳需要各自平台的原生工具链，
  脚本检测到无法构建时会明确说明跳过了什么，而不是半途失败留下一个看起来完整的目录。
- **`scripts/update-manifests.sh`**：把 `wails.json` 的版本号同步到全部包管理器清单
  （已接入 `bump-version.sh`）；`--fill-sha256` 从 `SHA256SUMS` 自动回填校验值。

### 变更

- 笔记导出的 MCP 工具数 9 → 10。
- 安装文档（README / getting-started / SKILL.md）三处统一，并说明各平台实际安装位置。
- `docs/features/ai-integration.md` 补充"就绪度 vs 数据可信度"的分工说明。

### 已知问题

- **历史 release 的资产名仍是旧品牌**：v1.7.9 及之后发的是 `gitbuddy-*`，
  v1.7.6 及之前是 `gitboard-*`。1.7.7 的更名只覆盖了仓库内容，没有传导到发布产物文件名。
  当前 `release.yml` 已统一为 `reponest-*`，从 1.8.0 起一致；历史 release 需要维护者决定是否清理。
- 本版本的 Homebrew / Scoop 清单摘要已用 release 实际产物回填（8/8 匹配）。
  后续版本需在产物发布后跑 `./scripts/update-manifests.sh --fill-sha256`；
  摘要为空的清单 brew / scoop 会拒绝安装——刻意设计，宁可安装失败也不装未校验的二进制。
### 新增

- 手册新增[数据与备份](docs/data-management.md)页：备份、换机迁移、重置与卸载指引
- 新增 `CONTRIBUTING.md` 与 `CODE_OF_CONDUCT.md`
- README 新增目录与「命名分层」规约说明

### 变更

- 项目标识更名为 `repo-nest`：Go module 名、npm 包名、GitHub 仓库名（旧 URL 由 GitHub 自动重定向）；命令名、数据目录与 MCP 工具前缀 `reponest` 保持不变（冻结标识）
- 品牌分层落地（对齐「品牌负责被记住，品类词负责被搜索」原则）：窗口标题与 HTML `<title>` 采用完整展示名 `RepoNest: Local Git Knowledge Base`；`wails.json` comments 同步
- 架构文档新增「命名分层」一节，固化冻结标识（命令名/数据目录/MCP 前缀 `reponest`）不随品牌变化

### 修复

- 架构文档「构建与产物」表移除重复的 `reponest-mcp` 行


## [1.7.9] - 2026-09-03

### 修复

- **项目详情页崩溃**：`GetProjectOverview` 的空切片被 `json.Marshal` 序列化为 `null`，前端读 `recent_commits.length` 抛 TypeError 导致整页白屏；后端在响应出口把 `RecentCommits` / `Dependencies` / `TopContributors` / `TechStack` / `Languages` 归一化为 `[]T{}`，前端补 optional chaining，并加回归测试
- **扫描根目录为空时添加根目录崩溃**：同源问题（`GetConfig` 返回 `scan_roots: null`，且 `db.GetScanRoots` 的错误被 `_` 吞掉）；后端归一化 + 不再吞错，前端 add/remove handler 加 `?? []` 兜底，并加回归测试
- **复制降级**：`wails://` 非安全上下文下 `navigator.clipboard` 不可用，新增 `utils/clipboard.ts`（优先 async Clipboard API，回退 `execCommand('copy')`），详情页与知识库页复用

### 变更

- **详情页设计成熟度三波改进**：右上操作区重排（Copy AI Context 升为主操作、Quick Note 移入侧栏、项目级别 ± 下沉到 meta 区）；头部新增摘要行（主语言 / 最近提交 / 子仓库数）；仓库列表统计标签由「日期: ±N」改为「作者: ±N」，日期移入 hover；趋势图与热力图共用 `ScopeToggle`（周 / 月 / 全部），无活动自动塌缩为空态；侧栏取消 sticky
- 热力图统计计算优化：`today` 提出循环，日期改用 `YYYY-MM-DD` 字符串区间比较，避免逐日 `new Date()` 解析
- 清理 code review 遗留项：删除死代码（`dateParam === 'newNote'` 分支、`.heatmap-title` 死规则、`scope-toggle` 空 class）、消除 `.map(t => ...)` 对 i18n `t` 的遮蔽
- 新增 14 条中英 i18n 键（scope / mainLanguage / lastCommit / groupLevel / noDataInRange / rangeWeek|Month|All 等），双语对齐

### 文档

- 用真实应用截图替换自动生成的仪表盘 / 知识库配图，新增设置页截图
- 新增「存储结构优化与 AI 价值」并补全 AI 集成定位

## [1.7.8] - 2026-09-01

### 修复

- 搜索片段截取越界 panic；schema 版本解析静默吞错
- 收窄笔记版本快照触发条件（不再为无意义变更建快照）；修复首次扫描的新项目没有历史数据

### 变更

- `knowledge` chunk 从 1.83MB 降至 471kB，构建告警清零：`highlight.js` 改用 `lib/common`（37 语言而非全量 190），`mermaid` 改动态 `import()` 按需加载
- 清理 ESLint 10 + react-hooks v7 报出的 7 处问题
- 删除未注册的 MCP 工具文件并统一 gofmt 格式
- 同步 `package-lock.json`，移除 `vite-plugin-pwa` 及其传递依赖
- 升级 brace-expansion 5.0.7 → 5.0.9，修复 high 级 DoS 漏洞

## [1.7.7] - 2026-09-01

### 变更

- 产品正式更名为 RepoNest（旧名 GitBoard）：模块与包路径、文档、可执行文件与文档站徽章一并同步

## [1.7.6] - 2026-09-01

### 修复

- 修复 Wails 绑定命名空间取错导致的桌面端 UI 失效（命名空间由 Go 包名决定：`package app` → `go.app.App`，而非 `go.main.App`）
- code review 的 P0 / P1 / P2 问题全部修复（sprint 6-8 收尾）
- 恢复被误删的 DMG 资源与 MCP `main.go`

## [1.7.5] - 2026-08-31

### 变更

- UI 改版：蓝色主色调，对比度与视觉层级优化

## [1.7.4] - 2026-08-27

### 新增

- 同步远程待办事项
- MCP 笔记工具；`docs/features/ai-integration.md` 的 MCP 工具表补齐为 9 个（含 2 个写操作）

### 变更

- **定位治理（PR1）**：统一对外定位为「本地优先的代码项目上下文库」，核心闭环为「发现本地项目 → 理解项目 → 沉淀知识 → 检索知识 → 交给 AI 使用」；仪表盘与统计降级为支持能力（前端默认页已是知识库，导航顺序 知识库 → 仪表盘 → 设置）。README 截图与功能特性表按核心能力优先重排（知识库 / 项目详情 / AI 接口在前，仪表盘在后）；`docs/positioning-brief.md` 新增「统一对外口径（权威短文案）」节，写入中英文三句定位与 release note。详见 [ADR-0006](docs/adr/0006-scope-freeze.md)。
- **错误一致性与可用性（PR2）**：新增 `web/src/components/ErrorBanner.tsx` 作为页面级错误的统一渲染路径（消息 + 重试按钮 + i18n），替换 Dashboard / ProjectDetail / Knowledge / Settings 四页各自内联的 `error-banner` JSX 与重复 catch 样板。ProjectDetail 与 Settings 的重试改为复用与初次加载相同的加载函数，修复旧实现只 setProject / 漏 reset loading 的半状态问题。`NoteSection` 的 create/save/move/delete/pin/restore/diff 等原本 `/* ignore */` 的静默 catch 改走统一的 `run(op, errMsg)` 包装：失败时 setError 并记录最近失败操作供 ErrorBanner 的重试按钮重放；乐观 pin 失败回滚原状态。
- **TODO 收尾（中优先级）**：补齐 `useConfirmClick` 测试；将 `useApiData`（TTL 缓存 + 请求去重）接入 NoteSection 与 CommandPalette 的「全部项目」下拉列表，共享缓存键 `projects:all`，跨组件只发一次请求；Dashboard 的 projects 拉取现已迁移到 `useApiData`（独立键 `dashProjects`，按 date/starredOnly 作用域，因卡片依赖按日统计），star 切换 `invalidateCache('projects:all')` 使三处组件 starred 状态一致，保留乐观 star 覆盖层避免骨架闪烁；移除 `wails.json` 中从未使用的 `wailsjsdir`；校验 `examples/plugins` 两个示例插件（宿主 SPI 未变，`go build ./...` 通过，预期兼容）；`openapi.json` 契约说明与 `build-docs.mjs` 对 `marked` 的依赖经核实已满足，无额外改动。
- **桌面端路由（HashRouter）**：Wails WebView 在自定义源下 BrowserRouter 的 history/location 变更会抛 DOMException，故桌面壳改用 `HashRouter`、PWA/浏览器仍用 `BrowserRouter`；`spaFallback` 注释同步说明该约定。
- sprint 6-8 代码重构与清理：NoteSection hooks 化、大文件拆分、CSS 死代码清理、懒加载、recover 防护

### 维护

- 依赖升级：wails 2.13.0 → 2.14.0、mcp-go 0.57.0 → 0.58.0、highlight.js 11.11.1 → 11.12.0、katex 0.18.3 → 0.18.4、actions/setup-node 5 → 7
- 从仓库移除 `.omo` 运行态产物

## [1.7.3] - 2026-08-20

### 新增

- 项目上下文主页：快速笔记入口、Copy AI Context、概览空状态引导

### 变更

- 范围冻结 ADR-0006、产品闭环叙事与 CLI/MCP 收敛（#73 #74 #76）

### 修复

- agent-score 的幽灵 CLI 检查替换为 MCP 二进制 + llms.txt 导出检查（#76）；修复 `cliPath` 未定义导致的编译错误

### 测试

- 补齐搜索闭环覆盖：`useApiData` / `useDebouncedCallback` / transport / endpoints（#75）

### 维护

- 加固运行时检查与质量门禁

## [1.7.2] - 2026-08-18

深度代码审查（`docs/code-review/2026-08-18-deep-review.md`）缺陷修复：

### 修复

- **🔴 并发数据库锁**：`InitDB` 限制连接池为单连接（`SetMaxOpenConns(1)` + `SetConnMaxLifetime(0)`）并设置 `PRAGMA busy_timeout=5000`，消除并发 Wails/扫描/插件访问导致的 `database is locked`
- **🔴 知识缓存静默失效（存量库）**：新增 v10 幂等迁移，为早期版本创建的 `repo_meta` 补齐 `dependencies` / `top_contributors` / `activity` 三列（`createTables` 新建表已含，存量库需此修复才能命中缓存）
- **🟠 知识源状态误报**：插件导入 `TriggerImport` 的 `lastErr` 改为记录逐文档 upsert 真实错误，知识源 `Enabled` 不再恒为 true
- **🟠 大仓库挂起**：`DetectContributors` 用 30s 上下文超时包裹 `git shortlog`
- **🟠 首屏阻塞**：`GetProjects` / `GetProjectStats` 的按需 git 统计刷新移至后台 goroutine，仪表盘/概览首开不再卡顿
- **🟠 `git_author` 配置生效**：运行时可设置个人作者，覆盖自动检测的 `git user.name`（"我的"统计/热力图/最近提交随之更新）
- **🟡 健壮性**：`mineAndCache` 记录 `UpsertRepoMeta` 错误而非吞掉；`Mine` 返回非 nil 切片避免 JSON `null`；按语言行数统计 scanner 缓冲放大到 16MB（兼容 minified 文件）；`daily_stats` 新增真实提交数 `commits` 列（此前热力图误用 `COUNT(DISTINCT author)`）

### 维护

- **🟡 `InferRepoMeta` 无超时**：派生仓库展示名时读取 `git config user.name` 改用 30s 上下文超时包裹，避免挂掉的 working tree 阻塞扫描/发现路径
- **🟡 `refreshProjectStatsForDate` 缺失非零守卫**：与 `refreshRepoStatsRange` 对齐，git 出错返回的全 0 `Result` 不再写入每日统计行（原会令仪表盘显示「0」而非「无数据」，掩盖错误）
- **版本号对齐**：`internal/version/version.go` 经 `scripts/bump-version.sh` 同步至 `1.7.2`（`wails.json` / `web/package.json` 一并更新），消除应用内报告版本与 tag 长期漂移

## [1.7.1] - 2026-08-18

### 修复

- 修复 ESLint 被 TypeScript 7.0 兼容性阻塞问题：降级 TypeScript 至 6.0.3，修复 react-hooks/refs 违规（4 个 hook），修复 set-state-in-effect 违规（7 个文件），修复 markdown.ts 不必要转义和 seo.ts 缺失依赖

## [1.7.0] - 2026-08-17

深度重构版本：后端服务化、前端组件化，行为保持不变（除下述明示的修复与契约变更）。决策记录见 [ADR-0005](docs/adr/0005-service-layer.md)。

### 新增

- **internal/service 业务层**：Wails 桌面、CLI、MCP 三端共享同一实现，消除三处重复的查询/格式化逻辑
- **internal/app 薄绑定层**：根目录 14 个 handler 文件（约 1900 行）收敛为每方法 1-3 行委托；`package main` 只剩 `main.go`
- **internal/domain / internal/diff / internal/version**：跨层行类型独立、笔记行级 diff 独立成包、四处硬编码版本号统一为单一常量
- **db 层按域拆分**：1195 行 `queries.go` 拆为 projects / notes / note_versions / todos / repositories / daily_stats / repo_meta / config / scan_roots / search / cleanup；项目升降级 SQL 事务化为受测的 `SplitProjectDown` / `MergeProjectUp`
- **测试补齐**：knowledge 解析（含 go.mod 块状 require）、scanner、diff、项目拆分/合并事务、热力图项目过滤、service 层（fake git provider）；db 测试改用真实 `InitDB` schema（消除手抄 DDL 漂移）
- **前端 API 层拆分**：627 行 client.ts → types / transport / endpoints，统一 `call()` 路由
- **前端 hooks 层**：`useApiData`（TTL 缓存 + 请求去重，待接入页面）、`useDebouncedCallback`、`useScanPolling`、`useConfirmClick`
- **组件拆分**：Dashboard 搜索下拉、Settings 六个 tab、NoteSection 统一 NoteEditor + 版本历史面板、ProjectDetail 概览面板、Knowledge 卡片
- 日志路径按平台（Linux `$XDG_STATE_HOME`、Windows `%APPDATA%\reponest\logs`），修复非 macOS 平台写入 `~/Library/Logs`
- MCP server 进程内单次开库（此前每次工具调用都执行全套迁移）
- vitest + ESLint 工具链（ESLint 受 TS7 兼容性阻塞，见 [TODO](TODO.md)）；tsconfig 恢复 `noUnusedLocals/Parameters`
- 仓库卫生：移除误提交的 20MB 二进制与 AI 工具产物目录，遗留脚本归档至 `scripts/legacy/`

### 修复

- **存量 bug：repository 查询引用不存在的列**（`display_name` / `git_user` / `organization` 从未建列），导致项目仓库列表、扫描后统计刷新、状态栏最近提交、llms.txt 仓库目录在生产环境**全部静默失败**；已核对真实用户数据库确认并修复
- **存量 bug：go.mod 块状 `require (...)` 解析越界 panic**，可致项目概览后台挖掘崩溃；解析器重写并支持单行 + 块状两种形式
- 前端 Rules-of-Hooks 违规（普通函数内调用 `useTranslation`）
- toast 双重定时器互相重置；`EventsOn` 监听器随语言切换累积泄漏（现真实退订）
- 9 处裸 `<a href>` 全页刷新破坏 SPA；ProjectCard 中 `<button>` 嵌套 `<a>` 的非法 HTML/a11y 问题
- **项目详情页热力图显示全局数据**：`GetHeatmapData` 新增 `projectId` 参数（契约变更，前端已适配）
- 笔记两击确认删除、防抖搜索、扫描轮询等三处重复实现合并

### 变更

- 移除死代码约 1100 行：旧扫描管线（`ScanForRepositories` 等）、未使用的 `ExportProjectStats` / `ExportHeatmapCSV` / `GetNoteVersion` / 分支查询 / storage 垫片等（绑定面变更已同步至 API 参考）
- 约 300 行硬编码中文提取至 zh-CN / en locale 文件
- `GetStatusBar` 改为双检锁（不持锁执行 git 命令）；`ToggleProjectStar` 原子化（TOCTOU）；`ReorderTodos` 单事务（自 1.6.x 移植）

## [1.6.3] - 2026-08-11

### 修复

- 代码质量专项：CSV 注入防护（csvSafe）、TOCTOU 与事务化修复（ToggleProjectStar / ReorderTodos）、状态栏锁优化、`wail()` 空守卫、`ensurePath` 去重、跨平台日志目录（`getLogDir`）

## [1.6.2] - 2026-08-11

### 变更

- 第二轮代码质量清理：错误包裹（`%w`）、helpers 归并、`refreshStatsForRepo` 抽取

## [1.6.1] - 2026-08-10

### 新增

- 产品正式更名为 RepoNest（旧名 GitBoard，当时仅 module/包路径级引用待跟进）
- 记录产品定位决策 ADR 0002，并标记 RFC 0001（插件平台）为 Superseded
- 社区健康文件（issue #25）：CHANGELOG / CONTRIBUTING / SECURITY / CODE_OF_CONDUCT / SUPPORT / Issue+PR 模板 / Dependabot
- 进程内插件系统：yaegi 脚本运行时（目录扫描 / 加载 / 事件总线 / panic 隔离），插件接口见 `internal/core/plugin`
- Claude 记忆导入重构为内置 KnowledgeImporter 插件，与脚本插件共享运行时导入/去重/统计路径
- 知识源导入触发：启动自动导入（可开关）+ 设置页手动触发 + 前端 toast 结果通知
- 知识库升级为首页，支持快速创建笔记（issue #31）
- 知识库体验增强（issue #37）：首屏搜索框自动聚焦、顶部「最近编辑」快速访问区、空状态引导创建或导入 AI 记忆、编辑器「关联项目」下拉快速迁移笔记（新增 `MoveNote` API）
- 设计系统重构（issue #9）：拆分单体 global.css 为设计系统 token + 组件样式，LCH 自适应色板
- Markdown 渲染增强（issue #10）：highlight.js 语法高亮、Mermaid 图、GFM callout、KaTeX 数学公式、任务列表样式
- 全局无障碍基线（issue #12）：skip-link + focus-visible + reduced-motion
- 命令面板无障碍（issue #11）：focus trap + ARIA dialog/listbox
- BrowserRouter 路由升级（issue #13）：可分享 URL，Pages `_redirects` fallback
- PWA 规范修复（issue #14）：theme-color 一致 + 离线 fallback + 安全加固
- SEO 规范（issue #21）：OG / Twitter Card / canonical / sitemap
- AI-ready 内容分发层（issue #15）：llms.txt + .md 路由 + 问答入口
- FTS5 全文搜索升级（issue #18）：FTS5 + 中文分词 + 相关性排序 + snippet 高亮
- 笔记版本历史 + Diff view（issue #16）：复用本地 Git，超越 GitBook CRUD
- 仓库知识挖掘加深（issue #20）：LOC / 依赖图 / 活跃度 / 贡献者
- 安全规范（issue #24）：CSP / HSTS / 安全策略声明 / 依赖扫描
- i18n 框架（issue #23）：字符串外提 + hreflang + 语言切换（zh-CN / en）
- 用户文档站（issue #26）：使用手册 / 教程 / FAQ（docs/ 落地页）
- API 参考文档（issue #27）：OpenAPI 渲染 + 端点说明
- OpenAPI spec + REST 版本化（issue #22）+ OpenAPI 自动渲染与 Try-it（issue #17）
- CLI + MCP + agent-score（issue #28）：`reponest-mcp`、`tools/agent-score` 自检工具

## [1.5.7] - 2026-08-10

### 新增

- 块编辑器（issue #19）：Markdown 双轨编辑，输入 `/` 呼起 block 面板插入 callout/tabs/details/代码/Mermaid/公式等结构化块，块级拖拽排序与增删，编辑产物仍为可读 Markdown
- 桌面端深链路由兜底：Wails AssetServer 对未命中的 GET 请求回退 `index.html`

## [1.5.6] - 2026-08-10

### 修复

- 扫描稳定性与收藏同步修复

## [1.5.5] - 2026-08-10

### 新增

- 启动时历史回填、收藏同步及默认扫描根目录播种

## [1.5.3] - 2026-07-25

### 修复

- 全量扫描策略优化为合并同步，增强跨平台支持

## [1.5.2] - 2026-07-25

### 新增

- 全面更新应用图标系统

## [1.5.1] - 2026-08-01

### 新增

- 优化扫描与性能，支持自定义扫描路径

## [1.5.0] - 2026-08-01

### 新增

- 统一 UI 配色为灰阶并优化仪表盘布局
- 增强数据序列化与前端健壮性

## [1.4.0] - 2026-07-01

### 新增

- 知识库：跨项目笔记中心（Markdown、标签、置顶、全文搜索）
- 仓库知识挖掘：README / 技术栈 / 语言占比
- Claude 记忆导入
- 命令面板（⌘/Ctrl+K）
- PWA 可安装

## [1.3.0] - 2026-06-01

### 新增

- 智能项目分组（Monorepo 识别、级别调整）
- 仓库收藏与按需刷新历史
- 项目详情页：趋势折线图、提交热力图

## [1.2.0] - 2026-05-01

### 新增

- 仪表盘：每日目标进度环、项目卡片、工作日检查

## [1.1.0] - 2026-04-01

### 新增

- 自动发现本地 Git 仓库与基础提交统计
- 模糊搜索

## [1.0.0] - 2026-03-01

### 新增

- 首个正式版本：Wails 桌面应用骨架、GitHub Actions 多平台构建发布

[Unreleased]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.9.5...HEAD
[1.9.5]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.9.4...v1.9.5
[1.7.9]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.8...v1.7.9
[1.7.8]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.7...v1.7.8
[1.7.7]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.6...v1.7.7
[1.7.6]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.5...v1.7.6
[1.7.5]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.4...v1.7.5
[1.7.4]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.3...v1.7.4
[1.7.3]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.2...v1.7.3
[1.7.2]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.6.3...v1.7.0
[1.6.3]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.6.2...v1.6.3
[1.6.2]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.6.1...v1.6.2
[1.6.1]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.6.1
[1.5.7]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.5.5...v1.5.7
[1.5.6]: https://github.com/sky-jiangcheng/repo-nest/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.5.5
[1.5.3]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.5.3
[1.5.2]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.5.2
[1.5.1]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.5.1
[1.5.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.5.0
[1.4.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.4.0
[1.3.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.3.0
[1.2.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.2.0
[1.1.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.1.0
[1.0.0]: https://github.com/sky-jiangcheng/repo-nest/releases/tag/v1.0.0

