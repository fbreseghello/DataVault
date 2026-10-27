<script lang="ts">
  import { api } from './api'
  import { copy } from './clipboard'
  import { bits, DEFAULTS, forge, MAX_LENGTH, MIN_LENGTH, type ForgeOptions } from './forge'

  // The password generator of the grimoire. A new password is forged whenever
  // an option changes; "again" forges another with the same options.
  let { clipboardSeconds, onclose }: { clipboardSeconds: number; onclose: () => void } = $props()

  let opts = $state<ForgeOptions>({ ...DEFAULTS })
  let password = $state('')
  let copied = $state(false)
  let error = $state('')

  const strength = $derived(bits(opts))
  const none = $derived(!opts.upper && !opts.lower && !opts.digits && !opts.symbols)

  function again() {
    password = forge(opts)
    copied = false
  }

  // forge anew whenever an option changes
  $effect(() => {
    void [opts.length, opts.upper, opts.lower, opts.digits, opts.symbols, opts.plain]
    again()
  })

  async function copyIt() {
    if (!password) return
    error = ''
    try {
      await copy(password, clipboardSeconds)
    } catch {
      error = 'clipboard not available'
      return
    }
    copied = true
    setTimeout(() => (copied = false), 1500)
    api.copied('forge', 'password').catch(() => {})
  }

  function onkeydown(e: KeyboardEvent) {
    const t = e.target
    if (t instanceof HTMLInputElement && t.type !== 'checkbox' && t.type !== 'range') return
    if (e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === 'r') {
      e.preventDefault()
      again()
    } else if (e.key === 'c') {
      e.preventDefault()
      copyIt()
    }
  }

  const KINDS = [
    { key: 'upper', label: 'A–Z' },
    { key: 'lower', label: 'a–z' },
    { key: 'digits', label: '0–9' },
    { key: 'symbols', label: '!#$' },
  ] as const
</script>

<svelte:window {onkeydown} />

<div class="forge">
  <h2>forge <span class="dim">a new key</span></h2>

  <p class="password" class:empty={!password} aria-live="polite">{password || 'choose at least one kind of character'}</p>

  <div class="row">
    <span class="label">length</span>
    <input type="range" min={MIN_LENGTH} max={MAX_LENGTH} bind:value={opts.length} aria-label="length" />
    <input class="number" type="number" min={MIN_LENGTH} max={MAX_LENGTH} bind:value={opts.length} aria-label="length" />
  </div>

  <div class="row">
    <span class="label">characters</span>
    <div class="kinds">
      {#each KINDS as k (k.key)}
        <button type="button" class:on={opts[k.key]} aria-pressed={opts[k.key]} onclick={() => (opts[k.key] = !opts[k.key])}>{k.label}</button>
      {/each}
    </div>
    <span></span>
  </div>

  <div class="row">
    <span class="label">no lookalikes</span>
    <button type="button" class="toggle" class:on={opts.plain} aria-pressed={opts.plain} onclick={() => (opts.plain = !opts.plain)}>
      {opts.plain ? 'on' : 'off'}
    </button>
    <span class="unit">{opts.plain ? 'leaves out 0 O o 1 l I |' : 'easier to type by hand when on'}</span>
  </div>

  <p class="meta">{none ? '' : `${strength} bits · from the browser's cryptographic randomness · never saved`}</p>

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="buttons">
    <button class="primary" onclick={copyIt} disabled={!password}>{copied ? 'copied' : 'copy'}</button>
    <button onclick={again} disabled={none}>again</button>
    <button onclick={onclose}>close</button>
  </div>
  <p class="meta keys">c copy · r again · esc close</p>
</div>

<style>
  h2 {
    margin: 0 0 20px;
    font-size: 15px;
    font-weight: 700;
  }

  .dim {
    color: var(--dim);
    font-weight: 400;
    font-size: 13px;
    margin-left: 8px;
  }

  .password {
    margin: 0 0 24px;
    padding: 14px 0;
    border-top: 1px solid var(--line);
    border-bottom: 1px solid var(--line);
    font-size: 17px;
    letter-spacing: 0.04em;
    overflow-wrap: anywhere;
    user-select: all;
  }

  .password.empty {
    color: var(--dim);
    font-size: 13px;
    letter-spacing: 0;
  }

  .row {
    display: grid;
    grid-template-columns: 130px minmax(0, 1fr) 70px;
    gap: 14px;
    align-items: center;
    margin-bottom: 12px;
  }

  .label,
  .unit {
    color: var(--dim);
  }

  .unit {
    grid-column: 3 / 4;
    font-size: 12px;
    white-space: nowrap;
  }

  input[type='range'] {
    width: 100%;
    accent-color: var(--muted);
  }

  .number {
    width: 100%;
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--line);
    padding: 4px 0;
    outline: none;
    text-align: right;
  }

  .number:focus {
    border-bottom-color: var(--muted);
  }

  .kinds {
    display: flex;
    gap: 16px;
  }

  .kinds button:not(.on) {
    color: var(--dim);
  }

  .kinds .on,
  .toggle.on,
  .error {
    color: var(--accent);
  }

  .toggle {
    justify-self: start;
  }

  .meta {
    margin: 16px 0 0;
    color: var(--dim);
    font-size: 12px;
    min-height: 1.6em;
  }

  .keys {
    margin-top: 10px;
  }

  .buttons {
    display: flex;
    gap: 20px;
    margin-top: 20px;
  }

  .primary {
    color: var(--fg);
  }
</style>
