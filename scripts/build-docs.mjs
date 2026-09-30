#!/usr/bin/env node
// scripts/build-docs.mjs — 文档站生成器（双语）。
//
// 内容源分两个 locale：
//   en  docs/en/**/*.md   → 输出到 docs/ 根路径（默认语言，/repo-nest/ 即英文）
//   zh  docs/**/*.md（不含 docs/en/）→ 输出到 docs/zh/ 子路径
//
// 两个树的相对结构互为镜像；英文页链接到尚未翻译的页面（如 ADR 详情）时
// 自动回退到中文版输出路径。侧栏标签取 docs/sidebar.json 的 label/label_en。
// 生成物不入库（.gitignore 忽略 docs/**/*.html），GitHub Pages 部署时由
// .github/workflows/pages.yml 先执行本脚本。
//
// 依赖 marked（复用 web/node_modules），运行方式：
//   node scripts/build-docs.mjs

import { readFileSync, writeFileSync, existsSync, readdirSync, statSync, rmSync, mkdirSync } from 'node:fs'
import { dirname, join, relative, resolve, basename } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const docsDir = join(root, 'docs')

// Resolve marked from web/node_modules (the project keeps a single dep tree).
const requireFromWeb = createRequire(join(root, 'web', 'noop.js'))
const { marked } = requireFromWeb('marked')

const sidebar = JSON.parse(readFileSync(join(docsDir, 'sidebar.json'), 'utf8'))
const version = JSON.parse(readFileSync(join(root, 'web', 'package.json'), 'utf8')).version
const REPO = 'https://github.com/sky-jiangcheng/repo-nest'

// --- Locale registry -------------------------------------------------------------
// srcRoot: where the locale's .md sources live; outRoot: where its .html goes.
// English is the default language: en pages land at the site root, zh under /zh/.

const locales = {
  en: {
    srcRoot: join(docsDir, 'en'),
    outRoot: docsDir,
    htmlLang: 'en',
    titleSuffix: 'RepoNest Docs',
    brandSmall: `Docs · v${version}`,
    backLabel: '← GitHub repository',
    switcher: { self: 'English', other: '中文' },
    landingRedirect: true, // root landing only: send zh browsers to /zh/ once per session
  },
  zh: {
    srcRoot: docsDir,
    outRoot: join(docsDir, 'zh'),
    htmlLang: 'zh-CN',
    titleSuffix: 'RepoNest 文档',
    brandSmall: `用户文档 · v${version}`,
    backLabel: '← 返回 GitHub 仓库',
    switcher: { self: '中文', other: 'English' },
    landingRedirect: false,
  },
}

const isEnSource = (rel) => rel === 'en' || rel.startsWith('en/')

// Source file for `base` in a locale; null when that locale has no translation.
const sourceOf = (locale, base) => {
  const p = join(locales[locale].srcRoot, base + '.md')
  return existsSync(p) ? p : null
}

// Output file for `base` in a locale (independent of source existence).
const outputFileOf = (locale, base) => join(locales[locale].outRoot, base + '.html')

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

// Rewrite relative .md links for one page. Links stay inside the page's locale
// when a same-locale source exists; otherwise they fall back to the other
// locale's output (e.g. an English page linking an untranslated ADR detail
// lands on the Chinese page instead of 404ing). Links leaving docs/ (README,
// packaging, TODO...) become GitHub blob URLs — no .html is generated for them.
function rewriteLinks(html, mdPath, locale, pageOutDir) {
  const other = locale === 'en' ? 'zh' : 'en'
  return html.replace(/href="([^"#]*?)\.(md|go)(#[^"]*)?"/g, (m, path, ext, hash) => {
    if (path === '' || path.startsWith('http')) return m
    const target = resolve(dirname(mdPath), path + '.' + ext)
    const relDocs = relative(docsDir, target)
    if (relDocs.startsWith('..') || ext === 'go') {
      // Outside docs/: point at the repo blob. Chinese pages link the
      // Chinese README companion.
      let relRepo = relative(root, target).split('\\').join('/')
      if (locale === 'zh' && relRepo === 'README.md') relRepo = 'README.zh-CN.md'
      return `href="${REPO}/blob/master/${relRepo}${hash ?? ''}"`
    }
    const base = relDocs.replace(/^en\//, '').replace(/\.md$/, '')
    // Prefer the page's own locale; fall back to the other one when this page
    // links something not yet translated; if neither has it, link the source
    // on GitHub instead of shipping a guaranteed 404.
    const targetLocale = sourceOf(locale, base) ? locale : sourceOf(other, base) ? other : null
    if (!targetLocale) {
      console.warn(`⚠ 未解析的 .md 链接（${locale} ${mdPath}）: ${path}.md`)
      const relRepo = relative(root, target).split('\\').join('/')
      return `href="${REPO}/blob/master/${relRepo}${hash ?? ''}"`
    }
    return `href="${relative(pageOutDir, outputFileOf(targetLocale, base)).split('\\').join('/')}${hash ?? ''}"`
  })
}

function renderMarkdown(mdPath, locale, pageOutDir) {
  const { title, body } = parseDoc(mdPath)
  let html = marked.parse(body, { gfm: true })
  html = rewriteLinks(html, mdPath, locale, pageOutDir)
  return { title, html }
}

// --- Nav & switcher --------------------------------------------------------------

const labelFor = (locale, obj) => (locale === 'en' ? (obj.label_en ?? obj.label) : obj.label)
const titleFor = (locale, sec) => (locale === 'en' ? (sec.title_en ?? sec.title) : sec.title)

