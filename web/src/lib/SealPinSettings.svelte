<script lang="ts">
  import { onMount } from 'svelte'
  import { api, ApiError } from './api'
  import { word } from './theme.svelte'

  // Sets or changes the seal PIN, which lives inside the vault, not in config.toml.
  let hasPin = $state<boolean | null>(null)
  let current = $state('')
  let next = $state('')
  let confirm = $state('')
  let busy = $state(false)
  let message = $state('')

  onMount(() => {
    api.sealStatus().then((r) => (hasPin = r.has_pin), () => {})
  })

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    message = ''
    if (next !== confirm) {
      message = "the pins don't match"
      return
    }
    busy = true
    try {
      await api.setSealPin(current, next)
      message = hasPin ? `${word('seal')} pin changed` : `${word('seal')} pin set`
      hasPin = true
      current = next = confirm = ''
    } catch (err) {
      message = err instanceof ApiError && err.status === 403 ? 'wrong current pin' : err instanceof Error ? err.message : String(err)
    } finally {
      busy = false
    }
  }
</script>

{#if hasPin !== null}
  <form class="set-group" onsubmit={submit} aria-labelledby="set-seal">
    <header>
      <h3 id="set-seal">{word('grimoire')}</h3>
      <p>the {word('seal')} pin opens every {word('sealed')} entry; it saves on its own</p>
    </header>
    <div class="set-fields">
      {#if hasPin}
        <label class="set-field">
          <span class="set-label">current {word('seal')} pin</span>
          <span class="set-control">
            <input type="password" bind:value={current} required autocomplete="current-password" />
          </span>
        </label>
      {/if}
      <label class="set-field">
        <span class="set-label">{hasPin ? `new ${word('seal')} pin` : `${word('seal')} pin`}</span>
        <span class="set-control">
          <input type="password" bind:value={next} minlength="4" required autocomplete="new-password" />
        </span>
        {#if !hasPin}<span class="set-hint">not set yet</span>{/if}
      </label>
      <label class="set-field">
        <span class="set-label">confirm pin</span>
        <span class="set-control">
          <input type="password" bind:value={confirm} required autocomplete="new-password" />
        </span>
      </label>
      <div class="set-field">
        <span class="set-actions">
          <button type="submit" disabled={busy}>{hasPin ? `change ${word('seal')} pin` : `set ${word('seal')} pin`}</button>
          {#if message}<span class="set-note" role="status">{message}</span>{/if}
        </span>
      </div>
    </div>
  </form>
{/if}
