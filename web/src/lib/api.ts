export type ItemType = 'login' | 'api_key' | 'ssh_key' | 'seed' | 'note'

export interface Field {
  name: string
  value: string
  secret: boolean
}

export interface Item {
  id: string
  type: ItemType
  name: string
  fields: Field[]
  tags: string[]
  notes: string
  created_at: string
  updated_at: string
}

export type Summary = Pick<Item, 'id' | 'type' | 'name' | 'tags' | 'updated_at'>

export interface SessionInfo {
  id: string
  unlocked: boolean
  auto_lock_minutes: number
  clipboard_clear_seconds: number
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
  }
}

/** Fired whenever the server answers 423, so the app can fall back to the login. */
export const LOCKED_EVENT = 'dtvlt:locked'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'X-Dtvlt': '1' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  const res = await fetch(path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: 'same-origin',
    cache: 'no-store',
  })
  if (res.status === 423) window.dispatchEvent(new Event(LOCKED_EVENT))
  if (!res.ok) {
    let msg = res.statusText
    try {
      msg = (await res.json()).error ?? msg
    } catch {
      // not JSON
    }
    throw new ApiError(res.status, msg)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export type ItemInput = Omit<Item, 'id' | 'created_at' | 'updated_at'>

export const api = {
  session: () => request<SessionInfo>('GET', '/api/session'),
  unlock: (keyword: string, password: string) => request<void>('POST', '/api/unlock', { keyword, password }),
  lock: () => request<void>('POST', '/api/lock'),
  list: (q = '') => request<Summary[]>('GET', '/api/items' + (q ? '?q=' + encodeURIComponent(q) : '')),
  get: (id: string) => request<Item>('GET', '/api/items/' + encodeURIComponent(id)),
  create: (item: ItemInput) => request<Item>('POST', '/api/items', item),
  update: (id: string, item: ItemInput) => request<Item>('PUT', '/api/items/' + encodeURIComponent(id), item),
  remove: (id: string) => request<void>('DELETE', '/api/items/' + encodeURIComponent(id)),
  copied: (item: string, field: string) => request<void>('POST', '/api/events', { type: 'copy', item, field }),
}
