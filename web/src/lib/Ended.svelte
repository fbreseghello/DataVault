<script lang="ts">
  import Icon from './Icon.svelte'

  // What is left once the session is over: ended from the tombstone, by
  // `dtvlt stop`, or because the process went away. Nothing here can reach the
  // vault again; a new session starts from the terminal.
</script>

<main>
  <Icon name="tomb" size={64} />
  <p>session ended</p>
  <p class="dim">the vault is locked · run <span class="cmd">dtvlt</span> to start a new one</p>
</main>

<style>
  main {
    min-height: 100vh;
    display: grid;
    place-content: center;
    justify-items: center;
    text-align: center;
    gap: 4px;
    color: var(--fg);
  }
  main :global(svg) {
    color: var(--dim);
    margin-bottom: 20px;
  }
  p {
    margin: 0;
  }
  .dim {
    color: var(--dim);
  }
  .cmd {
    color: var(--muted);
  }
</style>
