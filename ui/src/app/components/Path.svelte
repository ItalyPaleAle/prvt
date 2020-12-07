<h1 title={'/' + path} aria-label={'/' + path}
  class="font-bold break-normal text-text-300 px-2 text-2xl md:text-3xl flex flex-row content-start flex-wrap space-x-2 break-all">
  {#if !parts || !parts.length}
    <!-- We're in the root folder -->
    <span class="text-accent-300">::</span>
    <span class="text-text-200">/</span>
  {:else}
    <a href="#/tree/">::</a>
    {#each parts as part, i}
      <span class="text-text-200">/</span>
      {#if i == (parts.length - 1)}
        <!-- Last part does not get a link, also because it could be a file name -->
        <span class="text-accent-300">{part}</span>
      {:else}
        <a href={'#/tree/' + parts.slice(0, i + 1).join('/')} style="min-width: 0%;">{part}</a>
      {/if}
    {/each}
  {/if}
</h1>

<script lang="ts">
export let path: string

// Split the path into various parts
$: parts = path?.split('/') || []
</script>
