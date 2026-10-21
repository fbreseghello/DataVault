<script lang="ts">
  import { onMount } from 'svelte'
  import { api, LOCKED_EVENT, type SessionInfo } from './lib/api'
  import { clearNow } from './lib/clipboard'
  import Login from './lib/Login.svelte'
  import Vault from './lib/Vault.svelte'

  let info = $state<SessionInfo | null>(null)
  let ended = $state(false)

  async function refresh() {
    try {
      const next = await api.session()
      if (info?.unlocked && !next.unlocked) void clearNow()
      info = next
      ended = false
    } catch {
      // The process is gone or this tab never got a session cookie.
      ended = true
      void clearNow()
    }
  }

  onMount(() => {
    void refresh()
    const poll = setInterval(refresh, 10_000)
    const onLocked = () => void refresh()
    window.addEventListener(LOCKED_EVENT, onLocked)
    return () => {
      clearInterval(poll)
      window.removeEventListener(LOCKED_EVENT, onLocked)
    }
  })

  async function lock() {
    await api.lock().catch(() => {})
    void clearNow()
    await refresh()
  }
</script>

{#if ended}
  <main class="ended">
    <p>session ended</p>
    <p class="dim">run <span class="cmd">dtvlt</span> to start a new one</p>
  </main>
{:else if info?.unlocked}
  <Vault {info} onlock={lock} onrefresh={refresh} />
{:else if info}
  <Login {info} onunlock={refresh} />
{/if}

<style>
  .ended {
    min-height: 100vh;
    display: grid;
    place-content: center;
    text-align: center;
    gap: 4px;
  }
  .ended p {
    margin: 0;
  }
  .dim {
    color: var(--dim);
  }
  .cmd {
    color: var(--muted);
  }
</style>
