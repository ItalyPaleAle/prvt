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
                    icon="{fileTypeIcon(el.mimeType)}"
                    link="/file/{el.fileId}"
                    date={el.date ? new Date(el.date) : null}
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

function fileTypeIcon(mimeType) {
    // Default is file-o
    if (!mimeType) {
        return 'fa-file-o'
    }

    // Specific types
    switch (mimeType) {
        case 'application/pdf':
        case 'application/x-pdf':
            return 'fa-file-pdf-o'

        case 'application/zip':
        case 'application/x-bzip':
        case 'application/x-bzip2':
        case 'application/gzip':
        case 'application/x-tar':
        case 'application/zip':
        case 'application/x-7z-compressed':
        case 'application/vnd.rar':
            return 'fa-file-archive-o'

        case 'text/plain':
        case 'application/rtf':
            return 'fa-file-text-o'

        case 'text/html':
        case 'text/javascript':
        case 'text/css':
        case 'text/xml':
        case 'application/json':
        case 'application/php':
        case 'application/x-sh':
        case 'application/x-csh':
        case 'application/xhtml+xml':
        case 'application/xml':
        case 'application/x-freearc':
            return 'fa-file-code-o'

        case 'application/vnd.ms-powerpoint':
        case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
        case 'application/vnd.oasis.opendocument.presentation':
            return 'fa-file-powerpoint-o'

        case 'text/csv':
        case 'application/vnd.ms-excel':
        case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
        case 'application/vnd.oasis.opendocument.spreadsheet':
            return 'fa-file-excel-o'

        case 'application/msword':
        case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
        case 'application/vnd.oasis.opendocument.text':
            return 'fa-file-word-o'

        default:
            if (mimeType.startsWith('image/')) {
                // All image types
                return 'fa-file-image-o'
            }
            else if (mimeType.startsWith('audio/')) {
                // All audio types
                return 'fa-file-audio-o'
            }
            else if (mimeType.startsWith('video/')) {
                // All video types
                return 'fa-file-video-o'
            }
            return 'fa-file-o'
    }
}
</script>