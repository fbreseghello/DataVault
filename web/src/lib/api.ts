export type ItemType = 'login' | 'api_key' | 'ssh_key' | 'seed' | 'note' | 'command'

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
  /** Sealed items need the seal PIN before their fields are shown. */
  sealed: boolean
  /** Renew every this many days, counted from the last renewal; 0 or absent is off. */
  renew_every?: number
  /** Renew by this date (YYYY-MM-DD); wins over the interval. */
  renew_by?: string
  /** When it was last renewed; only the server moves it. */
  renewed_at?: string
  created_at: string
  updated_at: string
}

export type Summary = Pick<Item, 'id' | 'type' | 'name' | 'tags' | 'sealed' | 'updated_at'> & {
  /** When the entry is due for renewal (YYYY-MM-DD), if ever. */
  renew_due?: string
}

export interface SessionInfo {
  id: string
  version: string
  unlocked: boolean
  auto_lock_minutes: number
  clipboard_clear_seconds: number
  /** Idle minutes on the login before the rest screen; 0 is off. */
  rest_minutes: number
  rest_style: RestStyle
}

/** What the rest screen shows: a still image, one of them at random, or the candle. */
export type RestStyle = 'pillar' | 'skull-i' | 'skull-ii' | 'random' | 'candle'

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

/** Header carrying the token that keeps one sealed item open. */
const UNSEAL_HEADER = 'X-Dtvlt-Unseal'

async function send(method: string, path: string, body?: unknown, token = '', keepalive = false): Promise<Response> {
  const headers: Record<string, string> = { 'X-Dtvlt': '1' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers[UNSEAL_HEADER] = token
  const res = await fetch(path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: 'same-origin',
    cache: 'no-store',
    keepalive,
  })
  return checked(res)
}

// checked raises an ApiError for any answer that is not ok, and tells the app
// when the vault locked.
async function checked(res: Response): Promise<Response> {
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
  return res
}

/** A document kept in the crypt; its contents are fetched only to take it out. */
export interface CryptFile {
  id: string
  title: string
  name: string
  type: string
  size: number
  created_at: string
  updated_at: string
}

/** The largest file the crypt keeps. */
export const MAX_FILE_BYTES = 25 * 1024 * 1024

/** Buries a file: its bytes go as the body, its name and title in headers. */
async function addFile(file: File, title = ''): Promise<CryptFile> {
  const res = await fetch('/api/files', {
    method: 'POST',
    headers: {
      'X-Dtvlt': '1',
      'Content-Type': file.type || 'application/octet-stream',
      'X-File-Name': encodeURIComponent(file.name),
      'X-File-Title': encodeURIComponent(title),
    },
    body: file,
    credentials: 'same-origin',
    cache: 'no-store',
  })
  return (await checked(res)).json()
}

/** Takes a file out of the crypt and hands it to the browser as a download. */
async function takeFile(f: CryptFile): Promise<void> {
  const res = await checked(await fetch(`/api/files/${encodeURIComponent(f.id)}/content`, { credentials: 'same-origin', cache: 'no-store' }))
  const url = URL.createObjectURL(await res.blob())
  const a = document.createElement('a')
  a.href = url
  a.download = f.name
  a.click()
  setTimeout(() => URL.revokeObjectURL(url), 10_000)
}

async function request<T>(method: string, path: string, body?: unknown, token = ''): Promise<T> {
  const res = await send(method, path, body, token)
  if (res.status === 204) return undefined as T
  return res.json()
}

/** Saves an item; when the result is sealed, the server also returns a token that keeps it open. */
async function saveItem(method: string, path: string, item: ItemInput, token: string): Promise<{ item: Item; token: string }> {
  const res = await send(method, path, item, token)
  return { item: await res.json(), token: res.headers.get(UNSEAL_HEADER) ?? token }
}

export type ItemInput = Omit<Item, 'id' | 'created_at' | 'updated_at'>

/** An Unspoken entry. */
export interface Note {
  id: string
  title: string
  body: string
  sealed: boolean
  buried: boolean
  created_at: string
  updated_at: string
}

/** A note as listed; sealed notes carry no title or preview. */
export interface NoteSummary {
  id: string
  title: string
  preview: string
  sealed: boolean
  buried: boolean
  created_at: string
  updated_at: string
}

export interface NoteInput {
  title: string
  body: string
  sealed: boolean
}

/** Saves a note; a sealed result comes with a token that keeps it open. */
async function saveNote(method: string, path: string, note: NoteInput, token: string, keepalive = false) {
  const res = await send(method, path, note, token, keepalive)
  return { note: (await res.json()) as Note, token: res.headers.get(UNSEAL_HEADER) ?? token }
}

const notePath = (id: string) => '/api/notes/' + encodeURIComponent(id)

/** Siege columns, in board order. */
export type Column = 'orders' | 'in_battle' | 'conquered' | 'fallen'

export interface Task {
  id: string
  title: string
  note: string
  column: Column
  /** The local date the task belongs to, YYYY-MM-DD. */
  day: string
  position: number
  created_at: string
  updated_at: string
}

export interface Board {
  tasks: Task[]
  /** Unfinished tasks left on earlier days, and the latest of those days. */
  pending_before: number
  pending_day: string
}

