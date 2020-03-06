{#await requesting}
    <i class="fa fa-spinner fa-spin fa-fw" aria-hidden="true"></i>
    Loading…
{:then list}
    {#if levelUp !== null}
        <a class="block p-3 my-1 leading-normal rounded shadow bg-white" href="#/tree/{levelUp}">
            <i class="fa fa-level-up" aria-hidden="true"></i> Up one level
        </a>
    {/if}
    {#each list as el}
        {#if el.isDir}
            <a class="block p-3 my-1 leading-normal rounded shadow bg-white" href="#/tree/{path ? path + '/' : ''}{el.path}">
                <i class="fa fa-folder" aria-hidden="true"></i> {el.path}
            </a>
        {:else if el.fileId}
            <a class="block p-3 my-1 leading-normal rounded shadow bg-white" href="/file/{el.fileId}">
                <i class="fa fa-file-o" aria-hidden="true"></i> {el.path}
            </a>
        {/if}
    {/each}
{:catch err}
    Error: {err}
{/await}

<script>
// Props for the view
// Path is the path to list
export let path = ''

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
}

function requestTree(reqPath) {
    // Request the tree
    return fetch('/api/tree/' + reqPath)
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
</script>