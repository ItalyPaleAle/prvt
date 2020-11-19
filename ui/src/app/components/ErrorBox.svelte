<div class="bg-error-100 border-l-4 border-error-200 text-error-300 p-4 mb-4 text-sm relative" role="alert">
    <p class="font-bold text-lg mb-1">{title || 'Error'}</p>
    <p>{message || 'An unspecified error occurreds'}</p>
    {#if !noClose}
        <span class="absolute top-0 bottom-0 right-0 mx-2 my-2 px-2 py-2 text-xl text-error-200 cursor-pointer" on:click={() => dispatch('close')} title="Close this box">
            <i class="fa fa-times" aria-hidden="true"></i>
            <span class="sr-only">Close this box</span>
        </span>
    {/if}
</div>

<script>
// Props for the view
export let title = ''
export let message = ''
export let noClose = false

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()
</script>
