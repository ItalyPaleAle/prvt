<li class="flex flex-row my-1 leading-normal rounded shadow bg-shade-neutral w-full max-w-full hover:bg-shade-100">
    <a class="flex-grow p-3" href="{link}">
        <span class="flex flex-row">
            <span class="flex-grow-0">
                <i class="fa {icon || ''} fa-fw" aria-hidden="true"></i>
            </span>
            <span class="flex-grow truncate w-0">
                {label}
            </span>
            {#if date}
                <span class="flex-grow-0 text-text-100 text-xs ml-2" title="{format(date, 'PPpp')}">
                    <!-- Replace spaces with non-breking spaces -->
                    {formatDistanceToNow(date).replace(' ', ' ')} ago
                </span>
            {/if}
        </span>
    </a>
    {#if actions}
        <span class="flex-grow-0 p-3 cursor-pointer bg-shade-100" on:click={() => dispatch('actions')}>
            <i class="fa fa-ellipsis-v fa-fw" aria-hidden="true"></i>
        </span>
    {:else}
        <span class="flex-grow-0 p-3 bg-shade-100 text-text-200">
            <i class="fa fa-fw" aria-hidden="true"></i>
        </span>
    {/if}
</li>

<script>
import formatDistanceToNow from 'date-fns/formatDistanceToNow'
import format from 'date-fns/format'

// Props for the view
export let label = ''
export let icon = ''
export let link = ''
export let date = null
export let actions = null

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()
</script>