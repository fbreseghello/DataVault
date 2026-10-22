<script lang="ts">
  import type { IconName } from './icons'
  import Placard from './Placard.svelte'

  // A page that has its place in the sidebar but nothing in it yet. It keeps
  // the frame of the other pages (toolbar, list column, pane) so its icon sits
  // at exactly the same point.
  let { icon, title }: { icon: IconName; title: string } = $props()
</script>

<div class="vacant">
  <div class="page-toolbar"></div>
  <div class="body">
    <div class="list"></div>
    <!-- position: relative anchors the front page icon (Placard) to this pane. -->
    <section>
      <Placard {icon} {title} hint="nothing here yet" />
    </section>
  </div>
</div>

<style>
  .vacant {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr;
    min-height: 0;
  }

  /* Same columns as the Grimoire and Unspoken. */
  .body {
    display: grid;
    grid-template-columns: minmax(240px, 360px) 1fr;
    min-height: 0;
  }

  .list {
    border-right: 1px solid var(--line);
  }

  section {
    position: relative;
    min-height: 0;
  }

  @media (max-width: 720px) {
    .body {
      grid-template-columns: 1fr;
    }
    .list {
      display: none;
    }
  }
</style>
