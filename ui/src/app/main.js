// Style
import '../css/style.css'

// Themes
import theme from './lib/theme'

// Libraries
import {push, location} from 'svelte-spa-router'
import {get} from 'svelte/store'
import controlled from './lib/sw-controlled'

// Stores and app info
import {wasm} from './stores'
import AppInfo from './lib/appinfo'

// Svelte app
import App from './App.svelte'
import LoadingApp from './LoadingApp.svelte'

let app = null
let wasmCb = null

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

        // See: https://github.com/w3c/ServiceWorker/issues/799
        //await navigator.serviceWorker.ready
        await controlled

        // eslint-disable-next-line no-console
        console.info('Service worker activated')
    }
    catch (err) {
        // eslint-disable-next-line no-console
        console.error('Service worker registration failed with ' + err)

        // TODO: SHOW ERROR IN PAGE AS THE SITE IS BROKEN NOW
    }

    // Listen to messages coming from the service worker
    navigator.serviceWorker.addEventListener('message', swMessage)

    // Request the current theme from the service worker
    // Initially, the theme is loaded from localStorage, but that might be out of sync
    navigator.serviceWorker.controller.postMessage({
        message: 'get-theme'
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

async function swMessage(event) {
    if (!event || !event.data) {
        return
    }

    switch (event.data.message) {
        // The repo was unlocked
        case 'unlocked':
            // Refresh app info cache
            await AppInfo.update()

            // If we are on the /unlock route, go to the main view
            if (app && get(location) == '/unlock') {
                push('/')
            }
            break

        // Wasm was enabled or disabled
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

        // Theme has changed
        case 'theme':
            theme.set(event.data.theme)
            break
    }
}
