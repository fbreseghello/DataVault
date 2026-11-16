<script lang="ts">
  import type { HollowClass } from './api'
  import { spritePaths } from './character'
  import sprites from './sprites.json'

  // The hollow in pixel art: a square grid (64x64, or 32x32 for the ones not yet
  // redrawn) in the theme's four tones, one SVG path per tone. It is drawn at the
  // largest whole multiple that fits in box pixels, so every art pixel lands on
  // whole screen pixels (docs/PERSONAGEM.md, "O sprite").
  let { name, box }: { name: HollowClass; box: number } = $props()

  const rows = $derived((sprites as Record<HollowClass, string[]>)[name])
  const n = $derived(rows.length)
  const scale = $derived(Math.max(1, Math.floor(box / n)))
  const paths = $derived(spritePaths(rows))
</script>

<svg
  class="sprite"
  width={n * scale}
  height={n * scale}
  viewBox="0 0 {n} {n}"
  shape-rendering="crispEdges"
  aria-hidden="true"
>
  {#each paths as p (p.tone)}
    <path d={p.d} class="t{p.tone}" />
  {/each}
</svg>

<style>
  .sprite {
    display: block;
  }

  /* the outline and deepest shadow, the shadow, the middle and the light */
  .t1 {
    fill: var(--line);
  }

  .t2 {
    fill: var(--dim);
  }

  .t3 {
    fill: var(--muted);
  }

  .t4 {
    fill: var(--fg);
  }
</style>
