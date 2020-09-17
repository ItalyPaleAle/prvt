// Style
import '../css/style.css'

// Themes
import '../shared/lib/theme'

// Stores
import {currentApp} from '../shared/stores'
currentApp.set('repo')

// Initialize the Svelte app and inject it in the DOM
import RepoApp from './RepoApp.svelte'
const app = new RepoApp({
    target: document.body,
})

export default app
