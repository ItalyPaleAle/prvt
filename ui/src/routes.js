// Components
import Tree from './routes/Tree.svelte'
import NotFound from './routes/NotFound.svelte'

const routes = {
    // Home
    '/': Tree,
    '/tree': Tree,
    '/tree/*': Tree,
        
    // Catch-all, must be last
    '*': NotFound,
}

export default routes
