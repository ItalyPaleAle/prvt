<div class="mx-3">
  <ConnectionAddOption opt={nameOpt} required={true} />
  {#each options.required as opt}
    <ConnectionAddOption {opt} required={true} />
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
        <ConnectionAddOption {opt} required={false} />
      {/each}
    </div>
  {/if}
</div>

<script>
// Components
import ConnectionAddOption from "./ConnectionAddOption.svelte"

// Props
export let options = {}

// Option for the name
const nameOpt = {
    name: 'name',
    type: 'string',
    label: 'Connection name',
    validate: '^[a-z][a-z0-9-_]{1,39}$',
    validateMessage: 'Name must only include lowercase letters, numbers, dashes and underscores; additionally, it must start with a letter and be within 2 and 40 characters'
}

let showOptional = false
</script>
