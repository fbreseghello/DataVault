<script lang="ts">
  import { onMount, untrack } from 'svelte'
  import { api, ApiError, type Item, type ItemInput, type SessionInfo, type Summary } from './api'
  import { copy } from './clipboard'
  import Editor from './Editor.svelte'
  import Icon from './Icon.svelte'
  import { isMultiline, typeShort } from './items'
  import Placard from './Placard.svelte'
  import SealGate from './SealGate.svelte'

  // active is false while another page is shown; the grimoire stays mounted but
  // returns to its front page every time it becomes active again.
  let { info, active }: { info: SessionInfo; active: boolean } = $props()

  let query = $state('')
  let items = $state<Summary[]>([])
  let selectedId = $state<string | null>(null)
  // opened holds the fields of the selected entry. For a sealed entry it stays
  // null until the seal PIN is given, and token keeps it open on the server.
  let opened = $state<Item | null>(null)
  let token = $state('')
  let editing = $state<Item | 'new' | null>(null)
  let hasSealPin = $state(false)
  let sealPrompt = $state(false)
  let sealPin = $state('')
  let sealPinConfirm = $state('')
  let revealed = $state<Record<number, boolean>>({})
  let copied = $state<number | null>(null)
  let confirmDelete = $state(false)
  let error = $state('')
  let search = $state<HTMLInputElement>()

  const selected = $derived(items.find((it) => it.id === selectedId) ?? null)

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

  onMount(() => {
    api.sealStatus().then((r) => (hasSealPin = r.has_pin), fail)
  })

  // Entering the grimoire never resumes an open entry or an unsaved draft, and
  // reloads the list in case it changed meanwhile.
  $effect(() => {
    if (!active) return
    untrack(() => {
      close()
      load().catch(fail)
    })
  })

  // close drops the open entry, which also seals it again.
  function close() {
    selectedId = null
    editing = null
    reseal()
    error = ''
  }

  function reseal() {
    opened = null
    token = ''
    revealed = {}
    copied = null
    confirmDelete = false
    sealPrompt = false
    sealPin = sealPinConfirm = ''
  }

  async function select(it: Summary) {
    close()
    selectedId = it.id
    if (it.sealed) return // the seal gate asks for the PIN
    try {
      opened = await api.get(it.id)
    } catch (e) {
      fail(e)
    }
  }

  async function unseal(pin: string) {
    if (!selectedId) return
    const res = await api.unseal(selectedId, pin)
    opened = res.item
    token = res.token
  }

  function inputOf(it: Item): ItemInput {
    const { id: _id, created_at: _c, updated_at: _u, ...rest } = it
    return rest
  }

  async function ensureSealPin(pin: string) {
    if (hasSealPin || !pin) return
    await api.setSealPin('', pin)
    hasSealPin = true
  }

  async function save(input: ItemInput, newSealPin: string) {
    await ensureSealPin(newSealPin)
    const res =
      editing === 'new' || editing === null ? await api.create(input) : await api.update(editing.id, input, token)
    editing = null
    selectedId = res.item.id
    opened = res.item
    token = res.token
    revealed = {}
    await load()
  }

  // sealNow seals an entry that was open without a PIN, creating the seal PIN
  // first when there is none, and leaves it sealed.
  async function sealNow(e?: SubmitEvent) {
    e?.preventDefault()
    if (!opened) return
    error = ''
    if (!hasSealPin) {
      if (!sealPrompt) {
        sealPrompt = true
        return
      }
      if (sealPin !== sealPinConfirm) {
        error = "the pins don't match"
        return
      }
    }
    try {
      await ensureSealPin(sealPin)
      await api.update(opened.id, { ...inputOf(opened), sealed: true }, token)
      reseal()
      await load()
    } catch (err) {
      fail(err)
    }
  }

  async function remove() {
    if (!opened) return
    if (!confirmDelete) {
      confirmDelete = true
      return
    }
    try {
      await api.remove(opened.id, token)
      close()
      await load()
    } catch (e) {
      fail(e)
    }
  }

  async function copyField(i: number) {
    if (!opened) return
    const field = opened.fields[i]
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
    api.copied(opened.name, field.name).catch(() => {})
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active) return
    if (e.key === 'Escape') {
      if (editing) editing = null
      else if (isTyping(e.target) && !selected?.sealed) (e.target as HTMLElement).blur()
      else close()
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === '/') {
      e.preventDefault()
      search?.focus()
    } else if (e.key === 'n') {
      e.preventDefault()
      close()
      editing = 'new'
    }
  }

  const masked = '••••••••••••'
</script>

<svelte:window {onkeydown} />

