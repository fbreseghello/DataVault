<script lang="ts">
  import { onMount, tick, untrack } from 'svelte'
  import { api, ApiError, type Note, type NoteInput, type NoteSummary } from './api'
  import Icon from './Icon.svelte'
  import Placard from './Placard.svelte'
  import SealGate from './SealGate.svelte'

  // Unspoken: notes for what is not said out loud. Notes save themselves while
  // you write; a note can be sealed, or buried in Dead Words. Like the grimoire,
  // the page stays mounted but returns to its front page each time it opens.
  let { active }: { active: boolean } = $props()

  const SAVE_AFTER = 800

  let view = $state<'live' | 'dead'>('live')
  let list = $state<NoteSummary[]>([])
  let selectedId = $state<string | null>(null)
  let opened = $state<Note | null>(null)
  let token = $state('')
  let title = $state('')
  let body = $state('')
  let dirty = $state(false)
  let status = $state<'' | 'saving' | 'saved'>('')
  let hasSealPin = $state(false)
  let sealPrompt = $state(false)
  let sealPin = $state('')
  let sealPinConfirm = $state('')
  let confirmDelete = $state(false)
  let error = $state('')
  let bodyField = $state<HTMLTextAreaElement>()
  let saveTimer: ReturnType<typeof setTimeout> | undefined

  const selected = $derived(list.find((n) => n.id === selectedId) ?? null)

  function fail(e: unknown) {
    if (e instanceof ApiError && e.status === 423) return
    error = e instanceof Error ? e.message : String(e)
  }

  async function load() {
    list = await api.notes(view === 'dead')
  }

  onMount(() => {
    api.sealStatus().then((r) => (hasSealPin = r.has_pin), fail)
    // A save still pending when the tab closes goes out with keepalive.
    const onPageHide = () => {
      if (opened && dirty) void api.updateNote(opened.id, input(), token, true).catch(() => {})
    }
    window.addEventListener('pagehide', onPageHide)
    return () => {
      window.removeEventListener('pagehide', onPageHide)
      onPageHide()
    }
  })

  // Entering Unspoken starts at its front page; leaving it saves and seals.
  $effect(() => {
    const on = active
    untrack(() => {
      if (on) {
        view = 'live'
        void close().then(() => load().catch(fail))
      } else {
        void close()
      }
    })
  })

  function input(): NoteInput {
    return { title, body, sealed: opened?.sealed ?? false }
  }

  function show(n: Note, t: string) {
    opened = n
    token = t
    title = n.title
    body = n.body
    dirty = false
    status = ''
    confirmDelete = false
  }

  function edited() {
    dirty = true
    status = ''
    clearTimeout(saveTimer)
    saveTimer = setTimeout(() => void save(), SAVE_AFTER)
  }

  async function save() {
    clearTimeout(saveTimer)
    if (!opened || !dirty) return
    const id = opened.id
    dirty = false
    status = 'saving'
    try {
      const res = await api.updateNote(id, input(), token)
      if (opened?.id === id) {
        opened = res.note
        token = res.token
        status = 'saved'
      }
      await load()
    } catch (e) {
      dirty = true
      status = ''
      fail(e)
    }
  }

  // close saves the open note and lets it go, which also seals it again.
  // A note left with nothing written in it is thrown away.
  async function close() {
    await save()
    const n = opened
    if (n && !n.buried && !title.trim() && !body.trim()) {
      await api.deleteNote(n.id, token).catch(() => {})
      await load().catch(() => {})
    }
    opened = null
    token = ''
    selectedId = null
    title = body = ''
    status = ''
    sealPrompt = false
    sealPin = sealPinConfirm = ''
    confirmDelete = false
    error = ''
  }

  async function select(n: NoteSummary) {
    await close()
    selectedId = n.id
    if (n.sealed) return // the seal gate asks for the PIN
    try {
      show(await api.note(n.id), '')
    } catch (e) {
      fail(e)
    }
  }

  async function unseal(pin: string) {
    if (!selectedId) return
    const res = await api.unsealNote(selectedId, pin)
    show(res.note, res.token)
  }

  async function write() {
    await close()
    view = 'live'
    try {
      const res = await api.createNote({ title: '', body: '', sealed: false })
      selectedId = res.note.id
      show(res.note, res.token)
      await load()
      await tick()
      bodyField?.focus()
    } catch (e) {
      fail(e)
    }
  }

  async function switchView(v: 'live' | 'dead') {
    if (v === view) return
    await close()
    view = v
    await load().catch(fail)
  }

  // The padlock: seal an open note (making the seal PIN first if needed), or
  // close a sealed one so it asks for the PIN again.
  async function seal(e?: SubmitEvent) {
    e?.preventDefault()
    if (!opened) return
    error = ''
    if (opened.sealed) {
      const id = opened.id
      await save()
      opened = null
      token = ''
      selectedId = id
      return
    }
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
      if (!hasSealPin) {
        await api.setSealPin('', sealPin)
        hasSealPin = true
      }
      const id = opened.id
      await api.updateNote(id, { title, body, sealed: true }, token)
      dirty = false
      opened = null
      token = ''
      sealPrompt = false
      selectedId = id
      await load()
    } catch (err) {
      fail(err)
    }
  }

  async function bury() {
    if (!opened) return
    await save()
    try {
      await api.buryNote(opened.id, view === 'live', token)
      opened = null
      await close()
      await load()
    } catch (e) {
      fail(e)
    }
  }

  async function remove() {
    if (!opened) return
    if (!confirmDelete) {
      confirmDelete = true
      return
    }
    try {
      await api.deleteNote(opened.id, token)
      opened = null
      await close()
      await load()
    } catch (e) {
      fail(e)
    }
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active) return
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
      e.preventDefault()
      void save()
      return
    }
    if (e.key === 'Escape') {
      if (isTyping(e.target) && !selected?.sealed) (e.target as HTMLElement).blur()
      else void close()
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === 'n') {
      e.preventDefault()
      void write()
    }
  }

  function when(iso: string) {
    const d = new Date(iso)
    const today = new Date().toDateString() === d.toDateString()
    return today
      ? d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : d.toLocaleDateString([], { day: '2-digit', month: 'short', year: 'numeric' })
  }
