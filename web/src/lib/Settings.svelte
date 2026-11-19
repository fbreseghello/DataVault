<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { api, ApiError, type LoginArt, type RestStyle, type Settings, type Theme, type ThemeSummary } from './api'
  import SealPinSettings from './SealPinSettings.svelte'
  import { applyTheme, previewTheme } from './theme.svelte'

  let { onsaved, onclose }: { onsaved: () => void; onclose: () => void } = $props()

  let saved = $state<Settings | null>(null)
  let pinEnabled = $state(false)
  let keepTerminal = $state(true)
  let autoLock = $state(15)
  let clipboard = $state(20)
  let checkUpdates = $state(true)
  let rest = $state(15)
  let restStyle = $state<RestStyle>('pillar')
  const restHints: Record<RestStyle, string> = {
    pillar: 'the broken pillar',
    'skull-i': 'the skull, drawn in lines',
    'skull-ii': 'the skull, drawn in dust',
    random: 'one of the three still images, drawn each time it rests',
    candle: 'a skull lit by a flickering candle, which moves',
  }
  let themeId = $state('souls')
  let loginArt = $state<LoginArt>('theme')
  const loginHints: Record<LoginArt, string> = {
    theme: "the chosen theme's own art",
    poison: 'datavault in poison, whatever the theme',
    gothic: 'datavault in gothic letters, whatever the theme',
  }
  let themes = $state<ThemeSummary[]>([])
  const chosen = $derived(themes.find((t) => t.id === themeId))
  // Picking a theme shows it at once; leaving without saving puts the saved one back.
  let previewed: Theme | null = null
  let themeMsg = $state('')
  let confirmRemove = $state(false)
  let fileInput = $state<HTMLInputElement>()
  let pin = $state('')
  let pinConfirm = $state('')
  let busy = $state(false)
  let error = $state('')
  let done = $state(false)

  // A PIN must be typed when turning it on; when it is already on, typing one changes it.
  const needsPin = $derived(pinEnabled && !saved?.pin_enabled)

  function load(s: Settings) {
    saved = s
    pinEnabled = s.pin_enabled
    keepTerminal = s.keep_terminal
    autoLock = s.auto_lock_minutes
    clipboard = s.clipboard_clear_seconds
    checkUpdates = s.check_updates
    rest = s.rest_minutes
    restStyle = s.rest_style
    themeId = s.theme
    loginArt = s.login_art
    pin = pinConfirm = ''
  }

  const fail = (e: unknown) => e instanceof Error ? e.message : String(e)

  async function loadThemes() {
    themes = await api.themes()
  }

  onMount(() => {
    api.settings().then(load, (e) => (error = fail(e)))
    loadThemes().catch((e) => (error = fail(e)))
  })

  onDestroy(() => previewTheme(null))

  async function preview(id: string) {
    themeId = id
    themeMsg = ''
    confirmRemove = false
    try {
      previewed = id === saved?.theme ? null : await api.theme(id)
      previewTheme(previewed)
    } catch (e) {
      themeMsg = fail(e)
    }
  }

  // An imported theme is a JSON file; see docs/TEMAS.md, or export one to start from.
  async function importFile() {
    const file = fileInput?.files?.[0]
    if (!file) return
    themeMsg = ''
    try {
      let parsed: unknown
      try {
        parsed = JSON.parse(await file.text())
      } catch {
        throw new Error(`${file.name} is not JSON`)
      }
      const t = await api.importTheme(parsed)
      await loadThemes()
      await preview(t.id)
      themeMsg = `imported ${t.name}, save to keep it`
    } catch (e) {
      themeMsg = fail(e)
    } finally {
      if (fileInput) fileInput.value = ''
    }
  }

  // Saves the chosen theme, every gap filled in, as a file to edit and import.
  async function exportTheme() {
    themeMsg = ''
    try {
      const t = await api.theme(themeId)
      const url = URL.createObjectURL(new Blob([JSON.stringify(t, null, 2) + '\n'], { type: 'application/json' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `${t.id}.json`
      a.click()
      URL.revokeObjectURL(url)
    } catch (e) {
      themeMsg = fail(e)
    }
  }

  async function removeTheme() {
    if (!confirmRemove) {
      confirmRemove = true
      return
    }
    confirmRemove = false
    try {
      const gone = chosen?.name ?? themeId
      await api.removeTheme(themeId)
      // The server falls back to the default when the removed theme was in use.
      load(await api.settings())
      await loadThemes()
      previewed = null
      previewTheme(null)
      onsaved()
      themeMsg = `removed ${gone}`
    } catch (e) {
      themeMsg = fail(e)
    }
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    error = ''
    done = false
    if (pinEnabled && (needsPin || pin) && pin !== pinConfirm) {
      error = "the pins don't match"
      return
    }
    busy = true
    try {
      load(
        await api.saveSettings({
          pin_enabled: pinEnabled,
          pin: pinEnabled ? pin : '',
          keep_terminal: keepTerminal,
          auto_lock_minutes: autoLock,
          clipboard_clear_seconds: clipboard,
          check_updates: checkUpdates,
          rest_minutes: rest,
          rest_style: restStyle,
          theme: themeId,
          login_art: loginArt,
        }),
      )
      // What was previewed is now saved; keep showing it until the session says so too.
      if (previewed) applyTheme(previewed)
      previewed = null
      previewTheme(null)
      done = true
      onsaved()
    } catch (err) {
      if (!(err instanceof ApiError && err.status === 423)) error = err instanceof Error ? err.message : String(err)
    } finally {
      busy = false
    }
  }
</script>

<div class="set">
  <h2 class="set-title">settings</h2>

  <form onsubmit={submit}>
    {#if saved}
      <section class="set-group" aria-labelledby="set-session">
        <header>
          <h3 id="set-session">session</h3>
          <p>applies now</p>
        </header>
        <div class="set-fields">
          <label class="set-field">
            <span class="set-label">auto-lock</span>
            <span class="set-control">
              <input class="set-num" type="number" min="1" max="1440" bind:value={autoLock} required />
              <span class="set-unit">minutes idle</span>
            </span>
          </label>
          <label class="set-field">
            <span class="set-label">clear clipboard</span>
            <span class="set-control">
              <input class="set-num" type="number" min="5" max="600" bind:value={clipboard} required />
              <span class="set-unit">seconds after copy</span>
            </span>
          </label>
        </div>
      </section>

      <section class="set-group" aria-labelledby="set-rest">
        <header>
          <h3 id="set-rest">rest screen</h3>
          <p>applies now; ctrl+shift+l rests at any time</p>
        </header>
        <div class="set-fields">
          <label class="set-field">
            <span class="set-label">rest after</span>
            <span class="set-control">
              <input class="set-num" type="number" min="0" max="240" bind:value={rest} required />
              <span class="set-unit">minutes idle on the login</span>
            </span>
            <span class="set-hint">{rest === 0 ? 'off, only ctrl+shift+l rests' : '0 turns it off'}</span>
          </label>
          <label class="set-field">
            <span class="set-label">rest style</span>
            <span class="set-control">
              <select bind:value={restStyle}>
                {#each Object.keys(restHints) as style (style)}
                  <option value={style}>{style}</option>
                {/each}
              </select>
            </span>
            <span class="set-hint">{restHints[restStyle]}</span>
          </label>
        </div>
      </section>

      <section class="set-group" aria-labelledby="set-appearance">
        <header>
          <h3 id="set-appearance">appearance</h3>
          <p>the theme changes the colors, icons, page names and login art</p>
        </header>
        <div class="set-fields">
          <div class="set-field">
            <label class="set-label" for="set-theme">theme</label>
            <span class="set-control">
              <select id="set-theme" bind:value={themeId} onchange={(e) => preview(e.currentTarget.value)} aria-describedby="theme-hint">
                {#each themes as t (t.id)}
                  <option value={t.id}>{t.name}{t.builtin ? '' : ' (imported)'}</option>
                {/each}
              </select>
            </span>
            <span class="set-hint" id="theme-hint">{chosen?.description ?? ''}</span>
            <span class="set-actions">
              <button type="button" onclick={() => fileInput?.click()}>import</button>
              <button type="button" onclick={exportTheme}>export</button>
              {#if chosen && !chosen.builtin}
                <button type="button" class:danger={confirmRemove} onclick={removeTheme} onblur={() => (confirmRemove = false)}>
                  {confirmRemove ? 'remove? click again' : 'remove'}
                </button>
              {/if}
            </span>
            <span class="set-hint" class:set-note={themeMsg} role="status">{themeMsg || 'a theme is a json file; export one to start from it'}</span>
            <input bind:this={fileInput} type="file" accept=".json,application/json" hidden onchange={importFile} />
          </div>
          <label class="set-field">
            <span class="set-label">login art</span>
            <span class="set-control">
              <select bind:value={loginArt}>
                {#each Object.keys(loginHints) as art (art)}
                  <option value={art}>{art}</option>
                {/each}
              </select>
            </span>
            <span class="set-hint">{loginHints[loginArt]}</span>
          </label>
        </div>
      </section>

      <section class="set-group" aria-labelledby="set-terminal">
        <header>
          <h3 id="set-terminal">terminal</h3>
          <p>applies the next time dtvlt starts</p>
        </header>
        <div class="set-fields">
          <div class="set-field">
            <span class="set-label">pin on start</span>
            <span class="set-control">
              <button type="button" class="set-toggle" class:on={pinEnabled} onclick={() => (pinEnabled = !pinEnabled)} aria-pressed={pinEnabled}>
                {pinEnabled ? 'on' : 'off'}
              </button>
            </span>
            <span class="set-hint">only gates starting dtvlt, it does not encrypt</span>
          </div>
          {#if pinEnabled}
            <label class="set-field">
              <span class="set-label">{needsPin ? 'pin' : 'new pin'}</span>
              <span class="set-control">
                <input type="password" bind:value={pin} minlength="4" required={needsPin} autocomplete="new-password" placeholder={needsPin ? '' : 'keep current'} />
              </span>
            </label>
            {#if needsPin || pin}
              <label class="set-field">
                <span class="set-label">confirm pin</span>
                <span class="set-control">
                  <input type="password" bind:value={pinConfirm} required autocomplete="new-password" />
                </span>
              </label>
            {/if}
          {/if}
          <div class="set-field">
            <span class="set-label">keep terminal open</span>
            <span class="set-control">
              <button type="button" class="set-toggle" class:on={keepTerminal} onclick={() => (keepTerminal = !keepTerminal)} aria-pressed={keepTerminal}>
                {keepTerminal ? 'on' : 'off'}
              </button>
            </span>
            <span class="set-hint">{keepTerminal ? 'the session logs to the terminal' : 'the session runs in the background'}</span>
          </div>
          <div class="set-field">
            <span class="set-label">check for updates</span>
            <span class="set-control">
              <button type="button" class="set-toggle" class:on={checkUpdates} onclick={() => (checkUpdates = !checkUpdates)} aria-pressed={checkUpdates}>
                {checkUpdates ? 'on' : 'off'}
              </button>
            </span>
            <span class="set-hint">on each start, asks github.com for a newer version before the pin</span>
          </div>
        </div>
      </section>
    {/if}

    <div class="set-bar">
      <span class="set-bar-note">the four groups above save together</span>
      <span class="set-actions">
        <button type="submit" class="primary" disabled={busy || !saved}>{busy ? 'saving' : 'save'}</button>
        <button type="button" onclick={onclose}>close</button>
        {#if error}<span class="set-note" role="alert">{error}</span>{/if}
        {#if done}<span class="set-note" role="status">saved</span>{/if}
      </span>
    </div>
  </form>

  <!-- A separate form: the seal PIN is stored in the vault and saved on its own. -->
  <SealPinSettings />

  {#if saved}
    <section class="set-group" aria-labelledby="set-files">
      <header>
        <h3 id="set-files">files</h3>
      </header>
      <dl class="set-fields">
        <div class="set-field">
          <dt class="set-label">config</dt>
          <dd class="set-path">{saved.config_path}</dd>
        </div>
        <div class="set-field">
          <dt class="set-label">vault</dt>
          <dd class="set-path">{saved.vault_path}</dd>
        </div>
      </dl>
    </section>
  {/if}
</div>

