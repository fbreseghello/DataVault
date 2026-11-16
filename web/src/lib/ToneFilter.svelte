<script lang="ts">
  import { theme } from './theme.svelte'

  // Paints a greyscale picture in the theme's tones (docs/PERSONAGEM.md, "Nos tons
  // do tema"): black becomes --bg, white becomes --fg, and the greys between pass
  // through --line, --dim and --muted. Pictures use it as `filter: url(#hollow-tone)`.
  const STOPS = ['bg', 'line', 'dim', 'muted', 'fg']

  function channel(hex: string, i: number): number {
    const h = hex.replace('#', '')
    const full = h.length === 3 ? [...h].map((c) => c + c).join('') : h
    return parseInt(full.slice(i * 2, i * 2 + 2), 16) / 255
  }

  const tables = $derived.by(() => {
    const colors = theme.current?.colors ?? {}
    const stops = STOPS.map((k) => colors[k] ?? '#000000')
    return [0, 1, 2].map((i) => stops.map((c) => channel(c, i).toFixed(3)).join(' '))
  })
</script>

<svg class="tone" aria-hidden="true" focusable="false">
  <filter id="hollow-tone" color-interpolation-filters="sRGB">
    <!-- the picture's brightness, as grey -->
    <feColorMatrix
      type="matrix"
      values="0.2126 0.7152 0.0722 0 0  0.2126 0.7152 0.0722 0 0  0.2126 0.7152 0.0722 0 0  0 0 0 1 0"
    />
    <!-- each grey mapped along the theme's tones -->
    <feComponentTransfer>
      <feFuncR type="table" tableValues={tables[0]} />
      <feFuncG type="table" tableValues={tables[1]} />
      <feFuncB type="table" tableValues={tables[2]} />
    </feComponentTransfer>
  </filter>
</svg>

<style>
  .tone {
    position: absolute;
    width: 0;
    height: 0;
    overflow: hidden;
  }
</style>
