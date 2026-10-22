import type { Field, ItemType } from './api'

export const TYPES: { type: ItemType; label: string; short: string }[] = [
  { type: 'login', label: 'login', short: 'login' },
  { type: 'api_key', label: 'api key', short: 'key' },
  { type: 'ssh_key', label: 'ssh key', short: 'ssh' },
  { type: 'seed', label: 'seed / passphrase', short: 'seed' },
  { type: 'note', label: 'secure note', short: 'note' },
  { type: 'command', label: 'server command', short: 'cmd' },
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
  }
}

const MULTILINE = new Set(['private_key', 'public_key', 'words', 'command'])

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
