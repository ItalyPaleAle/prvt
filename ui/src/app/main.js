// Style
import '../css/style.css'

// Themes
import './lib/theme'

// Stores
import {wasm} from './stores'

// Svelte app
import App from './App.svelte'
import LoadingApp from './LoadingApp.svelte'

;(async function main() {
    // Show the LoadingApp component while the app is initializing
    const loading = new LoadingApp({
        target: document.body,
    })

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

    // Check if we need to enable wasm
    if (localStorage.getItem('useWasm') == '1') {
        await enableWasm(true)
    }

    // Remove the loading component
    loading.$destroy()

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
                const active = event.data.enabled
                
                wasm.set(active)
                navigator.serviceWorker.removeEventListener('message', cb)

                // eslint-disable-next-line no-console
                console.log(active ? 'Wasm enabled' : 'Wasm disabled')

                resolve(event.data)
            }
        }
        navigator.serviceWorker.addEventListener('message', cb)
    })
}

/* global PRODUCTION */
if (!PRODUCTION) {
    window.enableWasm = enableWasm
}
