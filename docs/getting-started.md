---
title: 快速开始
order: 1
---

# 快速开始

RepoNest 是一款本地优先的桌面应用（Wails v2，单文件、零运行时依赖），核心价值是**本地项目的上下文理解与知识沉淀**：自动发现本机 Git 仓库，快速沉淀笔记、依赖、技术栈与活跃信息，方便你和 AI 检索复用。仪表盘与统计是支持能力，不是产品主入口。

## 下载安装

安装脚本会同时装上桌面应用和 `reponest-mcp`（AI 客户端用的 MCP 服务器）。

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/sky-jiangcheng/RepoNest/master/scripts/install.sh | bash
```

macOS 装到 `/Applications/RepoNest.app`，Linux 装到 `/usr/local/bin/reponest`；两者都会把 `reponest-mcp` 放进 `/usr/local/bin`。

### Windows

```powershell
iwr -useb https://raw.githubusercontent.com/sky-jiangcheng/RepoNest/master/scripts/install.ps1 | iex
```

装到 `%LOCALAPPDATA%\RepoNest` 并自动加入用户 PATH。

或从 [GitHub Releases](https://github.com/sky-jiangcheng/RepoNest/releases) 下载对应平台的归档文件（macOS `.dmg` / Linux `.tar.gz` / Windows `.zip`）手动解压。

桌面应用也可用包管理器安装（需先添加 tap，Linux 桌面版只提供 tarball）：

```bash
brew tap sky-jiangcheng/repo && brew install --cask sky-jiangcheng/repo/reponest
```

## 连接 AI 客户端

AI 走的是独立分发的 `reponest-mcp`（MCP stdio 服务器）。它**不需要桌面应用**——直接读同一个本地数据库，两端行为一致。安装脚本会顺带装上；也可以单独装：

| 方式 | 平台 | 命令 |
|------|------|------|
| 手动（无需前置） | 全平台 | 从 [Releases](https://github.com/sky-jiangcheng/RepoNest/releases) 下载 `reponest-mcp-<target>.tar.gz` / `.zip` |
| Homebrew | macOS | `brew tap sky-jiangcheng/repo && brew install --cask sky-jiangcheng/repo/reponest-mcp` |
| Homebrew | Linux | `brew tap sky-jiangcheng/repo && brew install sky-jiangcheng/repo/reponest-mcp` |
| Scoop | Windows | `scoop bucket add repo https://github.com/sky-jiangcheng/scoop-repo && scoop install repo/reponest-mcp` |

注册到 Claude Code：

```bash
claude mcp add reponest -- "$(which reponest-mcp)"
```

Cursor 等客户端在 **Settings → MCP → Add Server** 里填同样的 `command` 路径。

注册后可自检安装是否就绪：

```json
{"tool": "reponest_agent_score"}
```

包管理器清单见 [`packaging/`](../packaging/README.md)，版本号由 `wails.json` 统一派生，发布后需回填 `sha256`。

## 首次启动

启动后直接打开桌面窗口（Wails 应用，无需浏览器）。

### 1. 配置扫描目录

首次启动会自动播种默认扫描根目录：

| 平台 | 默认扫描范围 |
|------|-------------|
| macOS | 当前用户 HOME 目录 |
| Linux | 当前用户 HOME 目录 |
| Windows | 除 C: 盘外的所有磁盘 |

可在 **设置 → 扫描目录** 中修改（支持添加 / 移除，扫描深度 1-2 级）。

### 2. 执行扫描

仪表盘点击 **重新扫描**，应用递归发现扫描根下的所有 Git 仓库并按目录智能分组为项目。

> 首次扫描仅登记仓库与项目，不预扫历史提交数据。

### 3. 收藏仓库

在仪表盘中点击星标收藏关注的仓库。收藏后卡片展示完整统计（今日新增 / 文件 / 仓库数 / 净增 / 团队总量）。

### 4. 回填历史数据

在收藏的仓库卡片上点击 **刷新历史** 按钮，回填该仓库近 365 天的每日统计数据（进度环与热力图随即填充）。

## 核心功能速览

| 功能 | 定位 | 说明 |
|------|------|------|
| 知识库 | 核心 | 跨项目笔记中心：Markdown、标签、置顶、FTS5 全文搜索、版本历史 |
| 仓库知识挖掘 | 核心 | 自动提取 README、技术栈、语言占比、依赖、贡献者、活跃度 |
| 项目理解与检索 | 核心 | 命令面板 `⌘/Ctrl+K`、全局搜索、项目上下文跳转 |
| AI 集成 | 核心 | llms.txt、笔记导出、MCP server（含 agent-score 自检，见[AI 集成](features/ai-integration.md)） |
| 仪表盘 | 支持 | 每日目标进度环、项目卡片、趋势折线图、提交热力图 |
| 插件系统 | 实验 | yaegi 进程内 Go 脚本 + 知识源导入（见[知识源导入](plugins/overview.md)） |
| PWA | 暂缓 | 支持安装到桌面/主屏幕，当前不再继续扩展 |

## 数据与日志位置

| 内容 | macOS | Windows | Linux |
|------|-------|---------|-------|
| 数据库 | `~/Library/Application Support/reponest/dashboard.db` | `%APPDATA%\reponest\dashboard.db` | `~/.config/reponest/dashboard.db` |
| 插件目录 | `…/reponest/plugins/` | `…/reponest\plugins\` | `…/reponest/plugins/` |
| 日志 | `~/Library/Logs/reponest.log` | `%APPDATA%\reponest\logs\reponest.log` | `$XDG_STATE_HOME/reponest/reponest.log`（默认 `~/.local/state/reponest/`） |

升级时 schema 自动迁移，数据无需手工处理。详细排障见[故障排查](troubleshooting.md)。

## 从源码构建

**前置要求**：Go 1.25+、Node.js 20+；开发调试可选 [Wails CLI](https://wails.io) v2.13+。

```bash
# 前端依赖与构建（产物被 go:embed 进二进制）
cd web && npm install && npm run build && cd ..

# 桌面应用
go build -ldflags="-s -w" -o reponest .

# MCP server
go build -o reponest-mcp ./cmd/mcp/
```

开发模式：`wails dev`（前端热更新 + Wails 绑定注入）。测试：`go test ./...`；前端 `npm test` / `npm run build`（tsc 严格检查）。

## 下一步

- [仪表盘](features/dashboard.md)：目标进度、热力图与排序
- [知识库与笔记](features/knowledge.md)：块编辑器与全文搜索
- [设置](features/settings.md)：扫描、标准、外观、插件
