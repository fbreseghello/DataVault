import type { Theme, WordKey } from './api'

// The theme every screen draws. The session hands it over with every poll,
// already filled in from the default theme, so nothing here guesses.
export const theme = $state<{ current: Theme | null }>({ current: null })

let saved: Theme | null = null
// The settings screen shows a theme before it is saved; the polls leave it be.
let preview: Theme | null = null
let drawn = ''

/** The theme saved in the settings, as the session reports it. */
export function applyTheme(t: Theme) {
  saved = t
  draw()
}

/** Shows a theme without saving it; null goes back to the saved one. */
export function previewTheme(t: Theme | null) {
  preview = t
  draw()
}

function draw() {
  const t = preview ?? saved
  // The session is polled every few seconds; only a real change redraws.
  const key = JSON.stringify(t)
  if (!t || key === drawn) return
  drawn = key
  theme.current = t
  const style = document.documentElement.style
  for (const [name, value] of Object.entries(t.colors)) style.setProperty('--' + name, value)
}

/** The current theme's name for a page or feature. */
export function word(key: WordKey): string {
  return theme.current?.words[key] ?? key
}
