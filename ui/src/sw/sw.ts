// Globals
declare var self: ServiceWorkerGlobalScope
declare var APP_VERSION: string
declare var URL_PREFIX: string
declare var PRODUCTION: boolean

// Import the Go WebAssembly runtime
import './wasm_exec'

// Imports from workbox, useful for setting up the service worker
import {PrecacheController, PrecacheRoute} from 'workbox-precaching'
import {registerRoute} from 'workbox-routing'

// Handler for fetch requests
import requestHandler from './requests'

// Stores and settings
import stores from './stores'
import * as settings from './settings'

// Utils
import {BroadcastMessage} from './lib/utils'

// URL of the Wasm file
const wasmURL = 'assets/app-' + APP_VERSION + '.wasm'

// Flag used to know when the Go runtime has been loaded
let goLoaded = false

// Automatically pre-cache all assets from Webpack - this will contain auto-generated code
const precacheController = new PrecacheController()
const wbManifest = self.__WB_MANIFEST
if (wbManifest) {
    precacheController.addToCacheList(wbManifest)
}

// Listen to the service worker installation event
self.addEventListener('install', (event) => {
    // Enable the skipWaiting option, meaning that the service worker will become active immediately
    event.waitUntil(self.skipWaiting())

    // Install the precache controller
    // This calls event.waitUntil internally
    precacheController.install(event)
})

// On activation, load all settings and check if we want to enable wasm mode
// These are stored in IndexedDB for persisting
self.addEventListener('activate', (event) => {
    event.waitUntil((async () => {
        // Activate the precache controller
        // In parallel, check whether wasm is enabled
        await Promise.all([
            // precacheController.activate calls event.waitUntil internally
            // However, we need to have control and get things done on our own terms
            // So, we're passing a stub to the activate() method rather than the event object, then we call even.waitUntil here
            // See: https://github.com/GoogleChrome/workbox/issues/2694 
            precacheController.activate({waitUntil: () => {}} as any),
            // Check whether wasm is enabled
            settings.Get('wasm')
                .then((wasm) => enableWasm(!!wasm))
        ])

        // Invoke clients.claim, which makes all tabs use this service worker
        await self.clients.claim()
    })())
})

// Add the event listener that can capture fetch requests
self.addEventListener('fetch', requestHandler)

// Add another fetch event listener for precached resource, after the requestHandler above
registerRoute(
    new PrecacheRoute(
        precacheController,
        {
            // Ignore all URL parameters
            ignoreURLParametersMatching: [/.*/],
            // Do not add .html to files by default
            cleanURLs: false,
        }
    )
)

// Handle the events to turn on and off in-browser E2EE via Wasm
self.addEventListener('message', async (event) => {
    if (!(event?.data) || !(event?.source)) {
        return
    }

    switch (event.data.message) {
        // Message 'get-wasm' requests the status of Wasm
        case 'get-wasm':
            // Respond
            event.source.postMessage({
                message: 'wasm',
                enabled: stores.wasm
            }, [])
            break

        // Message 'set-wasm' is for enabling or disabling Wasm
        case 'set-wasm':
            // Enable or disable wasm
            await enableWasm(!!(event.data && event.data.enabled))

            // Notify all clients
            // No need to await on this, just let it run in background
            BroadcastMessage({
                message: 'wasm',
                enabled: stores.wasm
            })
            break

        // Message 'get-theme' requests the current theme
        case 'get-theme':
            // Respond
            event.source.postMessage({
                message: 'theme',
                theme: await settings.Get('theme')
            }, [])
            break

        // Message 'set-theme' sets a new theme
        case 'set-theme':
            // Set the preference
            await settings.Set('theme', (event.data && event.data.theme) || '')

            // Notify all clients
            // No need to await on this, just let it run in background
            BroadcastMessage({
                message: 'theme',
                theme: (event.data && event.data.theme) || ''
            })
            break

        // Message 'set-master-key' is for overriding the master key
        // This is normally used in development only
        case 'set-master-key':
            if (typeof event.data.masterKey != 'object' || !(event.data.masterKey instanceof Uint8Array)) {
                throw Error('Invalid type for masterKey: must be Uint8Array')
            }
            if (typeof event.data.keyId != 'string' || !event.data.keyId) {
                throw Error('KeyId must not be empty')
            }
            stores.masterKey = event.data.masterKey as Uint8Array
            stores.keyId = event.data.keyId as string
            stores.index = Prvt.getIndex(stores.masterKey)
            break

        // Do nothing otherwise
        default:
            break
    }
})

// Enable or disable Wasm
async function enableWasm(enable: boolean): Promise<void> {
    // Check if we are enabling or disabling Wasm
    if (enable && !stores.wasm) {
        // Initialize the Go object and load the Wasm file if this is the first time we're enabling Wasm
        if (!goLoaded) {
            const go = new Go()

            // Fetch the Wasm code
            const result = await WebAssembly.instantiateStreaming(fetchWasm(), go.importObject)
            go.run(result.instance)

            // Set the base URL
            if (URL_PREFIX) {
                Prvt.setBaseURL(URL_PREFIX)
            }
        }

        // Enable wasm functionality, by telling the service worker to start intercepting requests
        stores.wasm = true

        // Update the value in IndexedDB
        await settings.Set('wasm', true)
    }
    else if (!enable && stores.wasm) {
        // Turn off wasm functionality, by stopping intercepting requests
        stores.wasm = false

        // Unset the master key and related objects
        stores.masterKey = undefined
        stores.keyId = undefined
        stores.index = undefined

        // Update the value in IndexedDB
        await settings.Set('wasm', false)
    }
    else {
        // Just ensure that the values in the store and settings are in sync
        stores.wasm = enable
        await settings.Set('wasm', enable)
    }
}

// Fetches the Wasm file, trying the cache first
async function fetchWasm(): Promise<Response> {
    const req = new Request(wasmURL)
    
    // If we're not in production, skip the cache
    if (!PRODUCTION) {
        return fetch(req)
    }

    const wasmCache = await caches.open('wasm-cache')

    // Check if we have the URL in the cache, and remove the old ones
    const keys = await wasmCache.keys()
    for (let i = 0; i < keys.length; i++) {
        if (keys[i] && keys[i].url != req.url) {
            // Delete old keys
            await wasmCache.delete(keys[i])
            console.info('Deleted old wasm file', keys[i].url)
        }
    }

    // First try responding from the cache
    // Otherwise, request the file and store it in the cache
    let res = await wasmCache.match(req)
    if (res && res.status >= 200 && res.status < 300) {
        console.info('Loaded wasm from cache', wasmURL)
        return res
    }
    // If we're here, we did not find a match in the cache, so request it and store it in the cache
    res = await fetch(req)
    if (res && res.status >= 200 && res.status < 300) {
        wasmCache.put(req, res.clone())
    }
    return res
}
