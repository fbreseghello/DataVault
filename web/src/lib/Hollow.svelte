<script lang="ts">
  import { untrack } from 'svelte'
  import { api, type HollowClass } from './api'
  import { ATTRS, bar, byDay, CLASSES, kindName, titleFor } from './character'
  import { hollow, refreshHollow, seeHollow } from './character.svelte'
  import { showDate } from './renew'
  import Sprite from './Sprite.svelte'
  import { theme, word } from './theme.svelte'

  // The character that grows with the vault (docs/PERSONAGEM.md). On the left,
  // the deeds that earned its experience, by day; on the right, the character,
  // its level and its attributes. The first time, the page asks for a class.
  let { active }: { active: boolean } = $props()

  // picking a class: the first time, or from the class button
  let picking = $state(false)
  let pick = $state(0)
  let error = $state('')
  const choosing = $derived(!!hollow.state && (picking || !hollow.state.class))

  const s = $derived(hollow.state)
  const title = $derived(s ? titleFor(s.level, theme.current?.titles ?? []) : '')
  const days = $derived(s ? byDay(s.deeds) : [])
  const top = $derived(s ? Math.max(...ATTRS.map((a) => s.attrs[a] ?? 0)) : 0)

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

  function startPicking() {
    pick = Math.max(0, CLASSES.findIndex((c) => c.id === hollow.state?.class))
    error = ''
    picking = true
  }

  async function choose(id: HollowClass) {
    try {
      hollow.state = await api.setHollowClass(id)
      picking = false
      error = ''
    } catch (e) {
      error = e instanceof Error ? e.message : String(e)
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active || !choosing || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
      e.preventDefault()
      pick = (pick + (e.key === 'ArrowRight' ? 1 : CLASSES.length - 1)) % CLASSES.length
    } else if (e.key === 'Enter') {
      e.preventDefault()
      void choose(CLASSES[pick].id)
    } else if (e.key === 'Escape' && hollow.state?.class) {
      picking = false
    }
  }
</script>

<svelte:window {onkeydown} />

<div class="hollow">
  <div class="toolbar page-toolbar">
    {#if s}
      <span>lv {s.level}</span>
      <span class="dim">· {title}</span>
    {/if}
    <span class="spacer"></span>
    {#if s}<span class="count">{s.need - s.into} {word('xp')} to next</span>{/if}
    {#if s?.class && !choosing}<button onclick={startPicking} title="change the class · free, costs no {word('xp')}">class</button>{/if}
  </div>

  {#if choosing}
    <div class="choose">
      <p class="ask">{s?.class ? 'choose another class · it costs nothing' : `choose what your ${word('hollow')} becomes`}</p>
      <div class="classes" role="radiogroup" aria-label="class">
        {#each CLASSES as c, i (c.id)}
          <button
            class="class"
            class:on={pick === i}
            role="radio"
            aria-checked={pick === i}
            onclick={() => (pick === i ? choose(c.id) : (pick = i))}
            ondblclick={() => choose(c.id)}
          >
            <Sprite name={c.id} scale={3} />
            <span class="cname">{word(c.id)}</span>
            <span class="look">{c.look}</span>
            <span class="lean">leans to {word(c.attr)}</span>
          </button>
        {/each}
      </div>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <p class="hint">arrows choose · enter confirms{s?.class ? ' · esc keeps the current one' : ''} · it can be changed later</p>
    </div>
  {:else}
  <div class="body">
    <nav aria-label={word('deeds')}>
      <p class="group first">{word('deeds')}</p>
      {#each days as g (g.day)}
        <p class="group">{dayLabel(g.day)}</p>
        {#each g.deeds as d (d.source)}
          <div class="deed">
            <span class="gain">+{d.xp}</span>
            <span class="kind">{kindName(d.kind)}</span>
            <span class="dim">{time(d.at)}</span>
          </div>
        {/each}
      {:else}
        <p class="empty">nothing yet · conquer a task in the {word('siege')}</p>
      {/each}
    </nav>

    <section>
      {#if s}
        <div class="sheet">
          {#if s.class}<span class="figure"><Sprite name={s.class} scale={4} /></span>{/if}
          <p class="name">
            {word('level')} {s.level}
            {#if s.class}<span class="dim">· {word(s.class)}</span>{/if}
            <span class="dim">· {title}</span>
          </p>
          <p class="xp">
            <span class="dim">{word('xp')}</span>
            [{bar(s.into, s.need)}]
            <span class="nums">{s.into} / {s.need}</span>
          </p>
          <div class="attrs">
            {#each ATTRS as a (a)}
              <p>
                <span class="dim label">{word(a)}</span>
                <span class="meter">{bar(s.attrs[a] ?? 0, top, 10)}</span>
                <span class="nums">{s.attrs[a] ?? 0}</span>
              </p>
            {/each}
          </div>
        </div>
      {/if}
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
    grid-template-columns: minmax(240px, 360px) 1fr;
    min-height: 0;
  }

  nav {
    border-right: 1px solid var(--line);
    padding: 8px 0;
    overflow-y: auto;
  }

  .group {
    margin: 16px 0 4px;
    padding: 0 24px;
    color: var(--dim);
    font-size: 11px;
  }

  .group.first {
    margin-top: 4px;
    color: var(--muted);
  }

  .deed {
    display: grid;
    grid-template-columns: 4ch 1fr auto;
    gap: 10px;
    padding: 3px 24px;
    font-variant-numeric: tabular-nums;
  }

  .gain {
    color: var(--muted);
  }

  .empty {
    color: var(--dim);
    padding: 0 24px;
  }

  section {
    position: relative;
    padding: 32px;
    overflow-y: auto;
    min-height: 0;
  }

  .sheet {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    font-variant-numeric: tabular-nums;
  }

  .figure {
    margin-bottom: 16px;
  }

  .choose {
    padding: 32px 24px;
    overflow-y: auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
  }

  .ask {
    margin: 0;
    color: var(--muted);
  }

  .classes {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 16px;
  }

  .class {
    width: 150px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 16px 8px;
    border: 1px solid transparent;
    color: var(--dim);
    text-align: center;
  }

  .class:hover {
    color: var(--muted);
  }

  .class.on {
    border-color: var(--line);
    background: var(--raise);
    color: var(--fg);
  }

  .cname {
    font-weight: 700;
    margin-top: 8px;
  }

  .look,
  .lean {
    font-size: 11px;
    color: var(--dim);
  }

  .class.on .look {
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

  .name {
    margin: 0;
    font-weight: 700;
  }

  .xp {
    margin: 0;
    white-space: pre;
  }

  .nums {
    color: var(--muted);
  }

  .attrs {
    margin-top: 20px;
  }

  .attrs p {
    margin: 2px 0;
    display: grid;
    grid-template-columns: 10ch 10ch 5ch;
    gap: 12px;
    white-space: pre;
  }

  .attrs .nums {
    text-align: right;
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
