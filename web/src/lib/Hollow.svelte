<script lang="ts">
  import { untrack } from 'svelte'
  import { api } from './api'
  import Card from './Card.svelte'
  import CardBack from './CardBack.svelte'
  import { ATTRS, bar, byDay, kindName, titleFor } from './character'
  import { hollow, refreshHollow, seeHollow } from './character.svelte'
  import { categoryOf, ordered, roman, type Category } from './companions'
  import { showDate } from './renew'
  import ToneFilter from './ToneFilter.svelte'
  import { theme, word } from './theme.svelte'

  // The companion that grows with the vault (docs/PERSONAGEM.md). The page is a
  // tarot card: on the left the card, whose frame is the experience bar, the
  // level its numeral; on the right the reading, in plain text. The first time,
  // the page asks for a companion: pick a category's card, then one of its own.
  let { active }: { active: boolean } = $props()

  // how strongly each of the latest deeds is drawn, newest first
  const FADE = [1, 1, 1, 1, 0.7, 0.45, 0.25, 0.1]

  const s = $derived(hollow.state)
  const title = $derived(s ? titleFor(s.level, theme.current?.titles ?? []) : '')
  // the latest deeds: four in full, then four fading out, and no more
  const shown = $derived(s ? s.deeds.slice(0, FADE.length) : [])
  const days = $derived(byDay(shown))
  const top = $derived(s ? Math.max(...ATTRS.map((a) => s.attrs[a] ?? 0)) : 0)

  // ---- choosing: a category face down, then a companion face up ---------------

  let picking = $state(false)
  let category = $state<Category | null>(null)
  let pick = $state(0)
  let error = $state('')
  const choosing = $derived(!!s && (picking || !s.companion))
  const categories = $derived(ordered())
  // what the arrows move through: the categories, or the open category's companions
  const row = $derived(category ? category.companions.length : categories.length)

  function startPicking() {
    const current = s?.companion ? categoryOf(s.companion) : undefined
    category = null
    pick = Math.max(0, categories.findIndex((c) => c === current))
    error = ''
    picking = true
  }

  function openCategory(c: Category) {
    category = c
    pick = Math.max(0, c.companions.findIndex((x) => x.id === s?.companion))
  }

  async function choose(id: string) {
    try {
      hollow.state = await api.setHollowCompanion(id)
      picking = false
      category = null
      error = ''
    } catch (e) {
      error = e instanceof Error ? e.message : String(e)
    }
  }

  function confirm() {
    if (category) void choose(category.companions[pick].id)
    else openCategory(categories[pick])
  }

  // ---- sizes: the card takes the largest 2:3 that fits its pane ----------------

  let stageW = $state(0)
  let stageH = $state(0)
  const cardW = $derived(Math.max(160, Math.floor(Math.min((stageW - 64) / 1, (stageH - 96) / 1.5))))
  let chooseW = $state(0)
  // the choice deals its cards in one row when it can
  const dealW = $derived(Math.max(110, Math.min(170, Math.floor((chooseW - 48 - 16 * (row - 1)) / Math.max(row, 1)))))

  $effect(() => {
    if (!active) return
    untrack(async () => {
      await refreshHollow()
      await seeHollow()
    })
  })

  function dayLabel(day: string): string {
    const d = new Date()
    const iso = (x: Date) => `${x.getFullYear()}-${String(x.getMonth() + 1).padStart(2, '0')}-${String(x.getDate()).padStart(2, '0')}`
    if (day === iso(d)) return 'today'
    d.setDate(d.getDate() - 1)
    if (day === iso(d)) return 'yesterday'
    return showDate(day)
  }

  function time(at: string): string {
    return new Date(at).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active || e.ctrlKey || e.metaKey || e.altKey || isTyping(e.target)) return
    if (!choosing) {
      if (e.key === 'c' && s) {
        e.preventDefault()
        startPicking()
      }
      return
    }
    if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
      e.preventDefault()
      pick = (pick + (e.key === 'ArrowRight' ? 1 : row - 1)) % row
    } else if (e.key === 'Enter') {
      e.preventDefault()
      confirm()
    } else if (e.key === 'Escape') {
      if (category) {
        pick = Math.max(0, categories.indexOf(category))
        category = null
      } else if (s?.companion) picking = false
    }
  }
</script>

<svelte:window {onkeydown} />
<ToneFilter />