<div class="grimoire">
  <div class="toolbar">
    <span class="dim" aria-hidden="true">/</span>
    <input type="search" bind:this={search} bind:value={query} placeholder="search" aria-label="search" spellcheck="false" />
    <button onclick={() => (close(), (editing = 'new'))}>+ new</button>
    <span class="spacer"></span>
    <span class="dim count">{items.length} {items.length === 1 ? 'entry' : 'entries'}</span>
  </div>

  <div class="body">
    <nav aria-label="entries">
      {#each items as it (it.id)}
        <button class="item" class:active={selectedId === it.id} onclick={() => select(it)}>
          <span class="type">{typeShort(it.type)}</span>
          <span class="name">
            {it.name}
            {#if it.sealed}<span class="mark" title="sealed"><Icon name="seal" size={16} /></span>{/if}
          </span>
          <span class="tags">{it.tags.join(' ')}</span>
        </button>
      {:else}
        <p class="empty">{query ? 'no matches' : 'empty · press n to write the first entry'}</p>
      {/each}
    </nav>

    <section>
      {#if error}<p class="error" role="alert">{error}</p>{/if}

      {#if editing}
        <div class="page">
          {#key editing}
            <Editor item={editing === 'new' ? null : editing} {hasSealPin} onsave={save} oncancel={() => (editing = null)} />
          {/key}
        </div>
      {:else if opened}
        <div class="page">
          <h2>
            {opened.name}
            <span class="dim">{typeShort(opened.type)}</span>
            <span class="spacer"></span>
            <button
              class="seal"
              class:on={opened.sealed}
              onclick={() => (opened?.sealed ? reseal() : sealNow())}
              title={opened.sealed ? 'seal again' : 'seal this entry'}
              aria-label={opened.sealed ? 'seal again' : 'seal this entry'}
            >
              <Icon name="seal" size={32} />
            </button>
          </h2>

          {#if sealPrompt}
            <form class="sealpin fields" onsubmit={sealNow}>
              <p class="dim">no seal pin yet. choose one; it opens every sealed entry.</p>
              <label class="field">
                <span class="label">seal pin</span>
                <input type="password" bind:value={sealPin} minlength="4" required autocomplete="new-password" />
                <span></span>
              </label>
              <label class="field">
                <span class="label">confirm pin</span>
                <input type="password" bind:value={sealPinConfirm} required autocomplete="new-password" />
                <span></span>
              </label>
              <div class="buttons">
                <button type="submit" class="primary">seal</button>
                <button type="button" onclick={() => (sealPrompt = false)}>cancel</button>
              </div>
            </form>
          {/if}

          <div class="fields">
          {#each opened.fields as field, i (i)}
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

          {#if opened.tags.length}
            <div class="field">
              <span class="label">tags</span>
              <span class="value">{opened.tags.join(', ')}</span>
            </div>
          {/if}
          {#if opened.notes}
            <div class="field">
              <span class="label">notes</span>
              <pre class="value">{opened.notes}</pre>
            </div>
          {/if}

          </div>

          <p class="meta">updated {new Date(opened.updated_at).toLocaleString()}</p>

          <div class="buttons">
            <button onclick={() => (editing = opened)}>edit</button>
            <button onclick={remove} class:danger={confirmDelete}>{confirmDelete ? 'delete? click again' : 'delete'}</button>
          </div>
        </div>
      {:else if selected?.sealed}
        <SealGate name={selected.name} onunseal={unseal} />
      {:else if !selected}
        <Placard icon="book" title="grimoire" hint="select an entry · n to write a new one" />
      {/if}
    </section>
  </div>
</div>

<style>
  .grimoire {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  .toolbar {
    display: flex;
    align-items: center;
    padding: 12px 24px;
  }

  .spacer {
    flex: 1;
  }

  .count {
    font-size: 11px;
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
    min-height: 0;
  }

  .page {
    max-width: 820px;
  }

  .empty {
    color: var(--dim);
    padding: 0 24px;
  }

  h2 {
    margin: 0 0 20px;
    font-size: 15px;
    font-weight: 700;
    display: flex;
    gap: 12px;
    align-items: center;
  }

  /* The padlock in an open entry: dim when the entry opens freely, lit when it
     is sealed and was opened with the PIN. Clicking seals it (again). */
  .seal {
    color: var(--dim);
  }

  .seal.on {
    color: var(--accent);
  }

  .seal:hover {
    color: var(--fg);
  }

  .mark {
    display: inline-block;
    vertical-align: -3px;
    margin-left: 6px;
    color: var(--dim);
  }

  .sealpin {
    margin: 0 0 20px;
  }

  .sealpin input {
    width: 100%;
  }

  .primary {
    color: var(--fg);
  }

  h2 .dim {
    font-size: 12px;
    font-weight: 400;
  }

  /* One grid for all rows, so every value starts at the same column. The name
     column grows to the longest field name; very long names wrap. */
  .fields {
    display: grid;
    grid-template-columns: minmax(110px, max-content) minmax(0, 1fr) auto;
    column-gap: 16px;
  }

  .field {
    grid-column: 1 / -1;
    display: grid;
    grid-template-columns: subgrid;
    align-items: baseline;
    padding: 6px 0;
    border-bottom: 1px solid var(--line);
  }

  .label {
    color: var(--dim);
    max-width: 28ch;
    overflow-wrap: anywhere;
  }

  .sealpin p,
  .sealpin .buttons {
    grid-column: 1 / -1;
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
  }
</style>
