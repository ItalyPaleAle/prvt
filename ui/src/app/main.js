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

    // Initialize the Svelte app and inject it in the DOM
    new App({
        target: document.body,
    })
})()

window.enableWasm = (enable) => {
    navigator.serviceWorker.controller.postMessage({
        message: 'wasm',
        enable
    })
}

navigator.serviceWorker.onmessage = (event) => {
    if (event && event.data) {
        // eslint-disable-next-line no-console
        console.log('Message from SW:', event.data)
    }
}
