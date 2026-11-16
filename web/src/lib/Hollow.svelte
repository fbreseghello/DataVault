<script lang="ts">
  import { untrack } from 'svelte'
  import { ATTRS, bar, byDay, kindName, titleFor } from './character'
  import { hollow, refreshHollow, seeHollow } from './character.svelte'
  import Icon from './Icon.svelte'
  import { showDate } from './renew'
  import { theme, word } from './theme.svelte'

  // The character that grows with the vault (docs/PERSONAGEM.md). On the left,
  // the deeds that earned its experience, by day; on the right, the character,
  // its level and its attributes. The sprite comes in a later version; until
  // then the theme's icon stands in for it.
  let { active }: { active: boolean } = $props()

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
</script>

<div class="hollow">
  <div class="toolbar page-toolbar">
    {#if s}
      <span>lv {s.level}</span>
      <span class="dim">· {title}</span>
    {/if}
    <span class="spacer"></span>
    {#if s}<span class="count">{s.need - s.into} {word('xp')} to next</span>{/if}
  </div>

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
          <span class="figure"><Icon name="hollow" size={128} /></span>
          <p class="name">{word('level')} {s.level} <span class="dim">· {title}</span></p>
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
    color: var(--muted);
    margin-bottom: 16px;
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
