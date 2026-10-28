<script lang="ts">
  import type { SessionInfo } from './api'
  import Crypt from './Crypt.svelte'
  import Grimoire from './Grimoire.svelte'
  import Icon from './Icon.svelte'
  import Ledger from './Ledger.svelte'
  import Settings from './Settings.svelte'
  import Siege from './Siege.svelte'
  import { theme, word } from './theme.svelte'
  import Unspoken from './Unspoken.svelte'
  import { showVersion } from './version'

  let {
    info,
    onlock,
    onrefresh,
    onend,
  }: { info: SessionInfo; onlock: () => void; onrefresh: () => void; onend: () => void } = $props()

  // Each page draws its own icon slot and takes its name from the theme.
  type Page = { id: 'grimoire' | 'unspoken' | 'siege' | 'crypt' | 'ledger'; label: string; size: number }

  // The sidebar shows icons only; label is for the hover tip and screen readers.
  // size is a multiple of 16 so the art stays crisp; a theme draws an icon that
  // leaves more of its grid empty a step larger, to stand level with the others.
  const pages: Page[] = $derived(
    (['grimoire', 'unspoken', 'siege', 'crypt', 'ledger'] as const).map((id) => ({
      id,
      label: word(id),
      size: theme.current?.large_icons?.includes(id) ? 48 : 32,
    })),
  )

  let page = $state('grimoire')
  let settings = $state(false)
  const current = $derived(pages.find((p) => p.id === page)!)

  function go(id: string) {
    page = id
    settings = false
  }

  // Ending the session shuts dtvlt down (and its console, when it runs in a
  // terminal), so the tombstone asks for a second click within a few seconds.
  let ending = $state(false)
  let endTimer: ReturnType<typeof setTimeout> | undefined

  function endClick() {
    if (!ending) {
      ending = true
      clearTimeout(endTimer)
      endTimer = setTimeout(() => (ending = false), 3000)
      return
    }
    clearTimeout(endTimer)
    onend()
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && settings) {
      settings = false
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    const n = Number(e.key)
    if (n >= 1 && n <= pages.length) {
      e.preventDefault()
      go(pages[n - 1].id)
    }
  }
</script>

<svelte:window {onkeydown} />

