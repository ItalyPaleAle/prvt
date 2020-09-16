// Style
import '../css/style.css'

// Themes
import '../shared/lib/theme'

// Initialize the Svelte app and inject it in the DOM
import RepoApp from './RepoApp.svelte'
const app = new RepoApp({
    target: document.body,
})

export default app
