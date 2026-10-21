<script lang="ts">
  import { api, ApiError, type Item, type ItemInput, type SessionInfo, type Summary } from './api'
  import { copy } from './clipboard'
  import Editor from './Editor.svelte'
  import { isMultiline, typeShort } from './items'

  let { info, onlock }: { info: SessionInfo; onlock: () => void } = $props()

  let query = $state('')
  let items = $state<Summary[]>([])
  let selected = $state<Item | null>(null)
  let editing = $state<Item | 'new' | null>(null)
  let revealed = $state<Record<number, boolean>>({})
  let copied = $state<number | null>(null)
  let confirmDelete = $state(false)
  let error = $state('')
  let search = $state<HTMLInputElement>()

  function fail(e: unknown) {
    // 423 means the vault locked; App already switches back to the login.
    if (e instanceof ApiError && e.status === 423) return
    error = e instanceof Error ? e.message : String(e)
  }

  async function load() {
    items = await api.list(query)
  }

  $effect(() => {
    void query
    const t = setTimeout(() => load().catch(fail), 120)
    return () => clearTimeout(t)
  })

  async function select(id: string) {
    error = ''
    try {
      selected = await api.get(id)
      revealed = {}
      copied = null
      confirmDelete = false
      editing = null
    } catch (e) {
      fail(e)
    }
  }

  async function save(input: ItemInput) {
    const saved = editing === 'new' || editing === null ? await api.create(input) : await api.update(editing.id, input)
    editing = null
    selected = saved
    revealed = {}
    await load()
  }

  async function remove() {
    if (!selected) return
    if (!confirmDelete) {
      confirmDelete = true
      return
    }
    try {
      await api.remove(selected.id)
      selected = null
      confirmDelete = false
      await load()
    } catch (e) {
      fail(e)
    }
  }

  async function copyField(i: number) {
    if (!selected) return
    const field = selected.fields[i]
    try {
      await copy(field.value, info.clipboard_clear_seconds)
    } catch {
      error = 'clipboard not available'
      return
    }
    copied = i
    setTimeout(() => {
      if (copied === i) copied = null
    }, 1500)
    api.copied(selected.name, field.name).catch(() => {})
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      if (editing) editing = null
      else if (isTyping(e.target)) (e.target as HTMLElement).blur()
      else selected = null
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === '/') {
      e.preventDefault()
      search?.focus()
    } else if (e.key === 'n') {
      e.preventDefault()
      editing = 'new'
    }
  }

  const masked = '••••••••••••'
</script>

<svelte:window {onkeydown} />

