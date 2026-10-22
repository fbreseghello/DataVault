<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { api, type SessionInfo } from './api'
  import art from './art.txt?raw'
  import Rest from './Rest.svelte'
  import { restRequest } from './restscreen.svelte'
  import { showVersion } from './version'

  let { info, onunlock }: { info: SessionInfo; onunlock: () => void } = $props()

  // One field at a time: the keyword first, then the password.
  let step = $state<'keyword' | 'password'>('keyword')
  let keyword = $state('')
  let password = $state('')
  let busy = $state(false)
  let error = $state('')
  let input = $state<HTMLInputElement>()

  // After rest_minutes without a key or a mouse move, the login rests on the
  // pillar screen; 0 turns that off (the shortcut still works).
  let resting = $state(false)
  let restTimer: ReturnType<typeof setTimeout> | undefined
  // Now and then the art stutters like the tube is striking again.
  let flickering = $state(false)

  function scheduleRest() {
    clearTimeout(restTimer)
    if (info.rest_minutes > 0) restTimer = setTimeout(rest, info.rest_minutes * 60_000)
  }

  // The shortcut asked for the rest screen (possibly while the vault was open).
  $effect(() => {
    if (restRequest.pending) {
      restRequest.pending = false
      rest()
    }
  })

  function rest() {
    // Whatever was typed is dropped, so nobody finds it half entered.
    keyword = password = ''
    step = 'keyword'
    error = ''
    resting = true
  }

  function wake() {
    resting = false
    scheduleRest()
    tick().then(() => input?.focus())
  }

  onMount(() => {
    const activity = () => {
      if (!resting) scheduleRest()
    }
    const events = ['mousemove', 'mousedown', 'keydown', 'wheel', 'touchstart'] as const
    for (const ev of events) window.addEventListener(ev, activity, { passive: true })
    scheduleRest()

    const calm = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    let flickerTimer: ReturnType<typeof setTimeout>
    const nextFlicker = () => {
      flickerTimer = setTimeout(
        () => {
          if (!resting && !calm) flickering = true
          nextFlicker()
        },
        20_000 + Math.random() * 25_000,
      )
    }
    nextFlicker()

    return () => {
      clearTimeout(restTimer)
      clearTimeout(flickerTimer)
      for (const ev of events) window.removeEventListener(ev, activity)
    }
  })

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

{#if resting}
  <Rest style={info.rest_style} onwake={wake} />
{:else}
<!-- Clicking anywhere returns focus to the hidden input, like a terminal. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<main onclick={() => input?.focus()}>
  <!-- Two copies of the art: the crisp tube on top, a blurred halo behind it that breathes. -->
  <div
    class="art"
    class:flicker={flickering}
    aria-hidden="true"
    onanimationend={(e) => {
      if (e.animationName.endsWith('flicker')) flickering = false
    }}
  >
    <pre class="halo">{art}</pre>
    <pre class="tube">{art}</pre>
  </div>
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

  <footer>session {info.id} · {showVersion(info.version)}</footer>
</main>
{/if}

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
    position: relative;
    margin: 0 0 56px;
    /* The art is 127 columns; shrink it to fit narrow windows. */
    font-size: min(12px, calc((100vw - 32px) / 127 / 0.6));
    line-height: 1.12;
    animation: power-on 1.6s linear both;
    user-select: none;
  }

  .art pre {
    margin: 0;
    font: inherit;
    line-height: inherit;
  }

  /* The tube: white letters with a glow that hugs them. */
  .tube {
    position: relative;
    color: var(--neon-core);
    text-shadow:
      0 0 1px var(--neon-core),
      0 0 4px var(--neon-glow);
  }

  /* The halo: the same art blurred behind, breathing slowly. Only its opacity
     animates, which is cheap for the browser. */
  .halo {
    position: absolute;
    inset: 0;
    color: var(--neon-haze);
    filter: blur(6px);
    opacity: 0.5;
    animation: breathe 4s ease-in-out 1.6s infinite;
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

  /* The tube strikes, stutters twice, then holds. */
  @keyframes power-on {
    0%,
    8% {
      opacity: 0;
    }
    10% {
      opacity: 0.7;
    }
    13% {
      opacity: 0.1;
    }
    18% {
      opacity: 0.9;
    }
    21% {
      opacity: 0.3;
    }
    28% {
      opacity: 1;
    }
    100% {
      opacity: 1;
    }
  }

  .art.flicker {
    animation: flicker 0.9s linear;
  }

  /* A shorter stutter than power-on: the tube never goes fully dark. */
  @keyframes flicker {
    0%,
    100% {
      opacity: 1;
    }
    8% {
      opacity: 0.25;
    }
    12% {
      opacity: 0.9;
    }
    18% {
      opacity: 0.35;
    }
    26% {
      opacity: 1;
    }
  }

  @keyframes breathe {
    0%,
    100% {
      opacity: 0.45;
    }
    50% {
      opacity: 0.8;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .art,
    .halo {
      animation: none;
    }
  }
</style>
