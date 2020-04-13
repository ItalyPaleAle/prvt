{#await requesting}
    <i class="fa fa-spinner fa-spin fa-fw" aria-hidden="true"></i>
    Loading…
{:then list}
    {#if operationResult}
        <OperationResult
            title={operationResult.title}
            message={operationResult.message}
            list={operationResult.list}
            on:close={() => operationResult = null}
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
                    actions={[{label: 'Delete folder', event: 'delete', icon: 'fa-trash'}]}
                    on:delete={deleteTree(el.path, true)}
                />
            {:else if el.fileId}
                <ListItem
                    label={el.path}
                    icon="fa-file-o"
                    link="/file/{el.fileId}"
                    actions={[{label: 'Delete file', event: 'delete', icon: 'fa-trash'}]}
                    on:delete={deleteTree(el.path)}
                />
            {/if}
        {/each}
    </ul>
{:catch err}
    Error: {err}
{/await}

<script>
// Components
import OperationResult from './OperationResult.svelte'
import ListItem from './ListItem.svelte'

// Props for the view
// Path is the path to list
export let path = ''

// Operation result object
let operationResult = null

// "Level up" link
let levelUp = null

// Promise requesting the list of files
let requesting
$: {
    // Clean the path
    path = path || ''
    if (path.charAt(0) == '/') {
        path = path.slice(1)
    }
    if (path.charAt(path.length) == '/') {
        path = path.slice(0, -1)
    }

    // If the path isn't empty, we can go one level up
    levelUp = null
    if (path != '') {
        const pos = path.lastIndexOf('/')
        levelUp = (pos > 0) ? path.slice(0, pos) : ''
    }

    // Request the tree
    requesting = requestTree(path)
    
    // Reset operation result object
    operationResult = null
}

function requestTree(reqPath) {
    // Request the tree
    return fetch('/api/tree/' + encodeURIComponent(reqPath))
        // Get response as JSON
        .then((resp) => {
            return resp.json()
        })
        .then((list) => {
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

function deleteTree(element, isDir) {
    const reqPath = path + '/' + element

    // First, ask for confirmation
    const confirmMessage = isDir ? 'Are you sure you want to delete the folder "/' + reqPath + '" and ALL of its content? This is irreversible' : 'Are you sure you want to delete the file "/' + reqPath + '"? This is irreversible.'
    if (!confirm(confirmMessage)) {
        return
    }

    // Sets "requesting" to a promise that does a sequence of operations
    requesting = Promise.resolve()
        // Submit the request
        .then(() => fetch('/api/tree/' + encodeURIComponent(reqPath + (isDir ? '/*' : '')), {
            method: 'DELETE'
        }))
        // Check the response
        .then((resp) => {
            if (resp.status != 200) {
                throw Error('Invalid response status code')
            }

            return resp.json()
        })
        .then((list) => {
            if (!list || !Array.isArray(list) || !list.length) {
                throw Error('Invalid response')
            }

            operationResult = {
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