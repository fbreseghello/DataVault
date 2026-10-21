<script lang="ts">
  import { iconPaths, type IconName } from './icons'

  // size should be a multiple of 16 so every art pixel lands on whole screen pixels.
  let { name, size = 32 }: { name: IconName; size?: number } = $props()
</script>

<svg width={size} height={size} viewBox="0 0 16 16" aria-hidden="true" shape-rendering="crispEdges">
  <path d={iconPaths[name]} fill="currentColor" />
</svg>

<style>
  svg {
    display: block;
  }
</style>