<div class="shell">
  <aside>
    <nav aria-label="pages">
      {#each pages as p, i (p.id)}
        <button
          class="page"
          class:active={!settings && page === p.id}
          onclick={() => go(p.id)}
          aria-label={p.label}
          aria-current={!settings && page === p.id ? 'page' : undefined}
        >
          <Icon name={p.id} size={p.size} />
          <span class="tip" aria-hidden="true">{p.label} <span class="key">{i + 1}</span></span>
        </button>
      {/each}
    </nav>
    <button
      class="page end"
      class:armed={ending}
      onclick={endClick}
      onblur={() => (ending = false)}
      aria-label={ending ? `${word('end')}: click again to confirm` : word('end')}
    >
      <Icon name="end" />
      <span class="tip" aria-hidden="true">{ending ? `click again to ${word('end')}` : word('end')}</span>
    </button>
  </aside>

  <header>
    <span class="brand">datavault</span>
    <span class="crumb">/ {settings ? 'settings' : current.label}</span>
    <span class="spacer"></span>
    <span class="dim">session {info.id}</span>
    <button onclick={() => (settings = !settings)} class:on={settings}>settings</button>
    <button onclick={onlock}>lock</button>
  </header>

  <main>
    <div class="page-view" hidden={settings || page !== 'grimoire'}>
      <Grimoire {info} active={!settings && page === 'grimoire'} />
    </div>
    <div class="page-view" hidden={settings || page !== 'unspoken'}>
      <Unspoken active={!settings && page === 'unspoken'} />
    </div>
    <div class="page-view" hidden={settings || page !== 'siege'}>
      <Siege active={!settings && page === 'siege'} />
    </div>
    <div class="page-view" hidden={settings || page !== 'crypt'}>
      <Crypt active={!settings && page === 'crypt'} />
    </div>
    <div class="page-view" hidden={settings || page !== 'ledger'}>
      <Ledger active={!settings && page === 'ledger'} />
    </div>
    {#if settings}
      <div class="page-view scroll">
        <div class="padded">
          <Settings onsaved={onrefresh} onclose={() => (settings = false)} />
        </div>
      </div>
    {/if}
  </main>

  <footer>
    <span>auto-lock {info.auto_lock_minutes}m · clipboard {info.clipboard_clear_seconds}s</span>
    <span class="spacer"></span>
    {#if !settings && page === 'grimoire'}<span>n new · g {word('forge')} · / search · esc close ·</span>{/if}
    {#if !settings && page === 'unspoken'}<span>n new · esc close · saves as you write ·</span>{/if}
    {#if !settings && page === 'siege'}<span>n new · arrows select · shift+arrows move · [ ] days ·</span>{/if}
    {#if !settings && page === 'crypt'}<span>n {word('keep')} · drop files here · esc close ·</span>{/if}
    {#if !settings && page === 'ledger'}<span>n new · esc close ·</span>{/if}
    <span>1–{pages.length} pages · ctrl+shift+l rest</span>
    <span class="version">{showVersion(info.version)}</span>
  </footer>
</div>

<style>
  .shell {
    height: 100vh;
    display: grid;
    grid-template-columns: 64px 1fr;
    grid-template-rows: auto 1fr auto;
    grid-template-areas:
      'side header'
      'side main'
      'side footer';
  }

  aside {
    position: relative;
    grid-area: side;
    /* An inset shadow instead of a border keeps the full 64px for the icons,
       so the pixel art lands on whole pixels. */
    box-shadow: inset -1px 0 0 var(--line);
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  aside nav {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }

  .page {
    position: relative;
    display: grid;
    place-items: center;
    width: 64px;
    height: 48px;
    color: var(--dim);
  }

  .page:hover,
  .page:focus-visible {
    color: var(--muted);
  }

  .page.active {
    color: var(--fg);
  }

  /* A lit bar on the sidebar edge marks the open page, like a cursor. */
  .page.active::before {
    content: '';
    position: absolute;
    left: 0;
    top: 12px;
    bottom: 12px;
    width: 2px;
    background: var(--fg);
  }


  /* The tombstone sits apart at the foot of the sidebar; the pages stay centred. */
  .end {
    position: absolute;
    bottom: 16px;
  }

  .end.armed {
    color: var(--accent);
  }

  .end.armed .tip {
    opacity: 1;
    color: var(--accent);
  }

  .tip {
    position: absolute;
    left: calc(100% + 4px);
    top: 50%;
    transform: translateY(-50%);
    padding: 2px 8px;
    background: var(--raise);
    border: 1px solid var(--line);
    color: var(--muted);
    font-size: 11px;
    white-space: nowrap;
    pointer-events: none;
    opacity: 0;
    z-index: 1;
  }

  .tip .key {
    color: var(--dim);
  }

  .page:hover .tip,
  .page:focus-visible .tip {
    opacity: 1;
  }

  header,
  footer {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 24px;
  }

  header {
    grid-area: header;
    border-bottom: 1px solid var(--line);
  }

  .brand {
    font-weight: 700;
  }

  .crumb {
    color: var(--muted);
    margin-left: -8px;
  }

  .spacer {
    flex: 1;
  }

  .dim {
    color: var(--dim);
  }

  header .on {
    color: var(--fg);
  }

  main {
    grid-area: main;
    min-height: 0;
    overflow: hidden;
  }

  .page-view {
    height: 100%;
  }

  .page-view[hidden] {
    display: none;
  }

  /* The whole pane scrolls, so a scroll bar sits at the window's edge; the
     settings themselves keep a readable width inside it. */
  .scroll {
    overflow-y: auto;
  }

  .padded {
    padding: 16px 32px;
    max-width: 880px;
  }

  footer {
    grid-area: footer;
    border-top: 1px solid var(--line);
    color: var(--dim);
    font-size: 11px;
    gap: 6px;
  }

  .version {
    margin-left: 12px;
    color: var(--muted);
  }

  @media (max-width: 720px) {
    .shell {
      grid-template-columns: 48px 1fr;
    }
    .page {
      width: 48px;
    }
    .tip {
      display: none;
    }
  }
</style>
