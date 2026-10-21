import type { Field, ItemType } from './api'

export const TYPES: { type: ItemType; label: string; short: string }[] = [
  { type: 'login', label: 'login', short: 'login' },
  { type: 'api_key', label: 'api key', short: 'key' },
  { type: 'ssh_key', label: 'ssh key', short: 'ssh' },
  { type: 'seed', label: 'seed / passphrase', short: 'seed' },
  { type: 'note', label: 'secure note', short: 'note' },
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
  }
}

const MULTILINE = new Set(['private_key', 'public_key', 'words'])

export function isMultiline(field: Field): boolean {
  return MULTILINE.has(field.name) || field.value.includes('\n')
}
