<div class="bg-success-100 border-l-4 border-success-200 text-success-300 p-4 mb-4 text-sm relative" role="alert">
    <p class="font-bold text-lg mb-1">{title}</p>
    <p>{message}</p>
    {#if list && list.length}
        {#if showDetails}
            <p class="text-xs my-2 cursor-pointer" on:click={() => showDetails = false}><i class="fa fa-caret-down" aria-hidden="true"></i> Hide details</p>
            <ul class="text-sm">
                {#each list as el}
                    <li>
                        <b>{formatStatus(el.status)}:</b>
                        <em>{el.path}</em>
                        {#if el.error}
                            ({el.error})
                        {/if}
                    </li>
                {/each}
            </ul>
        {:else}
            <p class="text-xs my-2 cursor-pointer" on:click={() => showDetails = true}><i class="fa fa-caret-right" aria-hidden="true"></i> Show details</p>
        {/if}
    {/if}
    <span class="absolute top-0 bottom-0 right-0 mx-2 my-2 px-2 py-2 text-xl text-success-200 cursor-pointer" on:click={() => dispatch('close')} title="Close this box">
        <i class="fa fa-times" aria-hidden="true"></i>
        <span class="sr-only">Close this box</span>
    </span>
</div>

<script>
// Props for the view
export let title = ''
export let message = ''
export let list = []

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()

// State
let showDetails = false

// Format status messages
function formatStatus(status) {
    switch (status) {
        case 'added':
            return 'Added'
        case 'removed':
            return 'Removed'
        case 'not-found':
            return 'Not found'
        case 'existing':
            return 'Skipped existing'
        case 'ignored':
            return 'Ignored'
        case 'internal-error':
            return 'Internal error'
        case 'error':
            return 'Error'
    }
}
</script>