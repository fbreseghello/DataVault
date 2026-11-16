<script lang="ts">
  import type { HollowClass } from './api'
  import { spritePaths } from './character'
  import sprites from './sprites.json'

  // The hollow in pixel art: a 32x32 grid in the theme's four tones, one SVG
  // path per tone, drawn at a whole multiple so every art pixel lands on whole
  // screen pixels (docs/PERSONAGEM.md, "O sprite").
  let { name, scale = 4 }: { name: HollowClass; scale?: number } = $props()

  const paths = $derived(spritePaths((sprites as Record<HollowClass, string[]>)[name]))
</script>

<svg
  class="sprite"
  width={32 * scale}
  height={32 * scale}
  viewBox="0 0 32 32"
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
