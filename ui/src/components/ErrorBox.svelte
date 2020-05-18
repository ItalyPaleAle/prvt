<div class="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 mb-4 text-sm relative" role="alert">
    <p class="font-bold text-lg mb-1">{title || 'Errors'}</p>
    <p>{message || 'An unspecified error occurreds'}</p>
    <span class="absolute top-0 bottom-0 right-0 mx-2 my-2 px-2 py-2 text-xl text-red-500 cursor-pointer" on:click={() => dispatch('close')} title="Close this box">
        <i class="fa fa-times" aria-hidden="true"></i>
        <span class="sr-only">Close this box</span>
    </span>
</div>

<script>
// Props for the view
export let title = ''
export let message = ''

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()
</script>