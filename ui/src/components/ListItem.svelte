<li class="flex flex-row my-1 leading-normal rounded shadow bg-white">
    <a class="flex-grow p-3" href="{link}">
        <i class="fa {icon} fa-fw" aria-hidden="true"></i> {label}
    </a>
    {#if hasAction}
        <span class="flex-grow-0 p-3 cursor-pointer bg-gray-100 text-gray-600" on:click={() => dispatch('action')}>
            <i class="fa fa-ellipsis-v fa-fw" aria-hidden="true"></i>
        </span>
    {:else}
        <span class="flex-grow-0 p-3 bg-gray-100 text-gray-600">
            <i class="fa fa-fw" aria-hidden="true"></i>
        </span>
    {/if}
</li>

<script>
// Props for the view
export let label = ''
export let icon = ''
export let link = ''
export let hasAction = false

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()
</script>