<div class="hollow">
  <div class="toolbar page-toolbar">
    {#if s}
      <span>{roman(s.level)}</span>
      <span class="dim">· {title}</span>
    {/if}
    <span class="spacer"></span>
    {#if s && !choosing}
      <span class="count">{s.need - s.into} {word('xp')} to next</span>
      <button onclick={startPicking} title="choose another companion · c · it costs nothing">companion</button>
    {/if}
  </div>

  {#if choosing}
    <div class="choose" bind:clientWidth={chooseW}>
      <p class="ask">
        {#if category}
          {category.id} · choose one
        {:else if s?.companion}
          draw another card · it costs nothing
        {:else}
          draw a card · what will walk with you
        {/if}
      </p>
      <div class="deal" role="radiogroup" aria-label={category ? 'companion' : 'category'}>
        {#if category}
          {#each category.companions as c, i (c.id)}
            <button
              class="pick"
              class:on={pick === i}
              role="radio"
              aria-checked={pick === i}
              aria-label={c.id}
              onclick={() => (pick === i ? choose(c.id) : (pick = i))}
              ondblclick={() => choose(c.id)}
            >
              <Card companion={c.id} width={dealW} />
              <span class="look">{c.look}</span>
            </button>
          {/each}
        {:else}
          {#each categories as c, i (c.id)}
            <button
              class="pick"
              class:on={pick === i}
              role="radio"
              aria-checked={pick === i}
              aria-label={c.id}
              onclick={() => (pick === i ? openCategory(c) : (pick = i))}
              ondblclick={() => openCategory(c)}
            >
              <CardBack seal={c.seal} label={c.id} width={dealW} />
              <span class="look">{c.companions.length} · {c.sealName}</span>
            </button>
          {/each}
        {/if}
      </div>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <p class="hint">
        arrows choose · enter {category ? 'turns the card' : 'opens the category'}{category
          ? ' · esc back'
          : s?.companion
            ? ' · esc keeps yours'
            : ''}
      </p>
    </div>
  {:else if s}
    <div class="body">
      <!-- the card on the left, as large as the pane allows -->
      <div class="stage" bind:clientWidth={stageW} bind:clientHeight={stageH}>
        {#if stageW}
          <Card companion={s.companion} level={s.level} progress={s.need ? s.into / s.need : 0} width={cardW} />
        {/if}
      </div>

      <!-- the reading on the right: plain text, like notes on a spread -->
      <section class="reading">
        <p class="name">{roman(s.level)} · the {s.companion}</p>
        <p class="dim">{title}</p>

        <p class="xp">
          <span class="dim label">{word('xp')}</span>
          <span class="nums">{s.into} / {s.need}</span>
        </p>
        <div class="attrs">
          {#each ATTRS as a (a)}
            <p>
              <span class="dim label">{word(a)}</span>
              <span class="meter">{bar(s.attrs[a] ?? 0, top, 16)}</span>
              <span class="nums">{s.attrs[a] ?? 0}</span>
            </p>
          {/each}
        </div>

        <!-- the latest deeds sit at the foot, quiet: the reading comes first -->
        <div class="deeds">
        <p class="group first">{word('deeds')}</p>
        {#each days as g (g.day)}
          <p class="group">{dayLabel(g.day)}</p>
          {#each g.deeds as d (d.source)}
            <div class="deed" style:opacity={FADE[shown.indexOf(d)]}>
              <span class="gain">+{d.xp}</span>
              <span class="kind">{kindName(d.kind)}</span>
              <span class="dim">{time(d.at)}</span>
            </div>
          {/each}
        {:else}
          <p class="empty">nothing yet · conquer a task in the {word('siege')}</p>
        {/each}
        </div>
      </section>
    </div>
  {/if}
</div>

<style>
  .hollow {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  .dim {
    color: var(--dim);
  }

  .body {
    display: grid;
    grid-template-columns: minmax(0, 3fr) minmax(320px, 2fr);
    min-height: 0;
  }

  .stage {
    display: grid;
    place-items: center;
    min-height: 0;
    overflow: hidden;
    border-right: 1px solid var(--line);
  }

  .reading {
    padding: 32px;
    overflow-y: auto;
    min-height: 0;
    font-variant-numeric: tabular-nums;
    display: flex;
    flex-direction: column;
  }

  /* pushed to the foot of the column, smaller and dimmer than the reading */
  .deeds {
    margin-top: auto;
    padding-top: 32px;
    font-size: 11px;
  }

  .reading p {
    margin: 0;
  }

  .name {
    font-weight: 700;
    letter-spacing: 0.04em;
  }

  .xp,
  .attrs p {
    margin: 2px 0;
    display: grid;
    grid-template-columns: 10ch auto 1fr;
    gap: 12px;
    white-space: pre;
  }

  .xp {
    margin-top: 24px !important;
  }

  .attrs {
    margin: 12px 0 8px;
  }

  .nums {
    color: var(--muted);
  }

  .group {
    margin: 16px 0 4px !important;
    color: var(--dim);
    font-size: 11px;
  }

  .group.first {
    margin-top: 0 !important;
    color: var(--dim);
  }

  .deed {
    display: grid;
    grid-template-columns: 4ch 1fr auto;
    gap: 10px;
    padding: 2px 0;
    max-width: 420px;
    color: var(--muted);
  }

  .gain {
    color: var(--dim);
  }

  .empty {
    color: var(--dim);
  }

  .choose {
    padding: 32px 24px;
    overflow-y: auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 28px;
  }

  .ask {
    margin: 0;
    color: var(--muted);
  }

  .deal {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 16px;
  }

  .pick {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    padding: 0;
    opacity: 0.55;
    transition: opacity 0.2s ease, transform 0.2s ease;
  }

  .pick:hover {
    opacity: 0.8;
  }

  /* the card under the cursor is lifted out of the spread */
  .pick.on {
    opacity: 1;
    transform: translateY(-8px);
  }

  @media (prefers-reduced-motion: reduce) {
    .pick {
      transition: none;
    }
  }

  .look {
    max-width: 170px;
    font-size: 11px;
    color: var(--dim);
    text-align: center;
  }

  .pick.on .look {
    color: var(--muted);
  }

  .hint {
    margin: 0;
    font-size: 11px;
    color: var(--dim);
  }

  .error {
    color: var(--accent);
    margin: 0;
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
      grid-template-rows: 70vh auto;
      overflow-y: auto;
    }
    .stage {
      border-right: 0;
      border-bottom: 1px solid var(--line);
    }
  }
</style>
