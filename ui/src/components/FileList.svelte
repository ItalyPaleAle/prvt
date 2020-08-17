{#await requesting}
    <i class="fa fa-spinner fa-spin fa-fw" aria-hidden="true"></i>
    Loading…
{:then list}
    {#if $operationResult}
        <OperationResult
            title={$operationResult.title}
            message={$operationResult.message}
            list={$operationResult.list}
            on:close={() => $operationResult = null}
        />
    {/if}
    <ul>
        {#if levelUp !== null}
            <ListItem
                label="Up one level"
                icon="fa-level-up"
                link="#/tree/{levelUp}"
            />
        {/if}
        {#each list as el}
            {#if el.isDir}
                <ListItem
                    label={el.path}
                    icon="fa-folder"
                    link="#/tree/{path ? path + '/' : ''}{el.path}"
                    actions={true}
                    on:actions={() => showActions(el)}
                />
            {:else if el.fileId}
                <ListItem
                    label={el.path}
                    icon="{fileTypeIcon(el.mimeType)}"
                    link="/file/{el.fileId}"
                    date={el.date ? new Date(el.date) : null}
                    actions={true}
                    on:actions={() => showActions(el)}
                />
            {/if}
        {/each}
    </ul>
{:catch err}
    <ErrorBox message={err} />
{/await}

<script>
// Utils
import {encodePath, fileTypeIcon, cloneObject} from '../utils'

// Components
import ErrorBox from './ErrorBox.svelte'
import OperationResult from './OperationResult.svelte'
import ListItem from './ListItem.svelte'
import ActionsModal from './ActionsModal.svelte'

// Stores
import {operationResult, modal} from '../stores'

// Props for the view
// Path is the path to list
export let path = ''

// "Level up" link
let levelUp = null

// Actions presets
const actionsFolder = [
    {label: 'Delete folder', icon: 'fa-trash', action: deleteFolder, isAlert: true}
]
const actionsFile = [
    {label: 'Download', icon: 'fa-download', action: downloadFile},
    {label: 'Delete file', icon: 'fa-trash', action: deleteFile, isAlert: true}
]

// Promise requesting the list of files
let requesting

$: {
    // If the path isn't empty, we can go one level up
    levelUp = null
    if (path != '') {
        const pos = path.lastIndexOf('/')
        levelUp = (pos > 0) ? path.slice(0, pos) : ''
    }

    // Request the tree
    requesting = requestTree(path)

    // Reset operation result object unless this is the first time it's shown
    if ($operationResult && !$operationResult.shown) {
        $operationResult.shown = true
    }
    else {
        $operationResult = null
    }
}

function requestTree(reqPath) {
    // Request the tree
    return fetch('/api/tree/' + encodePath(reqPath))
        // Get response as JSON
        .then((resp) => {
            return resp.json()
        })
        .then((list) => {
            // Check if we have an error message
            if (list && list.error) {
                return Promise.reject(list.error)
            }
            // Ensure the list is valid
            if (!Array.isArray(list)) {
                return Promise.reject('Invalid response: not an array')
            }
            // Sort the list
            return list.sort((a, b) => {
                // Directories go first no matter what
                if (a.isDir != b.isDir) {
                    return a.isDir ? -1 : 1
                }
                return (a.path || '').localeCompare(b.path || '')
            })
        })
}

// Displays the actions modal
function showActions(element) {
    // The ActionsModal expects "name" to be be the set
    const el = cloneObject(element)
    el.name = el.path

    // Display the modal
    $modal = {
        component: ActionsModal,
        props: {
            element: el,
            actions: el.isDir ? actionsFolder : actionsFile
        }
    }
}

// The next functions are used in the actions presets
function downloadFile(element) {
    // Close the modal
    $modal = null

    // Trigger a file download
    location.href = '/file/' + element.fileId + '?dl=1'
}

function deleteFile(element) {
    return deleteTree(element)
}

function deleteFolder(element) {
    return deleteTree(element, true)
}

function deleteTree(element, isDir) {
    const reqPath = (path ? path + '/' : '') + element.path

    // First, ask for confirmation
    const confirmMessage = isDir ? 'Are you sure you want to delete the folder "/' + reqPath + '" and ALL of its content? This is irreversible' : 'Are you sure you want to delete the file "/' + reqPath + '"? This is irreversible.'
    if (!confirm(confirmMessage)) {
        return
    }

    // Close the modal
    $modal = null

    // Sets "requesting" to a promise that does a sequence of operations
    requesting = Promise.resolve()
        // Submit the request
        .then(() => fetch('/api/tree/' + encodePath(reqPath + (isDir ? '/*' : '')), {
            method: 'DELETE'
        }))
        // Check the response
        .then((resp) => {
            if (resp.status != 200) {
                return resp.json()
                    .catch(() => {
                        throw Error('Invalid response status code')
                    })
                    .then((body) => {
                        if (body && body.error) {
                            throw Error(body.error)
                        }
                        throw Error('Invalid response status code')
                    })
            }

            return resp.json()
        })
        .then((list) => {
            if (!list || !Array.isArray(list) || !list.length) {
                throw Error('Invalid response')
            }

            $operationResult = {
                title: 'Deleted',
                message: isDir ? 'The folder "/' + reqPath + '" has been deleted.' : 'The file "/' + reqPath + '" has been deleted.',
                list
            }
        })
        // Catch errors
        .catch((err) => {
            alert('Could not delete the element: ' + err)
        })
        // Refresh the list of files regardless of errors
        .then(() => requestTree(path))
}
</script>
