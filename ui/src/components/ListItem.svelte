<li class="flex flex-row my-1 leading-normal rounded shadow bg-white w-full max-w-full hover:bg-gray-100">
    <a class="flex-grow p-3" href="{link}">
        <span class="flex flex-row">
            <span class="flex-grow-0">
                <i class="fa {icon || ''} fa-fw" aria-hidden="true"></i>
            </span>
            <span class="flex-grow truncate w-0">
                {label}
            </span>
            {#if date}
                <span class="flex-grow-0 text-gray-500 text-xs ml-2" title="{format(date, 'PPpp')}">
                    <!-- Replace with non-breking spaces -->
                    {formatDistanceToNow(date).replace(' ', ' ')} ago
                </span>
            {/if}
        </span>
    </a>
    {#if actions && actions.length}
        <span class="flex-grow-0 p-3 cursor-pointer bg-gray-100 text-gray-500" class:text-gray-800={expandActions} on:click={actionsMenuClick}>
            <i class="fa fa-ellipsis-v fa-fw" aria-hidden="true"></i>
            {#if expandActions}
                <div class="absolute text-sm">
                    <ul class="py-1 my-2 mx-2 w-48 bg-white rounded shadow">
                        {#each actions as action}
                            <li class="block px-4 py-2 text-gray-800 hover:bg-gray-200" on:click={() => dispatch(action.event)}>
                                <i class="fa {action.icon || ''} fa-fw" aria-hidden="true"></i>
                                {action.label}
                            </li>
                        {/each}
                    </ul>
                </div>
            {/if}
        </span>
    {:else}
        <span class="flex-grow-0 p-3 bg-gray-100 text-gray-600">
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

// Stores
import {dropdown} from '../stores'

// State
let expandActions = false

// There can only be one actions menu open in the entire application, so we use the $dropdown store as semaphore
$: expandActions = $dropdown === link

function actionsMenuClick() {
    if (expandActions) {
        $dropdown = null
    }
    else {
        $dropdown = link
    }
}

// Event dispatcher
import {createEventDispatcher} from 'svelte'
const dispatch = createEventDispatcher()
</script>