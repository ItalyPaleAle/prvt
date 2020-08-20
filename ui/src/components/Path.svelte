<h1 title={'/' + path} aria-label={'/' + path}
  class="font-bold break-normal text-text-300 px-2 text-xl md:text-2xl flex flex-row content-start flex-wrap space-x-2 break-all">
  <a class="text-accent-300" href="#/tree/">::</a>
  {#each parts as part, i}
    <span>/</span>
    {#if i == (parts.length - 1)}
      <!-- Last part does not get a link, also because it could be a file name -->
      <span>{part}</span>
    {:else}
      <a class="text-accent-300" href={'#/tree/' + parts.slice(0, i + 1).join('/')} style="min-width: 0%;">{part}</a>
    {/if}
  {/each}
</h1>

<script>
export let path

// Split the path into various parts
$: parts = path.split('/')
</script>
