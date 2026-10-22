<script lang="ts">
  import { untrack } from 'svelte'
  import type { RestStyle } from './api'
  import pillar from './pillar.txt?raw'
  import skullI from './skull-i.txt?raw'
  import skullII from './skull-ii.txt?raw'

  // The idle screen of the login: one still image on black, in the lower right.
  // Random draws one of the images each time the screen rests. Any click or key
  // wakes the login; the key itself is swallowed so it does not end up in the
  // keyword.
  let { style, onwake }: { style: RestStyle; onwake: () => void } = $props()

  const images = { pillar, 'skull-i': skullI, 'skull-ii': skullII }
  type Image = keyof typeof images
  const names = Object.keys(images) as Image[]

  // Drawn once, when the screen starts resting: it never changes under a viewer.
  const image: Image = untrack(() => (style === 'random' ? names[Math.floor(Math.random() * names.length)] : style))

  function onkeydown(e: KeyboardEvent) {
    e.preventDefault()
    onwake()
  }
</script>

<svelte:window {onkeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="rest" onclick={onwake}>
  <pre class={image} aria-hidden="true">{images[image]}</pre>
  <span class="sr-only">DataVault is resting. Click or press a key to continue.</span>
</div>

<style>
  .rest {
    position: fixed;
    inset: 0;
    background: var(--bg);
    cursor: default;
  }

  pre {
    position: absolute;
    right: 4vw;
    bottom: 5vh;
    margin: 0;
    color: var(--fg);
    line-height: 1;
    user-select: none;
  }

  /* Each image is sized from its cells (a cell is 0.6 wide, 1 tall): the pillar
     fills most of the height, the skulls a bit over half, and none gets wider
     than about half the screen. */

  /* 121 x 95 cells (scripts/art/scene.py), rubble included. */
  .pillar {
    font-size: min(calc(88vh / 95), calc(50vw / 121 / 0.6));
  }

  /* 60 x 40 cells. */
  .skull-i {
    font-size: min(calc(55vh / 40), calc(50vw / 60 / 0.6));
  }

  /* 64 x 36 cells. */
  .skull-ii {
    font-size: min(calc(55vh / 36), calc(50vw / 64 / 0.6));
  }
</style>
