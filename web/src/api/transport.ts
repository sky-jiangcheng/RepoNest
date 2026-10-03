// api/transport.ts — Dual-mode transport:
//   - Inside the Wails webview, invokes Go methods via window.go.app.App
//     (package "app" / struct "App" -> go.app.App.MethodName)
//   - Standalone (npm run dev in a plain browser), falls back to HTTP fetch

import type { ImportCompletedEvent } from './types'

interface WailsApp {
  [method: string]: (...args: unknown[]) => Promise<unknown>
}

interface WailsGlobal {
  go?: {
    app?: {
      App?: WailsApp
    }
  }
}

const isWails = (): boolean => {
  if (typeof window === 'undefined') return false
  const w = window as unknown as WailsGlobal
  return !!w.go?.app?.App
}

function wail<T>(method: string, ...args: unknown[]): Promise<T> {
  const w = window as unknown as WailsGlobal
  const app = w.go?.app?.App
  if (!app) {
    throw new Error('Wails runtime not available')
  }
  const fn = app[method]
  if (typeof fn !== 'function') {
    throw new Error(`Wails method not found: ${method}`)
  }
  return fn(...args) as Promise<T>
}

const BASE = '/api'

// Browser/standalone transport: a single JSON-RPC bridge to the bound App
// (httpapi /api/rpc), so every Wails binding works over HTTP without a REST
// route per method. Desktop mode uses window.go.app.App directly (wail()).
async function rpc<T>(method: string, args: unknown[]): Promise<T> {
  const res = await fetch(BASE + '/rpc', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ method, args }),
  })
  const data = await res.json().catch(() => ({ error: 'HTTP ' + res.status }))
  if (!res.ok || data.error) throw new Error(data.error || `RPC ${method} failed`)
  return data.result as T
}

/** A single call routed through whichever transport is active. */
export function call<T>(opts: {
  /** Wails binding method name. */
  method: string
  /** Arguments for the Wails binding call. */
  args?: unknown[]
  /** HTTP path (relative to /api) — unused in RPC mode, kept for compatibility. */
  path?: string
  /** HTTP request options — unused in RPC mode. */
  init?: RequestInit
}): Promise<T> {
  if (isWails()) return wail<T>(opts.method, ...(opts.args ?? []))
  return rpc<T>(opts.method, opts.args ?? [])
}

// --- Connection health tracking ------------------------------------------------

export type ConnectionKind = 'ok' | 'offline' | 'backend-down'

let currentKind: ConnectionKind = 'ok'
const listeners = new Set<(kind: ConnectionKind) => void>()

export function getConnectionKind(): ConnectionKind { return currentKind }

export function subscribeConnection(cb: (kind: ConnectionKind) => void): () => void {
  listeners.add(cb)
  cb(currentKind)
  return () => { listeners.delete(cb) }
}

function notify(kind: ConnectionKind) {
  if (kind === currentKind) return
  currentKind = kind
  for (const l of listeners) l(kind)
}

/** Health check: resolves true when backend responds, false otherwise. */
export function checkHealth(): Promise<boolean> {
  if (!isWails()) return Promise.resolve(true)
  return wail<{ ok?: boolean; status?: string }>('Health').then(r => {
    // The Go binding returns status: "ok"; accept the legacy ok boolean too.
    const ok = r?.ok ?? r?.status === 'ok'
    notify(ok ? 'ok' : 'backend-down')
    return ok
  }).catch(() => {
    notify('backend-down')
    return false
  })
}

/** Start periodic health polling. Returns a cleanup function. */
export function startHealthPoll(intervalMs = 30_000): () => void {
  let timer: number | null = null
  let stopped = false
  const schedule = () => {
    if (!stopped) timer = window.setTimeout(tick, intervalMs) as unknown as number
  }
  const tick = () => {
    checkHealth().finally(schedule)
  }
  tick()
  return () => {
    stopped = true
    if (timer !== null) clearTimeout(timer)
  }
}

/** Register online/offline listeners so we detect network state changes. */
function initOnlineDetection() {
  const update = () => {
    if (!navigator.onLine) { notify('offline'); return }
    checkHealth()
  }
  window.addEventListener('online', update)
  window.addEventListener('offline', update)
}
initOnlineDetection()

// --- Wails runtime events -----------------------------------------------------

interface RuntimeGlobal {
  runtime?: {
    EventsOn?: (name: string, cb: (data: unknown) => void) => () => void
  }
}

/**
 * Subscribe to the import.completed Wails event. Returns an unsubscribe
 * function when the runtime provides one (Wails v2.5+); callers must invoke
 * it on cleanup to avoid stacking duplicate listeners.
 */
export function listenImportCompleted(cb: (data: ImportCompletedEvent) => void): () => void {
  const w = window as unknown as RuntimeGlobal
  if (w.runtime?.EventsOn) {
    const off = w.runtime.EventsOn('import.completed', (data) => {
      const d = (data ?? {}) as Partial<ImportCompletedEvent>
      cb({
        source: d.source ?? 'unknown',
        created: d.created ?? 0,
        updated: d.updated ?? 0,
        skipped: d.skipped ?? 0,
        error: d.error,
      })
    })
    return () => off?.()
  }
  return () => {}
}
