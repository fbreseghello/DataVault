<script lang="ts">
  import { artFor } from './art'
  import { roman } from './companions'

  // A companion's card (docs/PERSONAGEM.md, "A página: a carta"): the picture in
  // the theme's tones, the level as a roman numeral at the top of the frame, the
  // name at its foot. With progress, the frame is the experience bar: it is drawn
  // clockwise from the numeral as the level fills, and the rest is dotted.
  let {
    companion,
    level = 0,
    progress,
    reversed = false,
    width,
  }: {
    companion: string
    /** The level, shown as the numeral; 0 shows none. */
    level?: number
    /** How much of the level is done, 0 to 1; without it the frame is whole. */
    progress?: number
    /** Upside down: the reading of a vault with something pending. */
    reversed?: boolean
    /** The card's width in pixels; the height follows at 2:3. */
    width: number
  } = $props()

  const height = $derived(Math.round(width * 1.5))
  // the frame sits a little inside the card's edge
  const m = 6
  const frame = $derived(`M${width / 2} ${m}H${width - m}V${height - m}H${m}V${m}Z`)
  const done = $derived(progress === undefined ? 1 : Math.max(0, Math.min(1, progress)))
  const numeral = $derived(roman(level))
</script>

<div class="card" class:reversed style:width="{width}px" style:height="{height}px">
  <img class="art" src={artFor(companion)} alt="" draggable="false" />
  <svg class="frame" {width} {height} aria-hidden="true">
    <path class="left" d={frame} />
    <path class="done" d={frame} pathLength="1000" stroke-dasharray="{done * 1000} 1000" />
  </svg>
  {#if numeral}<span class="numeral">{numeral}</span>{/if}
  <span class="name">the {companion}</span>
</div>

<style>
  .card {
    position: relative;
    flex: none;
    transition: transform 0.6s ease;
  }

  .card.reversed {
    transform: rotate(180deg);
  }

  @media (prefers-reduced-motion: reduce) {
    .card {
      transition: none;
    }
  }

  .art {
    position: absolute;
    inset: 16px;
    width: calc(100% - 32px);
    height: calc(100% - 32px);
    object-fit: cover;
    filter: url(#hollow-tone);
    user-select: none;
  }

  .frame {
    position: absolute;
    inset: 0;
    overflow: visible;
  }

  .frame path {
    fill: none;
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  /* what the level still needs: dotted, barely there */
  .left {
    stroke: var(--line);
    stroke-dasharray: 1 3;
  }

  /* what the level has: the frame drawn in ink */
  .done {
    stroke: var(--fg);
  }

  .numeral,
  .name {
    position: absolute;
    left: 50%;
    transform: translate(-50%, -50%);
    padding: 0 8px;
    background: var(--bg);
    white-space: nowrap;
    line-height: 1;
  }

  .numeral {
    top: 6px;
    font-size: 13px;
    letter-spacing: 0.12em;
    color: var(--fg);
  }

  .name {
    top: calc(100% - 6px);
    font-size: 11px;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--muted);
  }
</style>
