<script lang="ts">
  import { tick, untrack } from 'svelte'
  import { api, ApiError, type Board, type Column, type Task } from './api'

  // Siege: the day's battle plan. One board per day, four columns. Tasks move
  // by dragging or from the keyboard. The page stays mounted and returns to
  // today each time it opens.
  let { active }: { active: boolean } = $props()

  const COLUMNS: { id: Column; label: string; empty: string }[] = [
    { id: 'orders', label: 'Orders', empty: 'no orders · n to give one' },
    { id: 'in_battle', label: 'In Battle', empty: 'nothing in battle' },
    { id: 'conquered', label: 'Conquered', empty: 'nothing conquered yet' },
    { id: 'fallen', label: 'Fallen', empty: 'none fallen' },
  ]

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
  let confirmDelete = $state<string | null>(null)
  let dragId = $state<string | null>(null)
  let drop = $state<{ column: Column; index: number } | null>(null)
  let error = $state('')
  let addField = $state<HTMLInputElement>()
  let editField = $state<HTMLInputElement>()

  const isToday = $derived(day === today)

  function column(c: Column): Task[] {
    return board.tasks.filter((t) => t.column === c).sort((a, b) => a.position - b.position)
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

  function reset() {
    selected = editing = confirmDelete = null
    adding = false
    newTitle = ''
  }

  async function go(d: string) {
    day = d
    reset()
    await load()
  }

  async function openAdd() {
    adding = true
    await tick()
    addField?.focus()
  }

  async function add(e: SubmitEvent) {
    e.preventDefault()
    const title = newTitle.trim()
    if (!title) return
    // Clear right away so the next order can be typed while this one saves.
    newTitle = ''
    try {
      const t = await api.createTask(title, day)
      selected = t.id
      await load()
    } catch (err) {
      newTitle = title
      fail(err)
    }
  }

  async function move(id: string, c: Column, index: number) {
    try {
      await api.moveTask(id, c, index)
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
      else selected = confirmDelete = null
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
  <div class="toolbar">
    <button onclick={() => go(shift(day, -1))} aria-label="previous day">‹</button>
    <span class="day" class:today={isToday}>{label(day)}</span>
    <button onclick={() => go(shift(day, 1))} aria-label="next day">›</button>
    {#if !isToday}<button onclick={() => go(today)}>today</button>{/if}
    <button onclick={openAdd}>+ order</button>
    <span class="spacer"></span>
    <span class="dim tally">
      {#each COLUMNS as c, i (c.id)}{i ? ' · ' : ''}{column(c.id).length} {c.label.toLowerCase()}{/each}
    </span>
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
          <form onsubmit={add} class="add">
            <input
              bind:this={addField}
              bind:value={newTitle}
              placeholder="new order · enter to add"
              spellcheck="false"
              autocomplete="off"
              onblur={() => !newTitle.trim() && (adding = false)}
            />
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
              <span class="title">{t.title}</span>
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

  .toolbar {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 12px 24px;
  }

  .day {
    min-width: 11ch;
    text-align: center;
    color: var(--muted);
  }

  .day.today {
    color: var(--fg);
  }

  .spacer {
    flex: 1;
  }

  .dim,
  .tally {
    color: var(--dim);
  }

  .tally {
    font-size: 11px;
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

  .edit {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

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
