<div class="mx-3">
  {#each options.required as opt}
    <ConnectionAddOption {opt} />
  {:else}
    <p>This storage doesn't have any required option</p>
  {/each}
  {#if options.optional}
    <div class="mt-6 mb-4 cursor-pointer" on:click={() => {showOptional = !showOptional}}>
      {#if showOptional}
        <i class="fa fa-chevron-down" aria-hidden="true"></i> Hide advanced options
      {:else}
        <i class="fa fa-chevron-right" aria-hidden="true"></i> Show advanced options
      {/if}
    </div>
    <div class:hidden={!showOptional}>
      {#each options.optional as opt}
        <ConnectionAddOption {opt} />
      {/each}
    </div>
  {/if}
</div>

<script>
// Components
import ConnectionAddOption from "./ConnectionAddOption.svelte"

// Props
export let options = {}

let showOptional = false
</script>
