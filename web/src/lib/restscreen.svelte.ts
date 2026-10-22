// A request for the login to show its rest screen now. The shortcut sets it
// from anywhere; the login picks it up when it is (or becomes) visible.
export const restRequest = $state({ pending: false })

export function requestRest() {
  restRequest.pending = true
}

/** The rest shortcut: ctrl+shift+L (cmd+shift+L on macOS). */
export function isRestShortcut(e: KeyboardEvent): boolean {
  return (e.ctrlKey || e.metaKey) && e.shiftKey && !e.altKey && e.key.toLowerCase() === 'l'
}
