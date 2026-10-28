import type { Field, ItemType } from './api'

export const TYPES: { type: ItemType; label: string; short: string }[] = [
  { type: 'login', label: 'login', short: 'login' },
  { type: 'api_key', label: 'api key', short: 'key' },
  { type: 'ssh_key', label: 'ssh key', short: 'ssh' },
  { type: 'seed', label: 'seed / passphrase', short: 'seed' },
  { type: 'note', label: 'secure note', short: 'note' },
  { type: 'command', label: 'server command', short: 'cmd' },
  { type: 'recovery', label: 'recovery codes', short: 'codes' },
]

export function typeShort(t: ItemType): string {
  return TYPES.find((x) => x.type === t)?.short ?? t
}

/** Starting fields for a new item of each type. */
export function template(t: ItemType): Field[] {
  const f = (name: string, secret = false): Field => ({ name, value: '', secret })
  switch (t) {
    case 'login':
      return [f('username'), f('password', true), f('url')]
    case 'api_key':
      return [f('key', true), f('scope')]
    case 'ssh_key':
      return [f('private_key', true), f('public_key'), f('passphrase', true)]
    case 'seed':
      return [f('words', true)]
    case 'note':
      return []
    case 'command':
      return [f('command'), f('server'), f('workdir')]
    case 'recovery':
      return [f('account'), f('code 1', true)]
  }
}

const MULTILINE = new Set(['private_key', 'public_key', 'words', 'command', 'codes'])

/** The field of a command entry shown as a terminal line, with a prompt. */
export function isCommand(type: ItemType, field: Field): boolean {
  return type === 'command' && field.name === 'command'
}

export function commandLines(value: string): string[] {
  return value.replace(/\r\n/g, '\n').split('\n')
}

export function isMultiline(field: Field): boolean {
  return MULTILINE.has(field.name) || field.value.includes('\n')
}

// Recovery codes: each in a field of its own, marked dead once spent. Entries
// from 0.15.0 may keep them all in one "codes" field, one per line; a spent one
// there is a line that starts with "dead" (or "used", as 0.15.0 wrote it).
const USED = /^(dead|used)\s+/i

export interface RecoveryCode {
  code: string
  used: boolean
}

/** The one-per-line codes field of a recovery entry. */
export function isRecovery(type: ItemType, field: Field): boolean {
  return type === 'recovery' && field.name === 'codes' && field.value.includes('\n')
}

/** A field of a recovery entry that holds a single code. */
export function isCode(type: ItemType, field: Field): boolean {
  return type === 'recovery' && field.name !== 'account' && !isRecovery(type, field)
}

/** How many of an entry's recovery codes are still alive, out of how many. */
export function codesLeft(type: ItemType, fields: Field[]): { left: number; total: number } {
  let left = 0
  let total = 0
  for (const f of fields) {
    if (!f.value) continue
    if (isRecovery(type, f)) {
      const codes = recoveryCodes(f.value)
      total += codes.length
      left += codes.filter((c) => !c.used).length
    } else if (isCode(type, f)) {
      total++
      if (!f.dead) left++
    }
  }
  return { left, total }
}

/** The name for the next code field: "code 4" after three codes. */
export function nextCodeName(type: ItemType, fields: Field[]): string {
  return `code ${fields.filter((f) => isCode(type, f)).length + 1}`
}

export function recoveryCodes(value: string): RecoveryCode[] {
  return commandLines(value)
    .map((l) => l.trim())
    .filter(Boolean)
    .map((l) => ({ code: l.replace(USED, ''), used: USED.test(l) }))
}

/** The field value with the n-th code marked used, or unused again. */
export function toggleUsed(value: string, n: number): string {
  return recoveryCodes(value)
    .map((c, k) => ((k === n ? !c.used : c.used) ? `dead ${c.code}` : c.code))
    .join('\n')
}

/** Few enough codes left that it is time to generate new ones. */
export const LOW_CODES = 2
