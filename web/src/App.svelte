<script lang="ts">
  import { onMount } from 'svelte'
  import { api, LOCKED_EVENT, type SessionInfo } from './lib/api'
  import { clearNow } from './lib/clipboard'
  import Ended from './lib/Ended.svelte'
  import Login from './lib/Login.svelte'
  import { isRestShortcut, requestRest } from './lib/restscreen.svelte'
  import Shell from './lib/Shell.svelte'

  let info = $state<SessionInfo | null>(null)
  let ended = $state(false)
  // Set when the session was ended from this tab: from then on nothing may
  // bring the login back, not even a poll that slips in before the shutdown.
  let closed = false

  async function refresh() {
    if (closed) return
    try {
      const next = await api.session()
      if (closed) return
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

  // ctrl+shift+L: lock the vault if it is open and show the rest screen.
  async function onkeydown(e: KeyboardEvent) {
    if (!isRestShortcut(e) || ended || !info) return
    e.preventDefault()
    if (info.unlocked) await lock()
    requestRest()
  }

  async function endSession() {
    closed = true
    await api.endSession().catch(() => {})
    void clearNow()
    ended = true
  }

  async function lock() {
    await api.lock().catch(() => {})
    void clearNow()
    await refresh()
  }
</script>

<svelte:window {onkeydown} />

{#if ended}
  <Ended />
{:else if info?.unlocked}
  <Shell {info} onlock={lock} onrefresh={refresh} onend={endSession} />
{:else if info}
  <Login {info} onunlock={refresh} />
{/if}
