<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from './Icon.svelte'
  import type { IconName } from './icons'

  // The "icon, word, hint" screen shown when a page has nothing open. The icon's
  // centre is pinned to the same point of the content pane on every page, and
  // the text hangs below it, so changing pages never moves the icon.
  let { icon, title, hint, children }: { icon: IconName; title: string; hint?: string; children?: Snippet } = $props()
</script>

<div class="placard">
  <span class="icon"><Icon name={icon} size={64} /></span>
  <div class="text">
    <p class="title">{title}</p>
    {#if hint}<p class="hint">{hint}</p>{/if}
    {@render children?.()}
  </div>
</div>

<style>
  /* Positioned against the page's content pane (the nearest positioned parent). */
  .placard {
    position: absolute;
    inset: 0;
    text-align: center;
  }

  .icon {
    position: absolute;
    left: 50%;
    top: 40%;
    transform: translate(-50%, -50%);
    color: var(--dim);
  }

  .text {
    position: absolute;
    left: 0;
    right: 0;
    top: calc(40% + 52px);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  p {
    margin: 0;
  }

  .hint {
    color: var(--dim);
    font-size: 12px;
  }
</style>
