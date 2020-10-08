// Style
import '../css/style.css'

// Themes
import './lib/theme'

// Svelte app
import App from './App.svelte'

;(async function main() {
    // Register the service worker and wait for its activation
    try {
        await navigator.serviceWorker.register('sw.js')
        // eslint-disable-next-line no-console
        console.info('Service worker registered')
        await navigator.serviceWorker.ready
        // eslint-disable-next-line no-console
        console.info('Service worker activated')
    }
    catch (err) {
        // eslint-disable-next-line no-console
        console.error('Service worker registration failed with ' + err)
    }

    await enableWasm(true)

    // Initialize the Svelte app and inject it in the DOM
    new App({
        target: document.body,
    })
})()

const enableWasm = (enable) => {
    navigator.serviceWorker.controller.postMessage({
        message: 'wasm',
        enable
    })
    return new Promise((resolve) => {
        const cb = (event) => {
            if (event && event.data && event.data.message == 'wasm') {
                // eslint-disable-next-line no-console
                console.log(event.data.enabled ? 'Wasm enabled' : 'Wasm disable')
            }
            navigator.serviceWorker.removeEventListener('message', cb)
            resolve(event.data)
        }
        navigator.serviceWorker.addEventListener('message', cb)
    })
}
window.enableWasm = enableWasm
