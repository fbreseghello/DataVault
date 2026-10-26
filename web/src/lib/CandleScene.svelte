<script lang="ts">
  import { onMount } from 'svelte'
  import { draw, Flame, SIZE } from './candle'

  // The candle rest screen, drawn on a canvas at about 15 frames a second. With
  // "reduce motion" the flame stands still and the scene is drawn once.
  let canvas = $state<HTMLCanvasElement>()

  onMount(() => {
    const el = canvas!
    const ctx = el.getContext('2d')!
    const flame = new Flame()
    const calm = matchMedia('(prefers-reduced-motion: reduce)')
    // the page's text colour, as "r,g,b"
    const hex = getComputedStyle(document.documentElement).getPropertyValue('--fg').trim() || '#e4e4e4'
    const rgb = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16)).join(',')
    let cw = 0
    let size = 0
    let frame = 0
    let last = 0

    function layout() {
      // most of the height, never more than about 60% of the width
      size = Math.min((innerHeight * 0.85) / SIZE.rows, (innerWidth * 0.62) / SIZE.cols / 0.6)
      cw = size * 0.6
      const dpr = devicePixelRatio || 1
      el.style.width = `${SIZE.cols * cw}px`
      el.style.height = `${SIZE.rows * size}px`
      el.width = Math.round(SIZE.cols * cw * dpr)
      el.height = Math.round(SIZE.rows * size * dpr)
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
      ctx.font = `${size}px 'JetBrains Mono', monospace`
      ctx.textBaseline = 'top'
      paint(performance.now())
    }

    function paint(t: number) {
      ctx.clearRect(0, 0, el.width, el.height)
      draw(ctx, flame, t, cw, size, rgb)
    }

    function tick(t: number) {
      frame = requestAnimationFrame(tick)
      if (calm.matches || t - last < 66) return
      last = t
      flame.step(t)
      paint(t)
    }

    let gone = false
    document.fonts.load(`16px 'JetBrains Mono'`).then(() => {
      if (gone) return
      layout()
      frame = requestAnimationFrame(tick)
    })
    addEventListener('resize', layout)
    return () => {
      gone = true
      cancelAnimationFrame(frame)
      removeEventListener('resize', layout)
    }
  })
</script>

<canvas bind:this={canvas} aria-hidden="true"></canvas>

<style>
  canvas {
    position: absolute;
    right: 4vw;
    bottom: 5vh;
  }
</style>
