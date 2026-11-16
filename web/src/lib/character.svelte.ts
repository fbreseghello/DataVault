import { api, type HollowState } from './api'

// The hollow as the sidebar and its page see it. The pages that earn deeds
// call refreshHollow after a change; a new level shows once in the footer.
export const hollow = $state<{ state: HollowState | null; news: number | null }>({ state: null, news: null })

let newsTimer: ReturnType<typeof setTimeout> | undefined

/** Fetches the character again; a level above the last one known is news. */
export async function refreshHollow() {
  try {
    const next = await api.hollow()
    const before = hollow.state?.level
    hollow.state = next
    if (before !== undefined && next.level > before) {
      hollow.news = next.level
      clearTimeout(newsTimer)
      newsTimer = setTimeout(() => (hollow.news = null), 6000)
    }
  } catch {
    // a locked vault has no hollow to show; the next refresh tries again
  }
}

/** Marks the level as seen, once the page shows it. */
export async function seeHollow() {
  try {
    hollow.state = await api.hollowSeen()
  } catch {
    // shown again next time
  }
}

/** True while there is a level the page has not shown yet; level 1 is where every hollow starts. */
export function unseen(): boolean {
  return !!hollow.state && hollow.state.level > Math.max(hollow.state.seen_level, 1)
}