</script>

<svelte:window {onkeydown} />

<div class="unspoken">
  <div class="toolbar">
    <button class="tab" class:on={view === 'live'} onclick={() => switchView('live')}>unspoken</button>
    <button class="tab" class:on={view === 'dead'} onclick={() => switchView('dead')}>dead words</button>
    <button onclick={write}>+ new</button>
    <span class="spacer"></span>
    <span class="dim count">{list.length} {list.length === 1 ? 'note' : 'notes'}</span>
  </div>

  <div class="body">
    <nav aria-label="notes">
      {#each list as n (n.id)}
        <button class="note" class:active={selectedId === n.id} onclick={() => select(n)}>
          <span class="line">
            {#if n.sealed}
              <span class="sealed">sealed</span>
              <span class="mark" title="sealed"><Icon name="seal" size={16} /></span>
            {:else}
              <span class="name">{n.title || n.preview || 'untitled'}</span>
            {/if}
          </span>
          <span class="date">{when(n.updated_at)}</span>
        </button>
      {:else}
        <p class="empty">{view === 'dead' ? 'no dead words yet' : 'nothing written · press n'}</p>
      {/each}
    </nav>

    <section>
      {#if error}<p class="error" role="alert">{error}</p>{/if}

      {#if opened}
        <div class="editor">
          <div class="head">
            <input
              class="title"
              bind:value={title}
              oninput={edited}
              placeholder="untitled"
              spellcheck="false"
              autocomplete="off"
              aria-label="title"
            />
            <button
              class="seal"
              class:on={opened.sealed}
              onclick={() => seal()}
              title={opened.sealed ? 'seal again' : 'seal this note'}
              aria-label={opened.sealed ? 'seal again' : 'seal this note'}
            >
              <Icon name="seal" size={32} />
            </button>
          </div>

          {#if sealPrompt}
            <form class="sealpin" onsubmit={seal}>
              <p class="dim">no seal pin yet. choose one; it opens every sealed note and entry.</p>
              <input type="password" bind:value={sealPin} minlength="4" required placeholder="seal pin" autocomplete="new-password" />
              <input type="password" bind:value={sealPinConfirm} required placeholder="confirm pin" autocomplete="new-password" />
              <button type="submit" class="primary">seal</button>
              <button type="button" onclick={() => (sealPrompt = false)}>cancel</button>
            </form>
          {/if}

          <textarea
            bind:this={bodyField}
            bind:value={body}
            oninput={edited}
            placeholder="dead words go here."
            spellcheck="false"
            aria-label="note"
          ></textarea>

          <div class="foot">
            <span class="dim">
              written {when(opened.created_at)}{status === 'saving' ? ' · saving' : status === 'saved' ? ' · saved' : dirty ? ' · …' : ''}
            </span>
            <span class="spacer"></span>
            <button onclick={bury}>{view === 'live' ? 'bury' : 'unbury'}</button>
            {#if view === 'dead'}
              <button onclick={remove} class:danger={confirmDelete}>{confirmDelete ? 'delete forever? click again' : 'delete'}</button>
            {/if}
          </div>
        </div>
      {:else if selected?.sealed}
        <SealGate name="sealed note" onunseal={unseal} />
      {:else if view === 'dead'}
        <Placard icon="heart" title="dead words" hint={list.length ? 'buried notes rest here' : 'no dead words yet'} />
      {:else}
        <Placard icon="heart" title="unspoken" hint="say it here. no one else will read it." />
      {/if}
    </section>
  </div>
</div>

<style>
  .unspoken {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 24px;
  }

  .tab.on {
    color: var(--fg);
  }

  .spacer {
    flex: 1;
  }

  .dim {
    color: var(--dim);
  }

  .count {
    font-size: 11px;
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

  .note {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 10px;
    width: 100%;
    padding: 5px 24px;
    text-align: left;
    color: var(--fg);
  }

  .note:hover,
  .note.active {
    background: var(--raise);
  }

  .note.active {
    box-shadow: inset 2px 0 0 var(--fg);
  }

  .line,
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sealed,
  .mark,
  .date {
    color: var(--dim);
  }

  .mark {
    display: inline-block;
    vertical-align: -3px;
    margin-left: 6px;
  }

  .date {
    font-size: 11px;
  }

  .empty {
    color: var(--dim);
    padding: 0 24px;
  }

  section {
    padding: 16px 32px;
    min-height: 0;
    overflow: hidden;
  }

  .editor {
    height: 100%;
    max-width: 820px;
    display: grid;
    grid-template-rows: auto auto 1fr auto;
    gap: 8px;
  }

  .head {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .title {
    flex: 1;
    font-size: 15px;
    font-weight: 700;
    border-bottom-color: transparent;
  }

  .title:focus {
    border-bottom-color: var(--line);
  }

  textarea {
    min-height: 0;
    height: 100%;
    resize: none;
    border: 0;
    border-left: 1px solid var(--line);
    padding: 4px 0 4px 14px;
    line-height: 1.7;
    font-size: 13px;
  }

  textarea:focus {
    border-left-color: var(--muted);
  }

  textarea::placeholder,
  .title::placeholder {
    color: var(--dim);
    font-style: italic;
  }

  .seal {
    color: var(--dim);
  }

  .seal.on {
    color: var(--accent);
  }

  .seal:hover {
    color: var(--fg);
  }

  .sealpin {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 12px;
  }

  .sealpin p {
    flex-basis: 100%;
    margin: 0;
  }

  .sealpin input {
    width: 160px;
  }

  .foot {
    display: flex;
    align-items: baseline;
    gap: 20px;
    font-size: 12px;
  }

  .primary {
    color: var(--fg);
  }

  .danger,
  .error {
    color: var(--accent);
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
    }
    nav {
      border-right: 0;
      border-bottom: 1px solid var(--line);
      max-height: 35vh;
    }
    section {
      padding: 16px;
    }
  }
</style>
