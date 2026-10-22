<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from './Icon.svelte'
  import type { IconName } from './icons'

  // The centered "icon, word, hint" screen shared by empty and sealed pages.
  let { icon, title, hint, children }: { icon: IconName; title: string; hint?: string; children?: Snippet } = $props()
</script>

<div class="placard">
  <span class="icon"><Icon name={icon} size={64} /></span>
  <p class="title">{title}</p>
  {#if hint}<p class="hint">{hint}</p>{/if}
  {@render children?.()}
</div>

<style>
  .placard {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    text-align: center;
  }
  .icon {
    color: var(--dim);
    margin-bottom: 20px;
  }
  p {
    margin: 0;
  }
  .hint {
    color: var(--dim);
    font-size: 12px;
  }
</style>
