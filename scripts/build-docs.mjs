#!/usr/bin/env node
// scripts/build-docs.mjs — 文档站生成器。
//
// 以 docs/**/*.md 为唯一内容源，按 docs/sidebar.json 的导航结构渲染出
// 同目录的 .html（套用原手写文档站的模板样式）。生成物不入库
//（.gitignore 忽略 docs/**/*.html），GitHub Pages 部署时由
// .github/workflows/pages.yml 先执行本脚本。
//
// 依赖 marked（复用 web/node_modules），运行方式：
//   node scripts/build-docs.mjs

import { readFileSync, writeFileSync, existsSync, readdirSync, statSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const docsDir = join(root, 'docs')

// Resolve marked from web/node_modules (the project keeps a single dep tree).
const requireFromWeb = createRequire(join(root, 'web', 'noop.js'))
const { marked } = requireFromWeb('marked')

const sidebar = JSON.parse(readFileSync(join(docsDir, 'sidebar.json'), 'utf8'))
const version = JSON.parse(readFileSync(join(root, 'web', 'package.json'), 'utf8')).version
const REPO = 'https://github.com/sky-jiangcheng/reponest'
const PAGES = 'https://sky-jiangcheng.github.io/reponest/'

// --- Markdown helpers ---------------------------------------------------------

function parseDoc(mdPath) {
  const raw = readFileSync(mdPath, 'utf8')
  let title = ''
  let body = raw
  const fm = raw.match(/^---\n([\s\S]*?)\n---\n/)
  if (fm) {
    body = raw.slice(fm[0].length)
    const t = fm[1].match(/^title:\s*(.+)$/m)
    if (t) title = t[1].trim().replace(/^["']|["']$/g, '')
  }
  if (!title) {
    const h1 = body.match(/^#\s+(.+)$/m)
    if (h1) title = h1[1].trim()
  }
  return { title, body }
}

function renderMarkdown(mdPath) {
  const { title, body } = parseDoc(mdPath)
  let html = marked.parse(body, { gfm: true })
  // Rewrite relative .md links: links into docs/ become same-position .html
  // pages (the output file sits next to the source, so the relative path is
  // unchanged); links to files outside docs/ (README, packaging, TODO...) are
  // rewritten to GitHub blob URLs, because no .html is ever generated for
  // them and the raw .md path would 404 on the Pages site.
  html = html.replace(/href="([^"#]*?)\.md(#[^"]*)?"/g, (m, path, hash) => {
    if (path === '' || path.startsWith('http')) return m
    const target = resolve(dirname(mdPath), path + '.md')
    const relDocs = relative(docsDir, target)
    if (!relDocs.startsWith('..') && existsSync(target)) {
      return `href="${path}.html${hash ?? ''}"`
    }
    const relRepo = relative(root, target).split('\\').join('/')
    return `href="${REPO}/blob/master/${relRepo}${hash ?? ''}"`
  })
  return { title, html }
}

// --- Template -------------------------------------------------------------------

// Sidebar links must resolve from every generated page, and pages live at
// different depths (docs/x.html vs docs/features/x.html). Each page therefore
// gets a relPrefix ("", "../", "../..") that nav links are prefixed with.
const navHtml = (prefix) => sidebar.sections.map(sec => `
      <h3>${sec.title}</h3>
${sec.items.map(it => `      <a href="${prefix}${it.file}.html" data-page="${it.file}">${it.label}</a>`).join('\n')}
`).join('')

function page(title, activeFile, contentHtml, prefix = '', extraHead = '') {
  const active = activeFile
    ? `document.querySelectorAll('.sidebar a[data-page]').forEach(a => { if (a.dataset.page === ${JSON.stringify(activeFile)}) a.classList.add('active') })`
    : ''
  const nav = navHtml(prefix)
  return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${title ? title + ' · ' : ''}RepoNest 文档</title>
${extraHead}  <style>
    :root { --bg: #f8f9fa; --text: #1a1a2e; --muted: #6c757d; --accent: #4caf50; --border: #e2e8f0; --sidebar-w: 240px; }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: var(--bg); color: var(--text); display: flex; min-height: 100vh; }
    a { color: var(--accent); text-decoration: none; }
    a:hover { text-decoration: underline; }
    .sidebar { width: var(--sidebar-w); background: #1a1a2e; color: #fff; padding: 24px 0; flex-shrink: 0; overflow-y: auto; position: fixed; height: 100vh; }
    .sidebar-brand { padding: 0 20px 20px; border-bottom: 1px solid rgba(255,255,255,0.1); margin-bottom: 12px; }
    .sidebar-brand h1 { font-size: 18px; font-weight: 700; }
    .sidebar-brand span { color: var(--accent); }
    .sidebar-brand small { display: block; font-size: 11px; color: rgba(255,255,255,0.5); margin-top: 4px; }
    .sidebar nav { padding: 8px 0; }
    .sidebar h3 { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: rgba(255,255,255,0.4); padding: 12px 20px 6px; }
    .sidebar a { display: block; padding: 6px 20px; color: rgba(255,255,255,0.75); font-size: 13px; }
    .sidebar a:hover { background: rgba(255,255,255,0.08); color: #fff; text-decoration: none; }
    .sidebar a.active { color: var(--accent); font-weight: 600; }
    .main { margin-left: var(--sidebar-w); flex: 1; padding: 40px 48px; max-width: 860px; }
    .main h1 { font-size: 28px; margin-bottom: 8px; }
    .main .subtitle { color: var(--muted); margin-bottom: 32px; font-size: 15px; }
    .main h2 { font-size: 20px; margin-top: 36px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--border); }
    .main h3 { font-size: 16px; margin-top: 24px; margin-bottom: 8px; }
    .main p { line-height: 1.75; margin-bottom: 16px; font-size: 14px; color: #334155; }
    .main ul, .main ol { margin-bottom: 16px; padding-left: 24px; font-size: 14px; color: #334155; }
    .main li { margin-bottom: 6px; }
    .main code { background: #eef2f7; padding: 1px 5px; border-radius: 3px; font-size: 13px; font-family: 'SF Mono', Menlo, monospace; }
    .main pre { background: #1a1a2e; color: #e2e8f0; padding: 16px; border-radius: 8px; overflow-x: auto; margin-bottom: 16px; font-size: 13px; line-height: 1.6; }
    .main pre code { background: none; color: inherit; padding: 0; }
    .main table { width: 100%; border-collapse: collapse; margin-bottom: 16px; font-size: 13px; }
    .main th, .main td { border: 1px solid var(--border); padding: 8px 12px; text-align: left; }
    .main th { background: #f1f5f9; font-weight: 600; }
    .main blockquote { background: #eff6ff; border-left: 3px solid #3b82f6; padding: 12px 16px; border-radius: 0 6px 6px 0; margin-bottom: 16px; font-size: 13px; color: #334155; }
    .main blockquote.warning { background: #fef3c7; border-left-color: #f59e0b; }
    .main .nav-links { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 32px; }
    .main .nav-links a { background: var(--bg); border: 1px solid var(--border); padding: 6px 14px; border-radius: 6px; font-size: 13px; color: var(--text); }
    .main .nav-links a:hover { border-color: var(--accent); color: var(--accent); text-decoration: none; }
    .main hr { border: none; border-top: 1px solid var(--border); margin: 32px 0; }
    .main .back-to-top { display: inline-block; font-size: 13px; color: var(--muted); margin-top: 32px; }
    @media (max-width: 768px) { .sidebar { display: none; } .main { margin-left: 0; padding: 24px; } }
  </style>
</head>
<body>
  <aside class="sidebar">
    <div class="sidebar-brand">
      <h1>Repo<span>Nest</span></h1>
      <small>用户文档 · v${version}</small>
    </div>
    <nav>${nav}
    </nav>
  </aside>
  <main class="main">
${contentHtml}
    <hr>
    <a href="${REPO}" class="back-to-top">← 返回 GitHub 仓库</a>
  </main>
  <script>${active}</script>
</body>
</html>
`
}

// --- Build -----------------------------------------------------------------------

// Every .md under docs/ gets a page, sidebar entry or not: ADR detail pages,
// for instance, are linked from several docs but intentionally kept out of
// the navigation. Skipping them would ship dead links by construction.
function collectMdFiles(dir) {
  const out = []
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    if (statSync(p).isDirectory()) out.push(...collectMdFiles(p))
    else if (name.endsWith('.md')) out.push(p)
  }
  return out
}

const built = []
const sidebarFiles = new Set(sidebar.sections.flatMap(s => s.items.map(it => it.file)))
const missingSidebar = [...sidebarFiles].filter(f => !existsSync(join(docsDir, f + '.md')))
if (missingSidebar.length > 0) {
  for (const f of missingSidebar) console.error(`✗ sidebar 条目缺失源文件: docs/${f}.md`)
  process.exitCode = 1
}

for (const mdPath of collectMdFiles(docsDir)) {
  const base = relative(docsDir, mdPath).replace(/\.md$/, '').split('\\').join('/')
  const depth = base.split('/').length - 1
  const prefix = '../'.repeat(depth)
  const { title, html } = renderMarkdown(mdPath)
  const outPath = join(docsDir, base + '.html')
  if (base === 'index') {
    // Landing page: inject quick nav links into the <!--NAV_LINKS--> slot.
    const quick = sidebar.sections.flatMap(s => s.items).slice(0, 7)
      .map(it => `      <a href="${prefix}${it.file}.html">${it.label}</a>`).join('\n')
    writeFileSync(outPath, page(title || 'RepoNest 文档', '', html.replace('<!--NAV_LINKS-->', quick), prefix))
  } else {
    writeFileSync(outPath, page(title, base, html, prefix))
  }
  built.push(relative(root, outPath))
}

console.log(`✓ 生成 ${built.length} 个页面（v${version}）：\n  ${built.join('\n  ')}`)
