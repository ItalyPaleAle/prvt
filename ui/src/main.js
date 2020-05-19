// Style
import './css/style.css'

// Themes
import theme from './theme'
theme.subscribe((theme) => {
    // Remove existing theme tags
    document.body.classList.forEach((val) => {
        if (val && val.indexOf('theme-') === 0) {
            document.body.classList.remove(val)
        }
    })

    // Add the new theme
    document.body.classList.add(theme)
})

// Initialize the Svelte app and inject it in the DOM
import App from './App.svelte'
const app = new App({
    target: document.body
})

export default app