<div class="shell">
  <header>
    <span class="brand">datavault</span>
    <span class="spacer"></span>
    <span class="dim">session {info.id}</span>
    <button onclick={onlock}>lock</button>
  </header>

  <div class="toolbar">
    <span class="dim" aria-hidden="true">/</span>
    <input type="search" bind:this={search} bind:value={query} placeholder="search" aria-label="search" spellcheck="false" />
    <button onclick={() => (editing = 'new')}>+ new</button>
  </div>

  <div class="body">
    <nav aria-label="items">
      {#each items as it (it.id)}
        <button class="item" class:active={selected?.id === it.id} onclick={() => select(it.id)}>
          <span class="type">{typeShort(it.type)}</span>
          <span class="name">{it.name}</span>
          <span class="tags">{it.tags.join(' ')}</span>
        </button>
      {:else}
        <p class="empty">{query ? 'no matches' : 'empty vault · press n to add the first item'}</p>
      {/each}
    </nav>

    <section>
      {#if error}<p class="error" role="alert">{error}</p>{/if}

      {#if editing}
        {#key editing}
          <Editor item={editing === 'new' ? null : editing} onsave={save} oncancel={() => (editing = null)} />
        {/key}
      {:else if selected}
        <div class="detail">
          <h2>
            {selected.name}
            <span class="dim">{typeShort(selected.type)}</span>
          </h2>

          {#each selected.fields as field, i (i)}
            {#if field.value}
            <div class="field">
              <span class="label">{field.name}</span>
              {#if field.secret && !revealed[i]}
                <span class="value masked">{masked}</span>
              {:else if isMultiline(field)}
                <pre class="value">{field.value}</pre>
              {:else}
                <span class="value">{field.value}</span>
              {/if}
              <span class="actions">
                {#if field.secret}
                  <button onclick={() => (revealed[i] = !revealed[i])}>{revealed[i] ? 'hide' : 'show'}</button>
                {/if}
                <button onclick={() => copyField(i)} class:done={copied === i}>{copied === i ? 'copied' : 'copy'}</button>
              </span>
            </div>
            {/if}
          {/each}

          {#if selected.tags.length}
            <div class="field">
              <span class="label">tags</span>
              <span class="value">{selected.tags.join(', ')}</span>
            </div>
          {/if}
          {#if selected.notes}
            <div class="field">
              <span class="label">notes</span>
              <pre class="value">{selected.notes}</pre>
            </div>
          {/if}

          <p class="meta">updated {new Date(selected.updated_at).toLocaleString()}</p>

          <div class="buttons">
            <button onclick={() => (editing = selected)}>edit</button>
            <button onclick={remove} class:danger={confirmDelete}>{confirmDelete ? 'delete? click again' : 'delete'}</button>
          </div>
        </div>
      {:else}
        <p class="empty">select an item</p>
      {/if}
    </section>
  </div>

  <footer>
    <span>{items.length} {items.length === 1 ? 'item' : 'items'}</span>
    <span>auto-lock {info.auto_lock_minutes}m · clipboard {info.clipboard_clear_seconds}s</span>
    <span class="keys">n new · / search · esc close</span>
  </footer>
</div>

<style>
  .shell {
    min-height: 100vh;
    display: grid;
    grid-template-rows: auto auto 1fr auto;
  }

  header,
  .toolbar,
  footer {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 24px;
  }

  header {
    border-bottom: 1px solid var(--line);
  }

  .brand {
    font-weight: 700;
  }

  .spacer {
    flex: 1;
  }

  .dim {
    color: var(--dim);
  }

  .toolbar {
    gap: 10px;
    padding-top: 16px;
  }

  .toolbar input {
    max-width: 360px;
    border-bottom-color: transparent;
  }

  .toolbar input:focus {
    border-bottom-color: var(--line);
  }

  .body {
    display: grid;
    grid-template-columns: minmax(240px, 360px) 1fr;
    min-height: 0;
  }

  nav {
    border-right: 1px solid var(--line);
    padding: 8px 0;
    overflow-y: auto;
  }

  .item {
    display: grid;
    grid-template-columns: 44px 1fr auto;
    gap: 10px;
    width: 100%;
    padding: 5px 24px;
    text-align: left;
    color: var(--fg);
  }

  .item:hover {
    background: var(--raise);
  }

  .item.active {
    background: var(--raise);
    box-shadow: inset 2px 0 0 var(--fg);
  }

  .type,
  .tags {
    color: var(--dim);
  }

  .name,
  .tags {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tags {
    font-size: 11px;
  }

  section {
    padding: 16px 32px;
    overflow-y: auto;
    max-width: 820px;
  }

  .empty {
    color: var(--dim);
    padding: 0 24px;
  }

  section .empty {
    padding: 0;
  }

  h2 {
    margin: 0 0 20px;
    font-size: 15px;
    font-weight: 700;
    display: flex;
    gap: 12px;
    align-items: baseline;
  }

  h2 .dim {
    font-size: 12px;
    font-weight: 400;
  }

  .field {
    display: grid;
    grid-template-columns: 110px 1fr auto;
    gap: 12px;
    align-items: baseline;
    padding: 6px 0;
    border-bottom: 1px solid var(--line);
  }

  .label {
    color: var(--dim);
  }

  .value {
    overflow-wrap: anywhere;
    user-select: text;
  }

  pre.value {
    margin: 0;
    white-space: pre-wrap;
    font: inherit;
  }

  .masked {
    color: var(--muted);
    letter-spacing: 0.1em;
  }

  .actions {
    display: flex;
    gap: 12px;
    font-size: 12px;
  }

  .done,
  .danger,
  .error {
    color: var(--accent);
  }

  .meta {
    color: var(--dim);
    font-size: 11px;
    margin: 16px 0 0;
  }

  .buttons {
    display: flex;
    gap: 20px;
    margin-top: 20px;
  }

  footer {
    border-top: 1px solid var(--line);
    color: var(--dim);
    font-size: 11px;
  }

  .keys {
    margin-left: auto;
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
    }
    nav {
      border-right: 0;
      border-bottom: 1px solid var(--line);
      max-height: 40vh;
    }
    section {
      padding: 16px;
    }
    .keys {
      display: none;
    }
  }
</style>
