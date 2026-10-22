<script lang="ts">
  import type { Field, Item, ItemInput, ItemType } from './api'
  import { isMultiline, template, TYPES } from './items'

  // The parent remounts the editor per item, so reading props once is intended.
  // hasSealPin: when false, sealing asks for a new seal PIN, passed to onsave.
  let {
    item,
    hasSealPin,
    onsave,
    oncancel,
  }: {
    item: Item | null
    hasSealPin: boolean
    onsave: (input: ItemInput, newSealPin: string) => Promise<void>
    oncancel: () => void
  } = $props()

  // svelte-ignore state_referenced_locally
  const initial = item
  let type = $state<ItemType>(initial?.type ?? 'login')
  let name = $state(initial?.name ?? '')
  let fields = $state<Field[]>(initial ? initial.fields.map((f) => ({ ...f })) : template('login'))
  let tags = $state(initial?.tags.join(', ') ?? '')
  let notes = $state(initial?.notes ?? '')
  let sealed = $state(initial?.sealed ?? false)
  let sealPin = $state('')
  let sealPinConfirm = $state('')
  let shown = $state<Record<number, boolean>>({})
  let busy = $state(false)
  let error = $state('')

  function changeType(t: ItemType) {
    type = t
    // Keep what was typed into fields the new template also has.
    const old = new Map(fields.map((f) => [f.name, f.value]))
    fields = template(t).map((f) => ({ ...f, value: old.get(f.name) ?? '' }))
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    error = ''
    const newPin = sealed && !hasSealPin
    if (newPin && sealPin !== sealPinConfirm) {
      error = "the pins don't match"
      return
    }
    busy = true
    try {
      await onsave(
        {
          type,
          name,
          fields,
          tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
          notes,
          sealed,
        },
        newPin ? sealPin : '',
      )
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    } finally {
      busy = false
    }
  }
</script>

<form onsubmit={submit}>
  <h2>{initial ? 'edit' : 'new item'}</h2>

  <label class="row">
    <span>type</span>
    {#if initial}
      <span class="static">{TYPES.find((t) => t.type === type)?.label}</span>
    {:else}
      <select value={type} onchange={(e) => changeType(e.currentTarget.value as ItemType)}>
        {#each TYPES as t (t.type)}
          <option value={t.type}>{t.label}</option>
        {/each}
      </select>
    {/if}
  </label>

  <label class="row">
    <span>name</span>
    <!-- svelte-ignore a11y_autofocus -->
    <input type="text" bind:value={name} required autofocus spellcheck="false" autocomplete="off" />
  </label>

  {#each fields as field, i (i)}
    <div class="row field">
      <input class="fname" type="text" bind:value={field.name} placeholder="field" spellcheck="false" aria-label="field name" />
      <div class="fvalue">
        {#if isMultiline(field)}
          <textarea bind:value={field.value} spellcheck="false" aria-label={field.name}></textarea>
        {:else if field.secret && !shown[i]}
          <input type="password" bind:value={field.value} autocomplete="new-password" aria-label={field.name} />
        {:else}
          <input type="text" bind:value={field.value} spellcheck="false" autocomplete="off" aria-label={field.name} />
        {/if}
      </div>
      <div class="actions">
        {#if field.secret && !isMultiline(field)}
          <button type="button" onclick={() => (shown[i] = !shown[i])}>{shown[i] ? 'hide' : 'show'}</button>
        {/if}
        <button type="button" class:on={field.secret} onclick={() => (field.secret = !field.secret)}>secret</button>
        <button type="button" onclick={() => fields.splice(i, 1)} aria-label="remove field">×</button>
      </div>
    </div>
  {/each}
  <button type="button" class="add" onclick={() => fields.push({ name: '', value: '', secret: false })}>+ field</button>

  <label class="row">
    <span>tags</span>
    <input type="text" bind:value={tags} placeholder="work, infra" spellcheck="false" autocomplete="off" />
  </label>

  <label class="row top">
    <span>notes</span>
    <textarea bind:value={notes}></textarea>
  </label>

  <div class="row">
    <span>sealed</span>
    <button type="button" class="toggle" class:on={sealed} onclick={() => (sealed = !sealed)} aria-pressed={sealed}>
      {sealed ? 'on' : 'off'}
    </button>
    <span class="unit">{sealed ? 'opening it asks for the seal pin' : 'opens without a pin'}</span>
  </div>
  {#if sealed && !hasSealPin}
    <p class="note">no seal pin yet. choose one; it opens every sealed entry.</p>
    <label class="row">
      <span>seal pin</span>
      <input type="password" bind:value={sealPin} minlength="4" required autocomplete="new-password" />
      <span></span>
    </label>
    <label class="row">
      <span>confirm pin</span>
      <input type="password" bind:value={sealPinConfirm} required autocomplete="new-password" />
      <span></span>
    </label>
  {/if}

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="buttons">
    <button type="submit" class="primary" disabled={busy}>{busy ? 'saving' : 'save'}</button>
    <button type="button" onclick={oncancel}>cancel</button>
  </div>
</form>

<style>
  h2 {
    margin: 0 0 20px;
    font-size: 13px;
    font-weight: 700;
  }
  .row {
    display: grid;
    grid-template-columns: 110px 1fr auto;
    gap: 12px;
    align-items: baseline;
    margin-bottom: 10px;
  }
  .row > span:first-child {
    color: var(--dim);
  }
  .row.top {
    align-items: start;
  }
  .row.top > span:first-child {
    padding-top: 4px;
  }
  .static {
    color: var(--muted);
  }
  .fname {
    color: var(--dim);
  }
  .actions {
    display: flex;
    gap: 10px;
    font-size: 12px;
  }
  .actions .on,
  .toggle.on {
    color: var(--accent);
  }
  .toggle {
    justify-self: start;
  }
  .unit,
  .note {
    color: var(--dim);
    font-size: 12px;
  }
  .note {
    margin: 0 0 10px 122px;
  }
  .add {
    margin: 2px 0 18px 122px;
    font-size: 12px;
  }
  .error {
    color: var(--accent);
  }
  .buttons {
    display: flex;
    gap: 20px;
    margin-top: 20px;
  }
  .primary {
    color: var(--fg);
  }
</style>
