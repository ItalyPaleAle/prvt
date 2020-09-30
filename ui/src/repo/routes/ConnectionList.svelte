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
    {/each}
    <div
        on:click={showAddItem}
        class="mb-6 px-4 py-2 flex flex-row items-center cursor-pointer rounded shadow bg-shade-neutral hover:bg-shade-100 list-item">
        <div class="flex-grow-0 pr-4">
            <i class="fa fa-plus-circle" aria-hidden="true"></i>
        </div>
        <div class="flex-grow">
            <p class="font-bold text-lg md:text-xl text-accent-300">New connection</p>
        </div>
    </div>
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
// Libraries
import {Request} from '../../shared/lib/request'
import {push} from 'svelte-spa-router'

// Components
import ConnectionAddModal from '../components/ConnectionAddModal.svelte'
import ConnectionDetailModal from '../components/ConnectionDetailModal.svelte'

// Stores
import {modal} from '../../shared/stores'

let requesting = null
getList()
function getList() {
    requesting = Request('/api/connection')
}

// Select the item on click
function selectItem(name) {
    const postData = {name}
    requesting = Request('/api/repo/select', {postData})
        .then((data) => {
            if (data && data.gpgUnlock) {
                push('/unlock?gpg=1')
            }
            else {
                push('/unlock')
            }
        })
}

// Open the modal to add new items
function showAddItem() {
    $modal = {
        component: ConnectionAddModal,
        props: {
            add: addItem
        }
    }
}

// Open the modal on click on the dots
function expandItem(name) {
    $modal = {
        component: ConnectionDetailModal,
        props: {
            name,
            remove: removeItem
        }
    }
}

// Adds an item to the list
function addItem(data) {
    // Close the modal
    $modal = null

    // Sets "requesting" to a promise that does a sequence of operations
    requesting = Promise.resolve()
        // Submit the request
        .then(() => Request('/api/connection', {postData: data}))
        // Catch errors
        .catch((err) => {
            alert('Could not save the connection: ' + err)
        })
        // Refresh the list of connections regardless of errors
        .then(() => getList())
}

// Remove an item from the list - this is fired by an event
function removeItem(name) {
    // First, ask for confirmation
    if (!confirm('Are you sure you want to remove this connection?\nThis will only remove the bookmark and will not delete the repository or any file in it.')) {
        return
    }

    // Close the modal
    $modal = null

    // Sets "requesting" to a promise that does a sequence of operations
    requesting = Promise.resolve()
        // Submit the request
        .then(() => Request('/api/connection/' + name, {
            method: 'DELETE'
        }))
        // Catch errors
        .catch((err) => {
            alert('Could not remove the connection: ' + err)
        })
        // Refresh the list of connections regardless of errors
        .then(() => getList())
}
</script>
