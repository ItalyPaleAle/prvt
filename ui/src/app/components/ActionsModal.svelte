{#if element}
  <div class="flex flex-col justify-between h-full">
    {#if element.isDir}
      <FolderInfoBox {element} {path} />
    {:else}
      <FileInfoBox {element} {path} />
    {/if}
    {#if actions && actions.length}
      <div class="mt-8 flex items-center justify-around flex-wrap">
        {#each actions as {label, icon, action, isAlert}}
          <button type="button"
            class="w-11/12 sm:w-2/5 p-2 my-2 flex-grow-0 bg-shade-neutral shadow {isAlert ? 'text-alert' : 'text-accent-200'} hover:bg-shade-100"
            on:click={() => action(element)}>
              <i class="fa {icon || ''} fa-fw" aria-hidden="true"></i>
              {label}
            </button>
        {/each}
      </div>
    {/if}
  </div>
{:else}
  No item selected
{/if}

<script>
// Components
import FileInfoBox from './FileInfoBox.svelte'
import FolderInfoBox from './FolderInfoBox.svelte'

// Props
export let element = null
export let actions = {}
export let path = ''
</script>
