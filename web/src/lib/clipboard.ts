let timer: ReturnType<typeof setTimeout> | undefined
let dirty = false

/** Copies a value and clears the clipboard after `seconds`. */
export async function copy(value: string, seconds: number): Promise<void> {
  await navigator.clipboard.writeText(value)
  dirty = true
  clearTimeout(timer)
  timer = setTimeout(clearNow, seconds * 1000)
}

/**
 * Overwrites the clipboard if we put something there. Browsers refuse while the
 * tab is in the background, so in that case it retries when the tab regains focus.
 */
export async function clearNow(): Promise<void> {
  clearTimeout(timer)
  if (!dirty) return
  try {
    await navigator.clipboard.writeText('')
    dirty = false
  } catch {
    window.addEventListener('focus', () => void clearNow(), { once: true })
  }
}
