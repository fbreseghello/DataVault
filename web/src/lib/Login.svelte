<script lang="ts">
  import { tick } from 'svelte'
  import { api, type SessionInfo } from './api'
  import art from './art.txt?raw'

  let { info, onunlock }: { info: SessionInfo; onunlock: () => void } = $props()

  // One field at a time: the keyword first, then the password.
  let step = $state<'keyword' | 'password'>('keyword')
  let keyword = $state('')
  let password = $state('')
  let busy = $state(false)
  let error = $state('')
  let input = $state<HTMLInputElement>()

  $effect(() => {
    void step
    tick().then(() => input?.focus())
  })

  async function submit() {
    busy = true
    error = ''
    try {
      await api.unlock(keyword, password)
      keyword = password = ''
      onunlock()
    } catch {
      // Never say which of the two was wrong.
      error = 'invalid credentials'
      keyword = password = ''
      step = 'keyword'
    } finally {
      busy = false
      await tick()
      input?.focus()
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (step === 'keyword' && keyword) {
        step = 'password'
        error = ''
      } else if (step === 'password' && password && !busy) {
        void submit()
      }
    } else if (e.key === 'Escape' && step === 'password' && !busy) {
      password = ''
      step = 'keyword'
    }
  }
</script>

<!-- Clicking anywhere returns focus to the hidden input, like a terminal. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<main onclick={() => input?.focus()}>
  <pre class="art" aria-hidden="true">{art}</pre>
  <h1 class="sr-only">DataVault</h1>

  <div class="prompt" aria-hidden="true">
    {#if step === 'keyword'}<span class="typed">{'•'.repeat(keyword.length)}</span>{/if}<span
      class="cursor"
      class:busy>_</span
    >
  </div>

  <form onsubmit={(e) => e.preventDefault()}>
    {#if step === 'keyword'}
      <label class="sr-only" for="secret">keyword</label>
      <input
        id="secret"
        bind:this={input}
        bind:value={keyword}
        type="password"
        autocomplete="username"
        spellcheck="false"
        disabled={busy}
        {onkeydown}
      />
    {:else}
      <label class="sr-only" for="secret">password</label>
      <input
        id="secret"
        bind:this={input}
        bind:value={password}
        type="password"
        autocomplete="current-password"
        disabled={busy}
        {onkeydown}
      />
    {/if}
  </form>

  <p class="hint" aria-live="polite">
    {#if busy}deriving key{:else if error}<span class="error">{error}</span>{:else}{step}{/if}
  </p>

  <footer>session {info.id}</footer>
</main>

<style>
  main {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 16px;
    cursor: text;
  }

  .art {
    margin: 0 0 56px;
    /* The art is 121 columns; shrink it to fit narrow windows. */
    font-size: min(12px, calc((100vw - 32px) / 121 / 0.6));
    line-height: 1.12;
    color: var(--fg);
    animation: appear 1.8s ease-out both;
    user-select: none;
  }

  .prompt {
    min-width: 280px;
    height: 1.6em;
    padding-bottom: 2px;
    border-bottom: 1px solid var(--line);
    text-align: center;
    font-size: 15px;
    letter-spacing: 0.15em;
  }

  .typed {
    color: var(--muted);
  }

  .cursor {
    animation: blink var(--blink) steps(1) infinite;
  }

  .cursor.busy {
    animation: none;
    color: var(--dim);
  }

  /* The real input stays focusable for keyboards, password managers and
     screen readers, but what you see is the line above. */
  form input {
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
    min-height: 1.6em;
  }

  .error {
    color: var(--accent);
  }

  footer {
    position: fixed;
    bottom: 16px;
    right: 16px;
    color: var(--dim);
    font-size: 11px;
  }

  @keyframes appear {
    from {
      opacity: 0;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .art {
      animation: none;
    }
  }
</style>
