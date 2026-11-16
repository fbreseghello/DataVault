<script lang="ts">
  import { tick, untrack } from 'svelte'
  import { api, ApiError, type Board, type Column, type Task } from './api'
  import { hasTag, parseTitle, tagText, type Tag } from './tags'
  import { word } from './theme.svelte'

  // Siege: the day's battle plan. One board per day, four columns. Tasks move
  // by dragging or from the keyboard. The page stays mounted and returns to
  // today each time it opens.
  let { active }: { active: boolean } = $props()

  // The columns keep their ids; the theme names them.
  const COLUMNS: { id: Column; label: string; empty: string }[] = $derived(
    (['orders', 'in_battle', 'conquered', 'fallen'] as const).map((id, i) => {
      const label = word(id)
      return { id, label, empty: i ? `nothing ${label.toLowerCase()}` : `no ${label.toLowerCase()} · n to add one` }
    }),
  )

  function iso(d: Date) {
    const p = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
  }
  function fromIso(day: string) {
    const [y, m, d] = day.split('-').map(Number)
    return new Date(y, m - 1, d)
  }
  function shift(day: string, n: number) {
    const d = fromIso(day)
    d.setDate(d.getDate() + n)
    return iso(d)
  }
  function label(day: string) {
    // English like the rest of the interface: "Thu 22 Oct".
    return fromIso(day).toLocaleDateString('en-GB', { weekday: 'short', day: '2-digit', month: 'short' })
  }

  let today = $state(iso(new Date()))
  let day = $state(iso(new Date()))
  let board = $state<Board>({ tasks: [], pending_before: 0, pending_day: '' })
  let selected = $state<string | null>(null)
  let editing = $state<string | null>(null)
  let editTitle = $state('')
  let editNote = $state('')
  let adding = $state(false)
  let newTitle = $state('')
  let newNote = $state('')
  let confirmDelete = $state<string | null>(null)
  let dragId = $state<string | null>(null)
  let drop = $state<{ column: Column; index: number } | null>(null)
  let error = $state('')
  let addField = $state<HTMLInputElement>()
  let editField = $state<HTMLInputElement>()
  // a tag clicked on a card: the board shows only the tasks that carry it
  let filter = $state<Tag | null>(null)

  const isToday = $derived(day === today)

  // every task in a column, in order
  function full(c: Column): Task[] {
    return board.tasks.filter((t) => t.column === c).sort((a, b) => a.position - b.position)
  }

  // the tasks a column shows: all of them, or those with the tag being filtered
  function column(c: Column): Task[] {
    const f = filter
    return f ? full(c).filter((t) => hasTag(t.title, f.name)) : full(c)
  }

  // A place among the shown tasks, as a place in the whole column (both without
  // the task being moved): before the shown task at that place, or right after
  // the last shown one.
  function fullIndex(c: Column, index: number, moving: string): number {
    const all = full(c).filter((t) => t.id !== moving)
    const shown = column(c).filter((t) => t.id !== moving)
    if (!shown.length) return all.length
    if (index < shown.length) return all.indexOf(shown[index])
    return all.indexOf(shown[shown.length - 1]) + 1
  }

  function filterBy(tag: Tag) {
    filter = filter?.name.toLowerCase() === tag.name.toLowerCase() ? null : tag
    if (selected && !column(board.tasks.find((t) => t.id === selected)?.column ?? 'orders').some((t) => t.id === selected)) selected = null
  }

  function fail(e: unknown) {
    if (e instanceof ApiError && e.status === 423) return
    error = e instanceof Error ? e.message : String(e)
  }

  async function load() {
    try {
      board = await api.board(day)
      error = ''
    } catch (e) {
      fail(e)
    }
  }

  $effect(() => {
    if (!active) return
    untrack(() => {
      today = iso(new Date())
      day = today
      reset()
      void load()
    })
  })

  // A half-written order stays as a draft when the form closes or the day
  // changes; only cancel throws it away.
  function reset() {
    selected = editing = confirmDelete = null
    adding = false
  }

  async function go(d: string) {
    day = d
    reset()
    await load()
  }

  async function openAdd() {
    // while filtering, a new order starts with the tag, so it stays in sight
    if (filter && !newTitle.trim()) newTitle = tagText(filter) + '  '
    adding = true
    await tick()
    addField?.focus()
  }

  async function add(e: SubmitEvent) {
    e.preventDefault()
    const title = newTitle.trim()
    if (!title) return
    const note = newNote
    // Clear right away so the next order can be typed while this one saves.
    newTitle = newNote = ''
    addField?.focus()
    try {
      const t = await api.createTask(title, day, note)
      selected = t.id
      await load()
    } catch (err) {
      newTitle = title
      newNote = note
      fail(err)
    }
  }

  // Ctrl+Enter in the note adds the order; Enter alone starts a new line.
  function noteKey(e: KeyboardEvent) {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      ;(e.currentTarget as HTMLTextAreaElement).form?.requestSubmit()
    }
  }

  // The form folds away when focus leaves it with nothing typed.
  function addFocusOut(e: FocusEvent) {
    const form = e.currentTarget as HTMLFormElement
    if (form.contains(e.relatedTarget as Node | null)) return
    if (!newTitle.trim() && !newNote.trim()) adding = false
  }

  // index is a place among the tasks the column shows
  async function move(id: string, c: Column, index: number) {
    try {
      await api.moveTask(id, c, fullIndex(c, index, id))
      await load()
    } catch (e) {
      fail(e)
    }
  }

  async function startEdit(t: Task) {
    editing = t.id
    selected = t.id
    editTitle = t.title
    editNote = t.note
    await tick()
    editField?.focus()
  }

  async function saveEdit(e?: SubmitEvent) {
    e?.preventDefault()
    if (!editing) return
    const id = editing
    editing = null
    if (!editTitle.trim()) return
    try {
      await api.updateTask(id, editTitle, editNote)
      await load()
    } catch (err) {
      fail(err)
    }
  }

  async function remove(id: string) {
    if (confirmDelete !== id) {
      confirmDelete = id
      return
    }
    try {
      await api.deleteTask(id)
      confirmDelete = null
      selected = null
      await load()
    } catch (e) {
      fail(e)
    }
  }

  async function carry() {
    try {
      await api.carryTasks(day)
      await load()
    } catch (e) {
      fail(e)
    }
  }

  // ---- dragging -------------------------------------------------------------

  function onDragStart(e: DragEvent, t: Task) {
    dragId = t.id
    selected = t.id
    e.dataTransfer?.setData('text/plain', t.id)
    if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
  }

  // Where in the column the pointer is: before the first card whose middle is below it.
  function onDragOver(e: DragEvent, c: Column) {
    if (!dragId) return
    e.preventDefault()
    const cards = [...(e.currentTarget as HTMLElement).querySelectorAll<HTMLElement>('[data-card]')].filter(
      (el) => el.dataset.card !== dragId,
    )
    let index = cards.length
    for (let i = 0; i < cards.length; i++) {
      const r = cards[i].getBoundingClientRect()
      if (e.clientY < r.top + r.height / 2) {
        index = i
        break
      }
    }
    drop = { column: c, index }
  }

  async function onDrop(e: DragEvent) {
    e.preventDefault()
    const id = dragId
    const target = drop
    dragId = null
    drop = null
    if (id && target) await move(id, target.column, target.index)
  }

  function onDragEnd() {
    dragId = null
    drop = null
  }

  // ---- keyboard ---------------------------------------------------------------

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement
  }

  function where(id: string | null): { col: number; index: number } | null {
    if (!id) return null
    const t = board.tasks.find((x) => x.id === id)
    if (!t) return null
    const col = COLUMNS.findIndex((c) => c.id === t.column)
    return { col, index: column(t.column).findIndex((x) => x.id === id) }
  }

  function selectAt(col: number, index: number) {
    const tasks = column(COLUMNS[col].id)
    if (tasks.length) selected = tasks[Math.max(0, Math.min(index, tasks.length - 1))].id
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active) return
    if (e.key === 'Escape') {
      if (editing) editing = null
      else if (adding) adding = false
      else if (isTyping(e.target)) (e.target as HTMLElement).blur()
      else if (selected || confirmDelete) selected = confirmDelete = null
      else filter = null
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return

    const at = where(selected)
    switch (e.key) {
      case 'n':
        e.preventDefault()
        void openAdd()
        return
      case '[':
        void go(shift(day, -1))
        return
      case ']':
        void go(shift(day, 1))
        return
      case 't':
        void go(today)
        return
    }
    if (!at) {
      if (e.key.startsWith('Arrow')) {
        e.preventDefault()
        const first = COLUMNS.findIndex((c) => column(c.id).length > 0)
        if (first >= 0) selectAt(first, 0)
      }
      return
    }
    const id = selected!
    if (e.shiftKey && (e.key === 'ArrowRight' || e.key === 'ArrowLeft')) {
      e.preventDefault()
      const to = at.col + (e.key === 'ArrowRight' ? 1 : -1)
      if (to >= 0 && to < COLUMNS.length) void move(id, COLUMNS[to].id, column(COLUMNS[to].id).length)
    } else if (e.shiftKey && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
      e.preventDefault()
      const to = at.index + (e.key === 'ArrowDown' ? 1 : -1)
      if (to >= 0) void move(id, COLUMNS[at.col].id, to)
    } else if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
      e.preventDefault()
      let col = at.col
      do col += e.key === 'ArrowRight' ? 1 : -1
      while (col >= 0 && col < COLUMNS.length && column(COLUMNS[col].id).length === 0)
      if (col >= 0 && col < COLUMNS.length) selectAt(col, at.index)
    } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault()
      selectAt(at.col, at.index + (e.key === 'ArrowDown' ? 1 : -1))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const t = board.tasks.find((x) => x.id === id)
      if (t) void startEdit(t)
    } else if (e.key === 'Delete' || e.key === 'Backspace') {
      e.preventDefault()
      void remove(id)
    }
  }
