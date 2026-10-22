<script lang="ts">
  import pillar from './pillar.txt?raw'

  // The idle screen of the login: a broken pillar on black, nothing moving.
  // Any click or key wakes the login; the key itself is swallowed so it does
  // not end up in the keyword.
  let { onwake }: { onwake: () => void } = $props()

  function onkeydown(e: KeyboardEvent) {
    e.preventDefault()
    onwake()
  }
</script>

<svelte:window {onkeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="rest" onclick={onwake}>
  <pre aria-hidden="true">{pillar}</pre>
  <span class="sr-only">DataVault is resting. Click or press a key to continue.</span>
</div>

<style>
  .rest {
    position: fixed;
    inset: 0;
    background: var(--bg);
    cursor: default;
  }

  pre {
    position: absolute;
    right: 8vw;
    bottom: 8vh;
    margin: 0;
    color: var(--muted);
    font-size: clamp(10px, 1.6vh, 15px);
    line-height: 1.15;
    user-select: none;
  }
</style>
