<script lang="ts">
  import { untrack } from 'svelte'
  import { api, ApiError, type Bill, type BillInput, type Cycle } from './api'
  import { amountText, CYCLES, money, nextCharge, parseAmount, perMonth, SOON_DAYS, todayISO, totals } from './ledger'
  import { daysUntil, longDue, showDate } from './renew'
  import Placard from './Placard.svelte'

  // The ledger: subscriptions and fixed expenses. The list runs in the order the
  // bills come due; the toolbar adds up what they cost a month and a year.
  let { active }: { active: boolean } = $props()

  let bills = $state<Bill[]>([])
  let selectedId = $state<string | 'new' | null>(null)
  let error = $state('')
  let busy = $state(false)
  let confirmStrike = $state(false)

  // the form
  let name = $state('')
  let amount = $state('')
  let currency = $state('BRL')
  let cycle = $state<Cycle>('monthly')
  let since = $state('')
  let method = $state('')
  let note = $state('')

  let today = $state(todayISO())
  const ordered = $derived(
    bills
      .map((b) => ({ ...b, next: nextCharge(b, today) }))
      .sort((a, b) => a.next.localeCompare(b.next) || a.name.localeCompare(b.name)),
  )
  const selected = $derived(ordered.find((b) => b.id === selectedId) ?? null)
  const sums = $derived(totals(bills))
  const soon = $derived(ordered.filter((b) => daysUntil(b.next) <= SOON_DAYS).length)
  // what the form would cost a month, as it is typed
  const draftCents = $derived(parseAmount(amount))

  function fail(e: unknown) {
    if (e instanceof ApiError && e.status === 423) return
    error = e instanceof Error ? e.message : String(e)
  }

  async function load() {
    bills = await api.bills()
  }

  $effect(() => {
    if (!active) return
    untrack(() => {
      today = todayISO()
      close()
      load().catch(fail)
    })
  })

  // the currency most bills use, for a new one
  function usualCurrency(): string {
    const n = new Map<string, number>()
    for (const b of bills) n.set(b.currency, (n.get(b.currency) ?? 0) + 1)
    return [...n].sort((a, b) => b[1] - a[1])[0]?.[0] ?? 'BRL'
  }

  function open(b: Bill | null) {
    error = ''
    confirmStrike = false
    selectedId = b ? b.id : 'new'
    name = b?.name ?? ''
    amount = b ? amountText(b.amount) : ''
    currency = b?.currency ?? usualCurrency()
    cycle = b?.cycle ?? 'monthly'
    since = b?.since ?? todayISO()
    method = b?.method ?? ''
    note = b?.note ?? ''
  }

  function close() {
    selectedId = null
    confirmStrike = false
    error = ''
  }

  async function save(e: SubmitEvent) {
    e.preventDefault()
    const cents = parseAmount(amount)
    if (cents === null) {
      error = 'the amount is not a number'
      return
    }
    const input: BillInput = { name, amount: cents, currency, cycle, since, method, note }
    busy = true
    error = ''
    try {
      const b = selectedId && selectedId !== 'new' ? await api.updateBill(selectedId, input) : await api.createBill(input)
      await load()
      open(b)
    } catch (err) {
      fail(err)
    } finally {
      busy = false
    }
  }

  async function strike() {
    if (!selected) return
    if (!confirmStrike) {
      confirmStrike = true
      return
    }
    try {
      await api.removeBill(selected.id)
      close()
      await load()
    } catch (e) {
      fail(e)
    }
  }

  function per(c: Cycle): string {
    return CYCLES.find((x) => x.cycle === c)?.short ?? c
  }

  function shortNext(next: string): string {
    const n = daysUntil(next)
    if (n === 0) return 'today'
    if (n === 1) return 'tomorrow'
    if (n <= SOON_DAYS) return `in ${n}d`
    return showDate(next).replace(/ \d{4}$/, '')
  }

  function isTyping(t: EventTarget | null) {
    return t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || t instanceof HTMLSelectElement
  }

  function onkeydown(e: KeyboardEvent) {
    if (!active) return
    if (e.key === 'Escape') {
      close()
      return
    }
    if (isTyping(e.target) || e.ctrlKey || e.metaKey || e.altKey) return
    if (e.key === 'n') {
      e.preventDefault()
      open(null)
    }
  }
</script>

<svelte:window {onkeydown} />

