// Style
import '../css/style.css'

// Themes
import theme from './lib/theme'

// Libraries
import {SvelteComponent, tick} from 'svelte'
import {push, location} from 'svelte-spa-router'
import {get} from 'svelte/store'
import controlled from './lib/sw-controlled'

// Stores and app info
import {wasm} from './stores'
import AppInfo from './lib/appinfo'

// Svelte app
import App from './App.svelte'
import LoadingApp from './LoadingApp.svelte'
import ErrorApp from './ErrorApp.svelte'

// App currently mounted
let app: SvelteComponent | null = null

// Flag that informs if the startup sequence was completed
let startupComplete = false

;(async function main() {
    // Show the LoadingApp component while the app is initializing
    mountApp('loading')

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
        const errMsg = 'Service worker registration failed with ' + err
        // eslint-disable-next-line no-console
        console.error(errMsg)

        // Show the error app and return
        mountApp('error', {
            error: errMsg
        })
        return
    }

    // Listen to messages coming from the service worker
    navigator.serviceWorker.addEventListener('message', swMessage)

    const controller = navigator.serviceWorker.controller
    if (!controller) {
        // Should never happen
        throw Error('navigator.serviceWorker.controller is empty')
    }

    // Request the current theme from the service worker
    // Initially, the theme is loaded from localStorage, but that might be out of sync
    controller.postMessage({
        message: 'get-theme'
    } as ServiceWorkerMessage)

    // Request wasm status
    // The receiver for the get-wasm message also initializes the app
    controller.postMessage({
        message: 'get-wasm'
    } as ServiceWorkerMessage)
})()

// Mount the desired app
function mountApp(type: 'loading' | 'app' | 'error', props?: Record<string,any>) {
    // If there's currently an app mounted, un-mount it first
    if (app) {
        app.$destroy()
    }

    // Mount the desired app
    const opts = {
        target: document.body,
        props
    }
    switch (type) {
        case 'app':
            app = new App(opts)
            break
        case 'loading':
            app = new LoadingApp(opts)
            break
        case 'error':
            app = new ErrorApp(opts)
            break
    }
}

// Handle messages from the service worker
async function swMessage(event: MessageEvent<ServiceWorkerMessage>) {
    if (!event?.data) {
        return
    }

    switch (event.data.message) {
        // We need to un-mount the app and display the loading component
        // This happens for example when wasm state is about to change
        case 'off':
            // Display the loading component
            mountApp('loading')
            break

        // The repo was unlocked
        case 'unlocked':
            // Refresh app info cache
            await AppInfo.update()

            // If we are on the /unlock route, go to the main view
            if (startupComplete && get(location) == '/unlock') {
                push('/')
            }
            break

        // Wasm was enabled or disabled
        case 'wasm':
            // Set the value in the wasm store
            wasm.set(event.data.enabled)

            // eslint-disable-next-line no-console
            console.info(event.data.enabled ? 'Wasm enabled' : 'Wasm disabled')

            // Wait for the next tick
            await tick()

            // If the startup sequence was already done, we need to do some other setup
            if (startupComplete) {
                // 1. …Refresh app info cache
                const info = await AppInfo.update()

                // 2. …If the repo is now locked, redirect users to unlock
                // Otherwise, if we're in the /unlock route already, go to the main view
                if (!info?.repoUnlocked) {
                    push('/unlock')
                }
                else {
                    if (get(location) == '/unlock') {
                        push('/')
                    }
                }
            }
            startupComplete = true

            // Ensure the app component is displayed
            mountApp('app')

            break

        // Theme has changed
        case 'theme':
            theme.set(event.data.theme)
            break
    }
}
