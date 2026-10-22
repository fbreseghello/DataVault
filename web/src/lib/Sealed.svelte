<script lang="ts">
  import type { IconName } from './icons'
  import Placard from './Placard.svelte'

  let { icon }: { icon: IconName } = $props()
</script>

<Placard {icon} title="sealed" hint="opens in a future version" />
