<script lang="ts">
  import { onMount } from 'svelte'
  import { api, ApiError, type RestStyle, type Settings } from './api'
  import SealPinSettings from './SealPinSettings.svelte'

  let { onsaved, onclose }: { onsaved: () => void; onclose: () => void } = $props()

  let saved = $state<Settings | null>(null)
  let pinEnabled = $state(false)
  let keepTerminal = $state(true)
  let autoLock = $state(15)
  let clipboard = $state(20)
  let checkUpdates = $state(true)
  let rest = $state(15)
  let restStyle = $state<RestStyle>('pillar')
  const restHints: Record<RestStyle, string> = {
    pillar: 'the broken pillar',
    'skull-i': 'the skull, drawn in lines',
    'skull-ii': 'the skull, drawn in dust',
    random: 'one of the three, drawn each time it rests',
  }
  let pin = $state('')
  let pinConfirm = $state('')
  let busy = $state(false)
  let error = $state('')
  let done = $state(false)

  // A PIN must be typed when turning it on; when it is already on, typing one changes it.
  const needsPin = $derived(pinEnabled && !saved?.pin_enabled)

  function load(s: Settings) {
    saved = s
    pinEnabled = s.pin_enabled
    keepTerminal = s.keep_terminal
    autoLock = s.auto_lock_minutes
    clipboard = s.clipboard_clear_seconds
    checkUpdates = s.check_updates
    rest = s.rest_minutes
    restStyle = s.rest_style
    pin = pinConfirm = ''
  }

  onMount(() => {
    api.settings().then(load, (e) => (error = e instanceof Error ? e.message : String(e)))
  })

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    error = ''
    done = false
    if (pinEnabled && (needsPin || pin) && pin !== pinConfirm) {
      error = "the pins don't match"
      return
    }
    busy = true
    try {
      load(
        await api.saveSettings({
          pin_enabled: pinEnabled,
          pin: pinEnabled ? pin : '',
          keep_terminal: keepTerminal,
          auto_lock_minutes: autoLock,
          clipboard_clear_seconds: clipboard,
          check_updates: checkUpdates,
          rest_minutes: rest,
          rest_style: restStyle,
        }),
      )
      done = true
      onsaved()
    } catch (err) {
      if (!(err instanceof ApiError && err.status === 423)) error = err instanceof Error ? err.message : String(err)
    } finally {
      busy = false
    }
  }
</script>

<form onsubmit={submit}>
  <h2>settings</h2>

  {#if saved}
    <h3>session <span>applies now</span></h3>

    <label class="row">
      <span>auto-lock</span>
      <input type="number" min="1" max="1440" bind:value={autoLock} required />
      <span class="unit">minutes idle</span>
    </label>
    <label class="row">
      <span>clear clipboard</span>
      <input type="number" min="5" max="600" bind:value={clipboard} required />
      <span class="unit">seconds after copy</span>
    </label>
    <label class="row">
      <span>rest screen</span>
      <input type="number" min="0" max="240" bind:value={rest} required />
      <span class="unit">{rest === 0 ? 'off · ctrl+shift+l still rests' : 'minutes idle on the login · ctrl+shift+l rests now'}</span>
    </label>
    <label class="row">
      <span>rest style</span>
      <select bind:value={restStyle}>
        {#each Object.keys(restHints) as style (style)}
          <option value={style}>{style}</option>
        {/each}
      </select>
      <span class="unit">{restHints[restStyle]}</span>
    </label>

    <h3>terminal <span>applies the next time dtvlt starts</span></h3>

    <div class="row">
      <span>pin on start</span>
      <button type="button" class="toggle" class:on={pinEnabled} onclick={() => (pinEnabled = !pinEnabled)} aria-pressed={pinEnabled}>
        {pinEnabled ? 'on' : 'off'}
      </button>
      <span class="unit">only gates starting dtvlt, it does not encrypt</span>
    </div>
    {#if pinEnabled}
      <label class="row">
        <span>{needsPin ? 'pin' : 'new pin'}</span>
        <input type="password" bind:value={pin} minlength="4" required={needsPin} autocomplete="new-password" placeholder={needsPin ? '' : 'keep current'} />
        <span></span>
      </label>
      {#if needsPin || pin}
        <label class="row">
          <span>confirm pin</span>
          <input type="password" bind:value={pinConfirm} required autocomplete="new-password" />
          <span></span>
        </label>
      {/if}
    {/if}
    <div class="row">
      <span>keep terminal open</span>
      <button type="button" class="toggle" class:on={keepTerminal} onclick={() => (keepTerminal = !keepTerminal)} aria-pressed={keepTerminal}>
        {keepTerminal ? 'on' : 'off'}
      </button>
      <span class="unit">{keepTerminal ? 'the session logs to the terminal' : 'the session runs in the background'}</span>
    </div>
    <div class="row">
      <span>check for updates</span>
      <button type="button" class="toggle" class:on={checkUpdates} onclick={() => (checkUpdates = !checkUpdates)} aria-pressed={checkUpdates}>
        {checkUpdates ? 'on' : 'off'}
      </button>
      <span class="unit">on each start, asks github.com for a newer version before the pin</span>
    </div>

    <h3>files</h3>
    <div class="row">
      <span>config</span>
      <span class="path">{saved.config_path}</span>
      <span></span>
    </div>
    <div class="row">
      <span>vault</span>
      <span class="path">{saved.vault_path}</span>
      <span></span>
    </div>
  {/if}

  {#if error}<p class="msg error" role="alert">{error}</p>{/if}
  {#if done}<p class="msg ok" role="status">saved</p>{/if}

  <div class="buttons">
    <button type="submit" class="primary" disabled={busy || !saved}>{busy ? 'saving' : 'save'}</button>
    <button type="button" onclick={onclose}>close</button>
  </div>
</form>

<!-- A separate form: the seal PIN is stored in the vault and saved on its own. -->
<SealPinSettings />

<style>
  h2 {
    margin: 0 0 8px;
    font-size: 13px;
    font-weight: 700;
  }
  h3 {
    margin: 24px 0 10px;
    font-size: 12px;
    font-weight: 400;
    color: var(--muted);
  }
  h3 span {
    color: var(--dim);
    margin-left: 8px;
  }
  .row {
    display: grid;
    grid-template-columns: 150px minmax(0, 180px) 1fr;
    gap: 12px;
    align-items: baseline;
    margin-bottom: 10px;
  }
  .row > span:first-child {
    color: var(--dim);
  }
  .unit {
    color: var(--dim);
    font-size: 12px;
  }
  input[type='number'] {
    width: 100%;
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--line);
    padding: 4px 0;
    outline: none;
  }
  input[type='number']:focus {
    border-bottom-color: var(--muted);
  }
  .toggle {
    justify-self: start;
  }
  .toggle.on {
    color: var(--accent);
  }
  select {
    cursor: pointer;
  }
  .path {
    grid-column: 2 / 4;
    color: var(--muted);
    overflow-wrap: anywhere;
    user-select: text;
  }
  .msg {
    margin: 16px 0 0;
  }
  .error,
  .ok {
    color: var(--accent);
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
