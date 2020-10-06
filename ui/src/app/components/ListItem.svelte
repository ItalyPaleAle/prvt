<li class="list-item flex flex-row my-1 rounded shadow bg-shade-neutral w-full max-w-full hover:bg-shade-100">
    <a class="flex-grow p-3" href="{link}">
        <span class="flex flex-row">
            <span class="flex-grow-0">
                <i class="fa {icon || ''} fa-fw mr-1" aria-hidden="true"></i>
            </span>
            <span class="flex-grow truncate w-0">
                {label}
            </span>
            {#if date}
                <span class="extra flex-grow-0 text-text-100 text-xs ml-2" title="{format(date, 'PPpp')}">
                    <!-- Replace spaces with non-breking spaces -->
                    {formatDistanceToNow(date).replace(' ', ' ')} ago
                </span>
            {/if}
        </span>
    </a>
    {#if actions}
        <span class="extra flex-grow-0 p-3 cursor-pointer" on:click={() => dispatch('actions')}>
            <i class="fa fa-ellipsis-v fa-fw" aria-hidden="true" title="Details"></i>
            <span class="sr-only">Expand details</span>
        </span>
    {/if}
</li>

<style>
.list-item .extra {
    display: none;
}

.list-item:hover .extra {
    display: block;
}
</style>

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
