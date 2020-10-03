// Style
import '../css/style.css'

// Themes
import '../shared/lib/theme'

// Stores
import {currentApp} from '../shared/stores'
currentApp.set('app')

// Initialize the Svelte app and inject it in the DOM
import App from './App.svelte'
const app = new App({
    target: document.body,
})

export default app
