<script lang="ts">
  import Icon from './Icon.svelte'
  import type { IconName } from './icons'

  let { icon }: { icon: IconName } = $props()
</script>

<div class="sealed">
  <span class="icon"><Icon name={icon} size={64} /></span>
  <p>sealed</p>
  <p class="dim">opens in a future version</p>
</div>

<style>
  .sealed {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }
  .icon {
    color: var(--dim);
    margin-bottom: 20px;
  }
  p {
    margin: 0;
  }
  .dim {
    color: var(--dim);
    font-size: 12px;
  }
</style>