// href from a page's output dir to `base` rendered in `locale`'s tree,
// falling back to the other locale when this one has no translation.
function hrefFor(locale, base, pageOutDir) {
  const target = sourceOf(locale, base) ? locale : (locale === 'en' ? 'zh' : 'en')
  return relative(pageOutDir, outputFileOf(target, base)).split('\\').join('/')
}

const navHtml = (locale, pageOutDir) => sidebar.sections.map(sec => `
      <h3>${titleFor(locale, sec)}</h3>
${sec.items.map(it => `      <a href="${hrefFor(locale, it.file, pageOutDir)}" data-page="${it.file}">${labelFor(locale, it)}</a>`).join('\n')}
`).join('')

// English | 中文 — links to the same page in the other locale when it exists.
function switcherHtml(locale, base, pageOutDir) {
  const loc = locales[locale]
  const other = locale === 'en' ? 'zh' : 'en'
  if (!sourceOf(other, base)) {
    return `<span class="lang-switch"><span class="lang-current">${loc.switcher.self}</span></span>`
  }
  return `<span class="lang-switch"><span class="lang-current">${loc.switcher.self}</span> · <a href="${relative(pageOutDir, outputFileOf(other, base)).split('\\').join('/')}">${loc.switcher.other}</a></span>`
}

// --- Template -------------------------------------------------------------------

function page(title, activeFile, contentHtml, locale, base, extraHead = '') {
  const loc = locales[locale]
  const outPath = outputFileOf(locale, base)
  const pageOutDir = dirname(outPath)
  const active = activeFile
    ? `document.querySelectorAll('.sidebar a[data-page]').forEach(a => { if (a.dataset.page === ${JSON.stringify(activeFile)}) a.classList.add('active') })`
    : ''
  const landingScript = loc.landingRedirect && base === 'index'
    ? `try{if(!sessionStorage.getItem('rn-lang')&&(navigator.language||'').toLowerCase().startsWith('zh')&&document.referrer.indexOf(location.host)===-1){sessionStorage.setItem('rn-lang','zh');location.replace('zh/index.html')}}catch(e){}`
    : ''
  return `<!DOCTYPE html>
<html lang="${loc.htmlLang}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${title ? title + ' · ' : ''}${loc.titleSuffix}</title>
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
    .lang-switch { display: block; margin-top: 6px; font-size: 11px; color: rgba(255,255,255,0.5); }
    .lang-switch .lang-current { font-weight: 700; color: #fff; }
    .lang-switch a { color: rgba(255,255,255,0.75); }
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
      <small>${loc.brandSmall}</small>
      ${switcherHtml(locale, base, pageOutDir)}
    </div>
    <nav>${navHtml(locale, pageOutDir)}
    </nav>
  </aside>
  <main class="main">
${contentHtml}
    <hr>
    <a href="${REPO}" class="back-to-top">${loc.backLabel}</a>
  </main>
  <script>${active}</script>
  <script>${landingScript}</script>
</body>
</html>
`
}

// --- Build -----------------------------------------------------------------------

function collectMdFiles(dir, skip) {
  const out = []
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    if (skip && p === skip) continue
    if (statSync(p).isDirectory()) out.push(...collectMdFiles(p, skip))
    else if (name.endsWith('.md')) out.push(p)
  }
  return out
}

// Generated HTML is never committed; wipe last build's output so pages whose
// sources moved or were removed can't linger at stale URLs.
function cleanGeneratedHtml(dir) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    if (statSync(p).isDirectory()) { cleanGeneratedHtml(p); continue }
    if (name.endsWith('.html')) rmSync(p)
  }
}

// Sanity gates: every sidebar entry needs a source in BOTH locales, otherwise
// the missing tree ships a broken nav by construction.
const sidebarFiles = new Set(sidebar.sections.flatMap(s => s.items.map(it => it.file)))
let failed = false
for (const locale of Object.keys(locales)) {
  const missing = [...sidebarFiles].filter(f => !sourceOf(locale, f))
  if (missing.length > 0) {
    for (const f of missing) console.error(`✗ sidebar 条目缺失 ${locale} 源文件: ${relative(root, join(loc.srcRoot, f + '.md'))}`)
    failed = true
  }
}
if (failed) { process.exitCode = 1 } else {

cleanGeneratedHtml(docsDir)
mkdirSync(join(docsDir, 'zh'), { recursive: true })

const built = []
for (const locale of Object.keys(locales)) {
  const loc = locales[locale]
  const sources = locale === 'en' ? collectMdFiles(loc.srcRoot) : collectMdFiles(docsDir, join(docsDir, 'en'))
  for (const mdPath of sources) {
    const rel = relative(loc.srcRoot, mdPath).replace(/\.md$/, '').split('\\').join('/')
    const base = rel
    const outPath = outputFileOf(locale, base)
    const pageOutDir = dirname(outPath)
    const { title, html } = renderMarkdown(mdPath, locale, pageOutDir)
    mkdirSync(pageOutDir, { recursive: true })
    if (base === 'index') {
      // Landing page: inject quick nav links into the <!--NAV_LINKS--> slot.
      const quick = sidebar.sections.flatMap(s => s.items).slice(0, 7)
        .map(it => `      <a href="${hrefFor(locale, it.file, pageOutDir)}">${labelFor(locale, it)}</a>`).join('\n')
      writeFileSync(outPath, page(title || loc.titleSuffix, '', html.replace('<!--NAV_LINKS-->', quick), locale, base))
    } else {
      writeFileSync(outPath, page(title, base, html, locale, base))
    }
    built.push(relative(root, outPath))
  }
}

console.log(`✓ 生成 ${built.length} 个页面（v${version}，en 默认 / zh 在 zh/）：\n  ${built.join('\n  ')}`)
}
