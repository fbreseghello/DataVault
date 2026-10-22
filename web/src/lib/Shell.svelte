<script lang="ts">
  import type { SessionInfo } from './api'
  import Grimoire from './Grimoire.svelte'
  import Icon from './Icon.svelte'
  import type { IconName } from './icons'
  import Settings from './Settings.svelte'
  import Siege from './Siege.svelte'
  import Unspoken from './Unspoken.svelte'
  import { showVersion } from './version'

  let { info, onlock, onrefresh }: { info: SessionInfo; onlock: () => void; onrefresh: () => void } = $props()

  type Page = { id: string; label: string; icon: IconName }

  // The sidebar shows icons only; label is for the hover tip and screen readers.
  const pages: Page[] = [
    { id: 'grimoire', label: 'grimoire', icon: 'book' },
    { id: 'heart', label: 'unspoken', icon: 'heart' },
    { id: 'sword', label: 'siege', icon: 'sword' },
  ]

  let page = $state('grimoire')
  let settings = $state(false)
  const current = $derived(pages.find((p) => p.id === page)!)

  function go(id: string) {
    page = id
    settings = false
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
          <Icon name={p.icon} />
          <span class="tip" aria-hidden="true">{p.label} <span class="key">{i + 1}</span></span>
        </button>
      {/each}
    </nav>
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
    <div class="page-view" hidden={settings || page !== 'heart'}>
      <Unspoken active={!settings && page === 'heart'} />
    </div>
    <div class="page-view" hidden={settings || page !== 'sword'}>
      <Siege active={!settings && page === 'sword'} />
    </div>
    {#if settings}
      <div class="page-view padded">
        <Settings onsaved={onrefresh} onclose={() => (settings = false)} />
      </div>
    {/if}
  </main>

  <footer>
    <span>auto-lock {info.auto_lock_minutes}m · clipboard {info.clipboard_clear_seconds}s</span>
    <span class="spacer"></span>
    {#if !settings && page === 'grimoire'}<span>n new · / search · esc close ·</span>{/if}
    {#if !settings && page === 'heart'}<span>n new · esc close · saves as you write ·</span>{/if}
    {#if !settings && page === 'sword'}<span>n order · arrows select · shift+arrows move · [ ] days ·</span>{/if}
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

  .padded {
    padding: 16px 32px;
    overflow-y: auto;
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
