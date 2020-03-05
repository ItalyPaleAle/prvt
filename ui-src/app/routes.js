// Components
import Home from './routes/Home.svelte'
import NotFound from './routes/NotFound.svelte'

const routes = {
    // Home
    '/': Home,
        
    // Catch-all, must be last
    '*': NotFound,
}

export default routes
