// Style
import '../css/style.css'

// Themes
import './lib/theme'

// Libraries
import {push, location} from 'svelte-spa-router'
import {get} from 'svelte/store'

// Stores and app info
import {wasm} from './stores'
import AppInfo from './lib/appinfo'

// Svelte app
import App from './App.svelte'
import LoadingApp from './LoadingApp.svelte'

;(async function main() {
    let app = null

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

        switch (event.data.message) {
            case 'wasm':                
                // Set the value in the wasm store
                wasm.set(event.data.enabled)

                // eslint-disable-next-line no-console
                console.log(event.data.enabled ? 'Wasm enabled' : 'Wasm disabled')

                // If there's an app mounted, that means this is not the startup sequence, so…
                if (app) {
                    // 1. …Refresh app info cache
                    const info = await AppInfo.update()

                    // 2. …If the repo is now locked, redirect users to unlock
                    // Otherwise, if we're in the /unlock route already, go to the main view
                    if (!info || !info.repoUnlocked) {
                        push('/unlock')
                    }
                    else {
                        if (get(location) == '/unlock') {
                            push('/')
                        }
                    }
                }

                // Invoke the callback if any
                if (wasmCb) {
                    wasmCb(event.data.enabled)
                    wasmCb = null
                }
                break
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
