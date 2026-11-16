<script lang="ts">
  import { toPath } from './icons'
  import { theme } from './theme.svelte'

  // The back of a card: the theme's hollow icon repeated as a lattice, so each
  // theme has its own deck, and a category's seal in the middle
  // (docs/PERSONAGEM.md, "Escolher e trocar: tirar uma carta").
  let { seal, label, width }: { seal: string[]; label: string; width: number } = $props()

  const height = $derived(Math.round(width * 1.5))
  const lattice = $derived(toPath(theme.current?.icons.hollow ?? []))
  const mark = $derived(toPath(seal))
  const id = `back-${Math.random().toString(36).slice(2, 8)}`
</script>

<div class="back" style:width="{width}px" style:height="{height}px">
  <svg {width} {height} aria-hidden="true" shape-rendering="crispEdges">
    <defs>
      <pattern {id} width="24" height="24" patternUnits="userSpaceOnUse">
        <path d={lattice} transform="translate(4 4)" class="lattice" />
      </pattern>
    </defs>
    <rect x="6" y="6" width={width - 12} height={height - 12} fill="url(#{id})" />
    <rect x="6.5" y="6.5" width={width - 13} height={height - 13} class="edge" />
    <!-- the seal, on a plain field so it reads over the lattice -->
    <rect x={width / 2 - 28} y={height / 2 - 28} width="56" height="56" class="field" />
    <path d={mark} transform="translate({width / 2 - 16} {height / 2 - 16}) scale(2)" class="seal" />
  </svg>
  <span class="label">{label}</span>
</div>

<style>
  .back {
    position: relative;
    flex: none;
  }

  svg {
    display: block;
  }

  .lattice {
    fill: var(--line);
  }

  .edge {
    fill: none;
    stroke: var(--dim);
  }

  .field {
    fill: var(--bg);
    stroke: var(--dim);
  }

  .seal {
    fill: var(--muted);
  }

  .label {
    position: absolute;
    left: 50%;
    top: calc(100% - 6px);
    transform: translate(-50%, -50%);
    padding: 0 8px;
    background: var(--bg);
    font-size: 11px;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--muted);
    white-space: nowrap;
  }
</style>
