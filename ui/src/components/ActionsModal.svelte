<div class="p-4 h-full">
  {#if element}
    <div class="flex flex-col justify-between h-full">
      {#if element.isDir}
        <div class="text-2xl truncate mb-4 text-accent-300">
          <i class="fa fa-folder-open-o fa-fw" aria-hidden="true"></i>
          <span class="ml-1" title="{element.name}" aria-label="{element.name}">{element.name}</span>
        </div>
      {:else}
        <InfoBox {element} />
      {/if}
      {#if actions && actions.length}
        <div class="mt-8 flex items-center justify-around">
          {#each actions as {label, icon, action, isAlert}}
            <button type="button"
              class="w-2/5 p-2 flex-none bg-shade-neutral shadow {isAlert ? 'text-alert' : 'text-accent-200'} hover:bg-shade-100"
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
</div>

<script>
// Components
import InfoBox from './InfoBox.svelte'

// Props
export let element = null
export let actions = {}
</script>