</script>

<svelte:window {onkeydown} />

<div class="siege">
  <div class="toolbar page-toolbar">
    <button onclick={() => go(shift(day, -1))} aria-label="previous day">‹</button>
    <span class="day" class:today={isToday}>{label(day)}</span>
    <button onclick={() => go(shift(day, 1))} aria-label="next day">›</button>
    {#if !isToday}<button onclick={() => go(today)}>today</button>{/if}
    {#if filter}
      <button class="filter" onclick={() => (filter = null)} title="show every task · esc">
        <span class="tag {filter.color ?? ''}">{filter.name}</span> ×
      </button>
    {/if}
    <span class="spacer"></span>
    <span class="count">
      {#each COLUMNS as c, i (c.id)}{i ? ' · ' : ''}{column(c.id).length} {c.label.toLowerCase()}{/each}
    </span>
    <button class="new" onclick={openAdd} title={adding || !(newTitle.trim() || newNote.trim()) ? "new order · n" : "back to the draft · n"}>{!adding && (newTitle.trim() || newNote.trim()) ? "+ draft" : "+ new"}</button>
  </div>

  {#if isToday && board.pending_before > 0}
    <div class="carry">
      <span>
        {board.pending_before}
        {board.pending_before === 1 ? 'task' : 'tasks'} left on the field since {label(board.pending_day)}
      </span>
      <button onclick={carry}>bring them to today</button>
    </div>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="board">
    {#each COLUMNS as c (c.id)}
      {@const tasks = column(c.id)}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <section
        class="col {c.id}"
        class:target={drop?.column === c.id}
        ondragover={(e) => onDragOver(e, c.id)}
        ondrop={onDrop}
        aria-label={c.label}
      >
        <h3>{c.label} <span>{tasks.length}</span></h3>

        {#if c.id === 'orders' && adding}
          <form onsubmit={add} onfocusout={addFocusOut} class="add">
            <input
              bind:this={addField}
              bind:value={newTitle}
              placeholder="new order · !tag(r)  to tag it · enter to add"
              spellcheck="false"
              autocomplete="off"
              aria-label="title"
            />
            <textarea bind:value={newNote} onkeydown={noteKey} placeholder="note · ctrl+enter to add" rows="2" aria-label="note"></textarea>
            <div class="actions">
              <button type="submit">add</button>
              <button type="button" onclick={() => ((adding = false), (newTitle = newNote = ''))} title="throw the draft away">cancel</button>
            </div>
          </form>
        {/if}

        {#each tasks as t, i (t.id)}
          {#if drop?.column === c.id && drop.index === i && dragId !== t.id}<div class="slot"></div>{/if}
          <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
          <div
            class="card"
            class:selected={selected === t.id}
            class:dragging={dragId === t.id}
            data-card={t.id}
            draggable={editing !== t.id}
            ondragstart={(e) => onDragStart(e, t)}
            ondragend={onDragEnd}
            onclick={() => (selected = t.id)}
            ondblclick={() => startEdit(t)}
          >
            {#if editing === t.id}
              <form onsubmit={saveEdit} class="edit">
                <input bind:this={editField} bind:value={editTitle} spellcheck="false" autocomplete="off" aria-label="title" />
                <textarea bind:value={editNote} placeholder="note" rows="2" aria-label="note"></textarea>
                <div class="actions">
                  <button type="submit">save</button>
                  <button type="button" onclick={() => (editing = null)}>cancel</button>
                </div>
              </form>
            {:else}
              {@const parsed = parseTitle(t.title)}
              <span class="title">
                {#each parsed.tags as tag, k (k)}<button
                    type="button"
                    class="tag {tag.color ?? ''}"
                    class:on={filter?.name.toLowerCase() === tag.name.toLowerCase()}
                    title="show only {tag.name}"
                    onclick={(e) => (e.stopPropagation(), filterBy(tag))}>{tag.name}</button
                  >{/each}{parsed.text}
              </span>
              {#if t.note}<span class="note">{t.note}</span>{/if}
              {#if selected === t.id}
                <span class="actions">
                  <button onclick={() => startEdit(t)}>edit</button>
                  <button onclick={() => remove(t.id)} class:danger={confirmDelete === t.id}>
                    {confirmDelete === t.id ? 'delete? again' : 'delete'}
                  </button>
                </span>
              {/if}
            {/if}
          </div>
        {/each}
        {#if drop?.column === c.id && drop.index >= tasks.filter((x) => x.id !== dragId).length}<div class="slot"></div>{/if}

        {#if !tasks.length && !(c.id === 'orders' && adding)}<p class="empty">{c.empty}</p>{/if}
      </section>
    {/each}
  </div>
</div>

<style>
  .siege {
    height: 100%;
    display: grid;
    grid-template-rows: auto auto auto 1fr;
    min-height: 0;
  }


  .day {
    min-width: 11ch;
    text-align: center;
    color: var(--muted);
  }

  .day.today {
    color: var(--fg);
  }

  .dim {
    color: var(--dim);
  }

  .carry {
    display: flex;
    gap: 16px;
    align-items: baseline;
    margin: 0 24px 8px;
    padding: 6px 12px;
    border-left: 2px solid var(--accent);
    background: var(--raise);
    color: var(--muted);
    font-size: 12px;
  }

  .carry button {
    color: var(--fg);
  }

  .error {
    margin: 0 24px 8px;
    color: var(--accent);
  }

  .board {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1px;
    background: var(--line);
    border-top: 1px solid var(--line);
    min-height: 0;
  }

  .col {
    background: var(--bg);
    padding: 12px 14px;
    overflow-y: auto;
    min-height: 0;
  }

  .col.target {
    background: #0d0d0d;
  }

  h3 {
    margin: 0 0 12px;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.04em;
    display: flex;
    gap: 8px;
    align-items: baseline;
  }

  h3 span {
    color: var(--dim);
    font-weight: 400;
  }

  .card {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    margin-bottom: 6px;
    border: 1px solid var(--line);
    background: var(--raise);
    cursor: grab;
  }

  .card.selected {
    border-color: var(--muted);
  }

  .card.dragging {
    opacity: 0.35;
  }

  .title {
    overflow-wrap: anywhere;
  }

  /* a tag from "!name" or "!name(r)" at the start of the title */
  .tag {
    --tag: var(--accent);
    display: inline-block;
    margin-right: 6px;
    padding: 0 4px;
    border: 1px solid var(--tag);
    color: var(--tag);
    font-size: 11px;
    line-height: 15px;
    vertical-align: 1px;
    text-decoration: none;
  }

  .tag.red {
    --tag: var(--red);
  }

  .tag.blue {
    --tag: var(--blue);
  }

  .tag.green {
    --tag: var(--green);
  }

  button.tag {
    font: inherit;
    font-size: 11px;
    background: transparent;
    cursor: pointer;
  }

  button.tag:hover,
  .tag.on {
    background: color-mix(in srgb, var(--tag) 18%, transparent);
  }

  .filter {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    color: var(--muted);
  }

  .filter .tag {
    margin-right: 4px;
  }

  .conquered .tag,
  .fallen .tag {
    opacity: 0.55;
  }

  .note {
    color: var(--dim);
    font-size: 11px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .conquered .title {
    color: var(--muted);
  }

  .fallen .title {
    color: var(--dim);
    text-decoration: line-through;
  }

  .actions {
    display: flex;
    gap: 12px;
    font-size: 11px;
  }

  .danger {
    color: var(--accent);
  }

  .slot {
    height: 2px;
    margin: 0 0 6px;
    background: var(--fg);
  }

  .add input,
  .edit input {
    width: 100%;
  }

  .add {
    margin-bottom: 8px;
  }

  .add,
  .edit {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .add textarea,
  .edit textarea {
    min-height: 2.6em;
    font-size: 11px;
  }

  .empty {
    color: var(--dim);
    font-size: 11px;
    margin: 4px 0;
  }

  @media (max-width: 900px) {
    .board {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>
