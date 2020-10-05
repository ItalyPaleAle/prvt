import {wrap} from 'svelte-spa-router/wrap'
import {readOnly} from './stores'

// Components
import Tree from './routes/Tree.svelte'
import Add from './routes/Add.svelte'
import View from './routes/View.svelte'
import NotFound from './routes/NotFound.svelte'
import ConnectionList from './routes/ConnectionList.svelte'
import UnlockRepo from './routes/UnlockRepo.svelte'

// Route definition object
export default {
    // Tree
    '/': Tree,
    '/tree': Tree,
    '/tree/*': Tree,

    // Add
    '/add/*': wrap({
        component: Add,
        conditions: [
            noReadOnly
        ]
    }),

    // View
    '/view/:fileId': View,

    // Repo select
    '/repo': ConnectionList,

    // Unlock repo
    '/unlock': UnlockRepo,
        
    // Catch-all, must be last
    '*': NotFound,
}

// Flags for when we're in read-only mode, for when the repo is selected, and for when the repo is unlocked
let readOnlyFlag = false
readOnly.subscribe((val) => readOnlyFlag = val)

// Allow a route only in non read-only mode
function noReadOnly(detail) {
    return !readOnlyFlag
}
