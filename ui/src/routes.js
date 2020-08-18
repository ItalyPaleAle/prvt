// Components
import Tree from './routes/Tree.svelte'
import Add from './routes/Add.svelte'
import View from './routes/View.svelte'
import NotFound from './routes/NotFound.svelte'

const routes = {
    // Tree
    '/': Tree,
    '/tree': Tree,
    '/tree/*': Tree,

    // Add
    '/add/*': Add,

    // View
    '/view/:fileId': View,
        
    // Catch-all, must be last
    '*': NotFound,
}

export default routes
