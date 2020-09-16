{#await requesting}
    Requesting…
{:then list}
    {#each Object.keys(list) as k}
        <div class="mb-6 px-4 py-2 flex flex-row items-center cursor-pointer rounded shadow bg-shade-neutral hover:bg-shade-100 list-item">
            <div class="flex-grow flex flex-row items-center" on:click={() => selectItem(k)}>
                <div class="flex-grow-0 pr-4">
                    <i class="fa fa-chevron-right" aria-hidden="true"></i>
                </div>
                <div class="flex-grow">
                    <p class="font-bold text-lg md:text-xl text-accent-300">{k}</p>
                    <p>{list[k].type} <span class="text-text-200">::</span> {list[k].account}</p>
                </div>
            </div>
            <div class="extra flex-grow-0 pl-4" on:click={() => expandItem(k)}>
                <i class="fa fa-ellipsis-v fa-fw" aria-hidden="true" title="Details"></i>
                <span class="sr-only">Expand details</span>
            </div>
        </div>
    {:else}
        Nothing here
    {/each}
{:catch err}
    {err}
{/await}

<style>
.list-item .extra {
    display: none;
}

.list-item:hover .extra {
    display: block;
}
</style>

<script>
import {Request} from "../../shared/lib/request"

let requesting = getList()
function getList() {
    return Request('/api/connection')
}

function selectItem(name) {
    return Request('/api/repo/select', {
        method: 'POST',
        postData: {
            name
        }
    })
}

function expandItem(name) {
    alert('expand ' + name)
}
</script>
