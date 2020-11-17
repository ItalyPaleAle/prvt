// Style
import '../css/style.css'

// Themes
import './lib/theme'

// Stores and app info
import {wasm} from './stores'
import AppInfo from './lib/appinfo'

// Libraries
import {tick} from 'svelte'

// Svelte app
import App from './App.svelte'
import LoadingApp from './LoadingApp.svelte'

;(async function main() {
    let app

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

    // Listen to messages coming from the service worker
    let wasmCb = null
    navigator.serviceWorker.addEventListener('message', async (event) => {
        if (!event || !event.data) {
            return
        }

        console.log('received message', event.data)
        if (event.data.message == 'wasm') {
            const active = event.data.enabled
            
            // Set the value in the wasm store
            wasm.set(active)

            // eslint-disable-next-line no-console
            console.log(active ? 'Wasm enabled' : 'Wasm disabled')

            // If there's an app mounted, that means this is not the startup sequence, so…
            if (app) {
                // 1. …Reset app info cache
                AppInfo.reset()

                // 2. …Force a reload of the route
                // TODO: THIS ISN'T WORKING
                app.$set('hide', true)
                await tick()
                app.$set('hide', false)
            }

            // Invoke the callback if any
            if (wasmCb) {
                wasmCb(active)
                wasmCb = null
            }
        }
    })

    // Request wasm status
    await new Promise((resolve) => {
        wasmCb = resolve
        // Send the request
        navigator.serviceWorker.controller.postMessage({
            message: 'get-wasm'
        })
    })

    // Remove the loading component
    loading.$destroy()

    // Initialize the Svelte app and inject it in the DOM
    app = new App({
        target: document.body,
    })
})()

const enableWasm = (enabled) => {
    navigator.serviceWorker.controller.postMessage({
        message: 'set-wasm',
        enabled
    })
}

/* global PRODUCTION */
if (!PRODUCTION) {
    window.enableWasm = enableWasm
}
