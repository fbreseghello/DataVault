<script lang="ts">
  import { toPath, type IconName } from './icons'
  import { theme } from './theme.svelte'

  // size should be a multiple of 16 so every art pixel lands on whole screen pixels.
  let { name, size = 32 }: { name: IconName; size?: number } = $props()

  const d = $derived(toPath(theme.current?.icons[name] ?? []))
</script>

<svg width={size} height={size} viewBox="0 0 16 16" aria-hidden="true" shape-rendering="crispEdges">
  <path {d} fill="currentColor" />
</svg>

<style>
  svg {
    display: block;
  }
</style>