<div class="ledger">
  <div class="toolbar page-toolbar">
    <span class="sums">
      {#each sums as s, i (s.currency)}
        {#if i}<span class="dim">{' · '}</span>{/if}
        <span>{money(Math.round(s.month), s.currency)}</span>
        <span class="dim">{`a month · ${money(Math.round(s.month * 12), s.currency)} a year`}</span>
      {/each}
    </span>
    <span class="spacer"></span>
    {#if soon}<span class="count due">{soon} due this week</span>{/if}
    <span class="count">{bills.length} {bills.length === 1 ? 'bill' : 'bills'}</span>
    <button class="new" onclick={() => open(null)}>+ new</button>
  </div>

  <div class="body">
    <nav aria-label="bills">
      {#each ordered as b (b.id)}
        <button class="bill" class:active={selectedId === b.id} onclick={() => open(b)}>
          <span class="name">{b.name}</span>
          <span class="amount">{money(b.amount, b.currency)}<span class="dim">/{per(b.cycle)}</span></span>
          <span class="next" class:due={daysUntil(b.next) <= SOON_DAYS}>{shortNext(b.next)}</span>
        </button>
      {:else}
        <p class="empty">nothing owed yet · press n to write the first bill</p>
      {/each}
    </nav>

    <section>
      {#if selectedId}
        <form class="page" onsubmit={save}>
          <h2>
            {selected ? selected.name : 'new bill'}
            {#if selected}
              <span class="dim">next {showDate(selected.next)} · {longDue(selected.next)}</span>
            {/if}
          </h2>

          <label class="row">
            <span>name</span>
            <!-- svelte-ignore a11y_autofocus -->
            <input type="text" bind:value={name} required autofocus={!selected} spellcheck="false" autocomplete="off" placeholder="streaming, server, gym" />
            <span></span>
          </label>
          <label class="row">
            <span>amount</span>
            <input type="text" inputmode="decimal" bind:value={amount} required autocomplete="off" placeholder="55,90" />
            <input class="currency" type="text" bind:value={currency} required maxlength="3" spellcheck="false" autocomplete="off" aria-label="currency" />
          </label>
          <label class="row">
            <span>every</span>
            <select bind:value={cycle}>
              {#each CYCLES as c (c.cycle)}
                <option value={c.cycle}>{c.label}</option>
              {/each}
            </select>
            <span class="unit">
              {#if cycle !== 'monthly' && draftCents !== null && currency.length === 3}
                ≈ {money(Math.round(perMonth({ amount: draftCents, cycle })), currency.toUpperCase())} a month
              {/if}
            </span>
          </label>
          <label class="row">
            <span>first charge</span>
            <input type="date" bind:value={since} required />
            <span class="unit">the rest follow</span>
          </label>
          <label class="row">
            <span>paid with</span>
            <input type="text" bind:value={method} spellcheck="false" autocomplete="off" placeholder="card or account" />
            <span></span>
          </label>
          <label class="row top">
            <span>note</span>
            <textarea bind:value={note}></textarea>
            <span></span>
          </label>

          {#if error}<p class="error" role="alert">{error}</p>{/if}

          <div class="buttons">
            <button type="submit" class="primary" disabled={busy}>{busy ? 'saving' : 'save'}</button>
            <button type="button" onclick={close}>close</button>
            {#if selected}
              <button type="button" class="strike" class:danger={confirmStrike} onclick={strike}>
                {confirmStrike ? 'strike it? click again' : 'strike'}
              </button>
            {/if}
          </div>
          <p class="meta">kept encrypted in the vault · strike a bill once it is cancelled</p>
        </form>
      {:else}
        {#if error}<p class="error" role="alert">{error}</p>{/if}
        <Placard
          icon="sack"
          title="ledger"
          hint={bills.length ? 'select a bill · n to write a new one' : 'subscriptions and fixed expenses · n to write the first'}
        />
      {/if}
    </section>
  </div>
</div>

<style>
  .ledger {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  .dim {
    color: var(--dim);
  }

  .sums {
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }

  .page-toolbar .count.due,
  .next.due,
  .danger,
  .error {
    color: var(--accent);
  }

  .body {
    display: grid;
    grid-template-columns: minmax(240px, 360px) 1fr;
    min-height: 0;
  }

  nav {
    border-right: 1px solid var(--line);
    padding: 8px 0;
    overflow-y: auto;
  }

  .bill {
    display: grid;
    grid-template-columns: 1fr auto 64px;
    gap: 10px;
    width: 100%;
    padding: 5px 24px;
    text-align: left;
    color: var(--fg);
  }

  .bill:hover {
    background: var(--raise);
  }

  .bill.active {
    background: var(--raise);
    box-shadow: inset 2px 0 0 var(--fg);
  }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .amount {
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  .next {
    color: var(--dim);
    font-size: 11px;
    text-align: right;
    align-self: center;
  }

  .empty {
    color: var(--dim);
    padding: 0 24px;
  }

  section {
    position: relative;
    padding: 16px 32px;
    overflow-y: auto;
    min-height: 0;
  }

  .page {
    max-width: 640px;
  }

  h2 {
    margin: 0 0 20px;
    font-size: 15px;
    font-weight: 700;
    overflow-wrap: anywhere;
  }

  h2 .dim {
    font-size: 12px;
    font-weight: 400;
    margin-left: 10px;
  }

  .row {
    display: grid;
    grid-template-columns: 110px 1fr 140px;
    gap: 12px;
    align-items: baseline;
    margin-bottom: 10px;
  }

  .row > span:first-child,
  .unit {
    color: var(--dim);
  }

  .unit {
    font-size: 12px;
  }

  .row.top {
    align-items: start;
  }

  .row.top > span:first-child {
    padding-top: 4px;
  }

  input[type='date'] {
    width: 100%;
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--line);
    padding: 4px 0;
    outline: none;
  }

  input[type='date']:focus {
    border-bottom-color: var(--muted);
  }

  input::placeholder {
    color: var(--dim);
  }

  .currency {
    text-transform: uppercase;
    width: 4ch;
  }

  .buttons {
    display: flex;
    gap: 20px;
    margin-top: 24px;
  }

  .strike {
    margin-left: auto;
  }

  .primary {
    color: var(--fg);
  }

  .meta {
    margin: 16px 0 0;
    color: var(--dim);
    font-size: 12px;
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
    }
    nav {
      border-right: 0;
      border-bottom: 1px solid var(--line);
    }
  }
</style>
