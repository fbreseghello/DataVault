<script lang="ts">
  import { tick } from 'svelte'
  import { ApiError } from './api'
  import Placard from './Placard.svelte'

  // A sealed entry waiting for the seal PIN, typed on one blinking line like the login.
  let { name, onunseal }: { name: string; onunseal: (pin: string) => Promise<void> } = $props()

  let pin = $state('')
  let busy = $state(false)
  let error = $state('')
  let input = $state<HTMLInputElement>()

  $effect(() => {
    void name
    tick().then(() => input?.focus())
  })

  async function onkeydown(e: KeyboardEvent) {
    if (e.key !== 'Enter' || !pin || busy) return
    e.preventDefault()
    busy = true
    error = ''
    try {
      await onunseal(pin)
    } catch (err) {
      error = err instanceof ApiError && err.status === 403 ? 'wrong pin' : err instanceof Error ? err.message : String(err)
    } finally {
      pin = ''
      busy = false
      await tick()
      input?.focus()
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="gate" onclick={() => input?.focus()}>
  <Placard icon="seal" title={name}>
    <div class="prompt" aria-hidden="true"><span class="cursor" class:busy>_</span></div>
    <label class="sr-only" for="seal-pin">seal pin for {name}</label>
    <input id="seal-pin" bind:this={input} bind:value={pin} type="password" autocomplete="off" disabled={busy} {onkeydown} />
    <p class="hint" aria-live="polite">
      {#if busy}checking{:else if error}<span class="error">{error}</span>{:else}seal pin{/if}
    </p>
  </Placard>
</div>

<style>
  .gate {
    height: 100%;
    cursor: text;
  }
  .prompt {
    margin-top: 24px;
    min-width: 200px;
    height: 1.6em;
    border-bottom: 1px solid var(--line);
    font-size: 15px;
  }
  .cursor {
    animation: blink var(--blink) steps(1) infinite;
  }
  .cursor.busy {
    animation: none;
    color: var(--dim);
  }
  input {
    position: absolute;
    opacity: 0;
    width: 1px;
    height: 1px;
    pointer-events: none;
  }
  .hint {
    margin: 10px 0 0;
    color: var(--dim);
    font-size: 12px;
  }
  .error {
    color: var(--accent);
  }
</style>
