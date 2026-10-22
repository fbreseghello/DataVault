<script lang="ts">
  import { onMount } from 'svelte'
  import { api, ApiError } from './api'

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
      message = hasPin ? 'seal pin changed' : 'seal pin set'
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
  <form onsubmit={submit}>
    <h3>grimoire <span>the seal pin opens every sealed entry</span></h3>
    {#if hasPin}
      <label class="row">
        <span>current seal pin</span>
        <input type="password" bind:value={current} required autocomplete="current-password" />
        <span></span>
      </label>
    {/if}
    <label class="row">
      <span>{hasPin ? 'new seal pin' : 'seal pin'}</span>
      <input type="password" bind:value={next} minlength="4" required autocomplete="new-password" />
      <span class="unit">{hasPin ? '' : 'not set yet'}</span>
    </label>
    <label class="row">
      <span>confirm pin</span>
      <input type="password" bind:value={confirm} required autocomplete="new-password" />
      <span></span>
    </label>
    <div class="buttons">
      <button type="submit" disabled={busy}>{hasPin ? 'change seal pin' : 'set seal pin'}</button>
      {#if message}<span class="msg" role="status">{message}</span>{/if}
    </div>
  </form>
{/if}

<style>
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
  .row > span:first-child,
  .unit {
    color: var(--dim);
  }
  .unit {
    font-size: 12px;
  }
  .buttons {
    display: flex;
    gap: 16px;
    align-items: baseline;
    margin-top: 6px;
  }
  .msg {
    color: var(--accent);
    font-size: 12px;
  }
</style>
