<script lang="ts">
  import { untrack } from 'svelte'
  import { api, ApiError, MAX_FILE_BYTES, type CryptFile } from './api'
  import Placard from './Placard.svelte'

  // The crypt: documents and files buried in the vault, each encrypted on its
  // own. Files show as tiles; hovering one shows its title over it. Files come
  // in from "+ new", the n key, or by dropping them on the page.
  let { active }: { active: boolean } = $props()

  let files = $state<CryptFile[]>([])
  let selectedId = $state<string | null>(null)
  let renaming = $state(false)
  let title = $state('')
  let confirmDestroy = $state(false)
  let busy = $state('')
  let error = $state('')
  let dragging = $state(false)
  let picker = $state<HTMLInputElement>()

  const selected = $derived(files.find((f) => f.id === selectedId) ?? null)

  function fail(e: unknown) {
    if (e instanceof ApiError && e.status === 423) return
    error = e instanceof Error ? e.message : String(e)
  }

  async function load() {
    files = await api.files()
  }

  // Entering the crypt starts from its front page and reloads the list.
  $effect(() => {
    if (!active) return
    untrack(() => {
      select(null)
      load().catch(fail)
    })
  })

  function select(id: string | null) {
    selectedId = id
    renaming = false
    confirmDestroy = false
    error = ''
  }

  async function bury(list: FileList | File[]) {
    error = ''
    const all = [...list]
    const tooBig = all.filter((f) => f.size > MAX_FILE_BYTES)
    if (tooBig.length) error = `files can be up to 25 MB · ${tooBig.map((f) => f.name).join(', ')} left out`
    let last: CryptFile | null = null
    for (const f of all.filter((f) => f.size <= MAX_FILE_BYTES)) {
      busy = `burying ${f.name}`
      try {
        last = await api.addFile(f)
      } catch (e) {
        fail(e)
      }
    }
    busy = ''
    await load().catch(fail)
    if (last) select(last.id)
  }

  function pick() {
    picker?.click()
  }

  async function take() {
    if (!selected) return
    busy = 'taking it out'
    try {
      await api.takeFile(selected)
    } catch (e) {
      fail(e)
    } finally {
      busy = ''
    }
  }

  function startRename() {
    if (!selected) return
    title = selected.title
    renaming = true
  }

  async function rename(e: SubmitEvent) {
    e.preventDefault()
    if (!selected) return
    try {
      await api.renameFile(selected.id, title)
      renaming = false
      await load()
    } catch (err) {
      fail(err)
    }
  }

  async function destroy() {
    if (!selected) return
    if (!confirmDestroy) {
      confirmDestroy = true
      return
    }
    try {
      await api.removeFile(selected.id)
      select(null)
      await load()
    } catch (e) {
      fail(e)
    }
  }

  // what the tile shows: the file's extension, or its kind
  function kind(f: CryptFile): string {
    const dot = f.name.lastIndexOf('.')
    if (dot > 0 && f.name.length - dot <= 6) return f.name.slice(dot + 1).toLowerCase()
    return f.type.split('/')[0] || 'file'
  }

  function size(n: number): string {
    if (n < 1024) return `${n} B`
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(n < 10240 ? 1 : 0)} KB`
    return `${(n / 1024 / 1024).toFixed(1)} MB`
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active) return
    if (e.key === 'Escape') {
      if (renaming) renaming = false
      else select(null)
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === 'n') {
      e.preventDefault()
      pick()
    }
  }

  function ondragover(e: DragEvent) {
    if (!active || !e.dataTransfer?.types.includes('Files')) return
    e.preventDefault()
    dragging = true
  }

  function ondrop(e: DragEvent) {
    if (!active) return
    e.preventDefault()
    dragging = false
    if (e.dataTransfer?.files.length) bury(e.dataTransfer.files)
  }
</script>

<svelte:window {onkeydown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="crypt" class:dragging {ondragover} ondragleave={() => (dragging = false)} {ondrop}>
  <div class="toolbar page-toolbar">
    <span class="dim">{busy || (dragging ? 'drop to bury' : '')}</span>
    <span class="spacer"></span>
    <span class="count">{files.length} {files.length === 1 ? 'file' : 'files'}</span>
    <button class="new" onclick={pick}>+ new</button>
    <input
      bind:this={picker}
      type="file"
      multiple
      hidden
      onchange={(e) => {
        const input = e.currentTarget
        if (input.files?.length) bury(input.files).finally(() => (input.value = ''))
      }}
    />
  </div>

  <div class="body">
    <nav aria-label="files">
      {#each files as f (f.id)}
        <button class="tile" class:active={selectedId === f.id} onclick={() => select(f.id)} aria-label={f.title}>
          <span class="kind">{kind(f)}</span>
          <span class="size">{size(f.size)}</span>
          <!-- the title passes over the file when the pointer is on it -->
          <span class="title" aria-hidden="true">{f.title}</span>
        </button>
      {:else}
        <p class="empty">nothing buried yet · n or drop a file</p>
      {/each}
    </nav>

    <!-- position: relative anchors the front page icon (Placard) to this pane. -->
    <section>
      {#if error}<p class="error" role="alert">{error}</p>{/if}

      {#if selected}
        <div class="page">
          {#if renaming}
            <form class="rename" onsubmit={rename}>
              <!-- svelte-ignore a11y_autofocus -->
              <input type="text" bind:value={title} required autofocus spellcheck="false" autocomplete="off" aria-label="title" />
              <button type="submit" class="primary">save</button>
              <button type="button" onclick={() => (renaming = false)}>cancel</button>
            </form>
          {:else}
            <h2>{selected.title}</h2>
          {/if}

          <div class="fields">
            <div class="field"><span class="label">file</span><span class="value">{selected.name}</span></div>
            <div class="field"><span class="label">kind</span><span class="value">{selected.type || kind(selected)}</span></div>
            <div class="field"><span class="label">size</span><span class="value">{size(selected.size)}</span></div>
            <div class="field">
              <span class="label">buried</span>
              <span class="value">{new Date(selected.created_at).toLocaleString('en-GB')}</span>
            </div>
          </div>

          <div class="buttons">
            <button class="primary" onclick={take} disabled={!!busy}>take</button>
            <button onclick={startRename}>rename</button>
            <button onclick={destroy} class:danger={confirmDestroy}>{confirmDestroy ? 'destroy? click again' : 'destroy'}</button>
          </div>
          <p class="meta">kept encrypted · taking it out saves a copy to your downloads</p>
        </div>
      {:else}
        <Placard
          icon="cross"
          title="crypt"
          hint={files.length ? 'select a file · n or drop one to bury it' : 'documents and files, kept encrypted · n or drop one here'}
        />
      {/if}
    </section>
  </div>
</div>

<style>
  .crypt {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  .crypt.dragging section,
  .crypt.dragging nav {
    background: var(--raise);
  }

  .dim {
    color: var(--dim);
    font-size: 12px;
  }

  .body {
    display: grid;
    grid-template-columns: minmax(240px, 360px) 1fr;
    min-height: 0;
  }

  nav {
    border-right: 1px solid var(--line);
    padding: 12px;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(92px, 1fr));
    grid-auto-rows: 92px;
    gap: 10px;
    align-content: start;
  }

  .tile {
    position: relative;
    display: grid;
    place-content: center;
    gap: 4px;
    border: 1px solid var(--line);
    color: var(--muted);
    overflow: hidden;
  }

  .tile:hover,
  .tile:focus-visible {
    border-color: var(--dim);
  }

  .tile.active {
    border-color: var(--fg);
    color: var(--fg);
  }

  .kind {
    font-size: 15px;
    text-transform: lowercase;
  }

  .size {
    font-size: 11px;
    color: var(--dim);
  }

  /* the title, laid over the file while the pointer is on it */
  .title {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    padding: 6px;
    background: var(--bg);
    color: var(--fg);
    font-size: 12px;
    line-height: 1.3;
    text-align: center;
    overflow-wrap: anywhere;
    opacity: 0;
  }

  .tile:hover .title,
  .tile:focus-visible .title {
    opacity: 1;
  }

  .empty {
    grid-column: 1 / -1;
    margin: 0;
    padding: 4px;
    color: var(--dim);
    font-size: 12px;
  }

  section {
    position: relative;
    padding: 16px 32px;
    overflow-y: auto;
    min-height: 0;
  }

  .page {
    max-width: 640px;
  }

  h2 {
    margin: 0 0 20px;
    font-size: 15px;
    font-weight: 700;
    overflow-wrap: anywhere;
  }

  .rename {
    display: flex;
    gap: 16px;
    align-items: baseline;
    margin-bottom: 20px;
  }

  .fields {
    display: grid;
    gap: 10px;
  }

  .field {
    display: grid;
    grid-template-columns: 110px 1fr;
    gap: 12px;
  }

  .label {
    color: var(--dim);
  }

  .value {
    overflow-wrap: anywhere;
  }

  .buttons {
    display: flex;
    gap: 20px;
    margin-top: 24px;
  }

  .primary {
    color: var(--fg);
  }

  .danger,
  .error {
    color: var(--accent);
  }

  .meta {
    margin: 16px 0 0;
    color: var(--dim);
    font-size: 12px;
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
    }
    nav {
      border-right: 0;
      border-bottom: 1px solid var(--line);
    }
  }
</style>
