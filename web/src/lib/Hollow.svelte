<script lang="ts">
  import Placard from './Placard.svelte'
  import { word } from './theme.svelte'

  // The character that grows with the vault (docs/PERSONAGEM.md). For now the
  // page only holds its place in the sidebar.
</script>

<div class="hollow">
  <Placard icon="hollow" title={word('hollow')} hint="nothing here yet · it wakes in a future version" />
</div>

<style>
  .hollow {
    position: relative;
    height: 100%;
  }
</style>