const taskPath = (id: string) => '/api/tasks/' + encodeURIComponent(id)

export type Cycle = 'weekly' | 'monthly' | 'yearly'

/** A ledger entry: a subscription or fixed expense. */
export interface Bill {
  id: string
  name: string
  /** In cents of currency. */
  amount: number
  /** ISO 4217 code, like BRL. */
  currency: string
  cycle: Cycle
  /** The first charge, YYYY-MM-DD; the next follow from it and the cycle. */
  since: string
  /** The card or account it is paid with. */
  method: string
  note: string
  created_at: string
  updated_at: string
}

export type BillInput = Omit<Bill, 'id' | 'created_at' | 'updated_at'>

const billPath = (id: string) => '/api/bills/' + encodeURIComponent(id)

export interface Settings {
  pin_enabled: boolean
  keep_terminal: boolean
  auto_lock_minutes: number
  clipboard_clear_seconds: number
  check_updates: boolean
  rest_minutes: number
  rest_style: RestStyle
  config_path: string
  vault_path: string
}

export type SettingsInput = Omit<Settings, 'config_path' | 'vault_path'> & {
  /** A new PIN; empty keeps the current one. */
  pin: string
}

export const api = {
  board: (day: string) => request<Board>('GET', '/api/tasks?day=' + encodeURIComponent(day)),
  createTask: (title: string, day: string, column: Column = 'orders') =>
    request<Task>('POST', '/api/tasks', { title, note: '', column, day }),
  updateTask: (id: string, title: string, note: string) => request<Task>('PUT', taskPath(id), { title, note }),
  moveTask: (id: string, column: Column, index: number) =>
    request<Task>('POST', taskPath(id) + '/move', { column, index }),
  deleteTask: (id: string) => request<void>('DELETE', taskPath(id)),
  carryTasks: (day: string) => request<{ moved: number }>('POST', '/api/tasks/carry', { day }),
  notes: (buried: boolean) => request<NoteSummary[]>('GET', '/api/notes' + (buried ? '?buried=1' : '')),
  note: (id: string, token = '') => request<Note>('GET', notePath(id), undefined, token),
  createNote: (note: NoteInput) => saveNote('POST', '/api/notes', note, ''),
  /** keepalive lets a save finish while the page is closing. */
  updateNote: (id: string, note: NoteInput, token = '', keepalive = false) =>
    saveNote('PUT', notePath(id), note, token, keepalive),
  buryNote: (id: string, buried: boolean, token = '') =>
    request<Note>('POST', notePath(id) + '/bury', { buried }, token),
  deleteNote: (id: string, token = '') => request<void>('DELETE', notePath(id), undefined, token),
  unsealNote: (id: string, pin: string) => request<{ note: Note; token: string }>('POST', notePath(id) + '/unseal', { pin }),
  session: () => request<SessionInfo>('GET', '/api/session'),
  unlock: (keyword: string, password: string) => request<void>('POST', '/api/unlock', { keyword, password }),
  lock: () => request<void>('POST', '/api/lock'),
  /** Ends the whole session: locks the vault and shuts dtvlt down. */
  endSession: () => request<void>('POST', '/api/session/end'),
  list: (q = '') => request<Summary[]>('GET', '/api/items' + (q ? '?q=' + encodeURIComponent(q) : '')),
  get: (id: string) => request<Item>('GET', '/api/items/' + encodeURIComponent(id)),
  create: (item: ItemInput) => saveItem('POST', '/api/items', item, ''),
  update: (id: string, item: ItemInput, token = '') => saveItem('PUT', '/api/items/' + encodeURIComponent(id), item, token),
  remove: (id: string, token = '') => request<void>('DELETE', '/api/items/' + encodeURIComponent(id), undefined, token),
  files: () => request<CryptFile[]>('GET', '/api/files'),
  addFile,
  takeFile,
  renameFile: (id: string, title: string) => request<CryptFile>('PUT', '/api/files/' + encodeURIComponent(id), { title }),
  removeFile: (id: string) => request<void>('DELETE', '/api/files/' + encodeURIComponent(id)),
  bills: () => request<Bill[]>('GET', '/api/bills'),
  createBill: (b: BillInput) => request<Bill>('POST', '/api/bills', b),
  updateBill: (id: string, b: BillInput) => request<Bill>('PUT', billPath(id), b),
  removeBill: (id: string) => request<void>('DELETE', billPath(id)),
  /** Marks an entry renewed today. */
  renewed: (id: string, token = '') => request<Item>('POST', '/api/items/' + encodeURIComponent(id) + '/renewed', undefined, token),
  unseal: (id: string, pin: string) =>
    request<{ item: Item; token: string }>('POST', '/api/items/' + encodeURIComponent(id) + '/unseal', { pin }),
  sealStatus: () => request<{ has_pin: boolean }>('GET', '/api/seal'),
  setSealPin: (current: string, pin: string) => request<void>('PUT', '/api/seal', { current, pin }),
  copied: (item: string, field: string) => request<void>('POST', '/api/events', { type: 'copy', item, field }),
  settings: () => request<Settings>('GET', '/api/settings'),
  saveSettings: (s: SettingsInput) => request<Settings>('PUT', '/api/settings', s),
}
