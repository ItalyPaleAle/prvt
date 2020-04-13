<div class="bg-green-100 border-l-4 border-green-500 text-green-700 p-4 mb-4 text-sm" role="alert">
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
</div>

<script>
export let title = ''
export let message = ''
export let list = []

let showDetails = false

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