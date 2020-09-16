import {wrap} from 'svelte-spa-router'
import {readOnly} from '../shared/stores'

// Components
import Tree from './routes/Tree.svelte'
import Add from './routes/Add.svelte'
import View from './routes/View.svelte'
import NotFound from './routes/NotFound.svelte'

// Route definition object
export default {
    // Tree
    '/': Tree,
    '/tree': Tree,
    '/tree/*': Tree,

    // Add
    '/add/*': wrap(Add, noReadOnly),

    // View
    '/view/:fileId': View,
        
    // Catch-all, must be last
    '*': NotFound,
}

// Flag for when we're in read-only mode
let readOnlyFlag = false
readOnly.subscribe((val) => readOnlyFlag = val)

// Allow a route only in non read-only mode
function noReadOnly(detail) {
    return !readOnlyFlag
}
