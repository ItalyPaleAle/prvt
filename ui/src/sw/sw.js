/// <reference path="./Prvt.d.ts" />
/* global Prvt, Go, PRODUCTION, URL_PREFIX, APP_VERSION */

// Import the Go WebAssembly runtime
import './wasm_exec'

// Imports from workbox, useful for setting up the service worker
import {clientsClaim} from 'workbox-core'
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

// Enable skipWaiting and clientsClaim options
clientsClaim()
self.skipWaiting()

// Automatically pre-cache all assets from Webpack - this will contain auto-generated code
const precacheController = new PrecacheController()
if (self.__WB_MANIFEST) {
    console.log('precached', self.__WB_MANIFEST)
    precacheController.addToCacheList(self.__WB_MANIFEST)
}

// Listen to the service worker installation event
self.addEventListener('install', (event) => {
    // Install the precache controller
    // This calls event.waitUntil internally
    precacheController.install(event)
})

// On activation, load all settings and check if we want to enable wasm mode
// These are stored in IndexedDB for persisting
self.addEventListener('activate', (event) => {
    // Activate the precache controller
    // This calls event.waitUntil internally
    precacheController.activate(event)

    // Check if we need to enable wasm
    event.waitUntil(
        settings.Get('wasm')
            .then((wasm) => enableWasm(!!wasm))
    )
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
            cleanUrls: false,
        }
    )
)

// Handle the events to turn on and off in-browser E2EE via Wasm
let go = null
self.addEventListener('message', async (event) => {
    if (!event || !event.data) {
        return
    }

    switch (event.data.message) {
        // Message 'get-wasm' requests the status of Wasm
        case 'get-wasm':
            // Respond
            event.source.postMessage({
                message: 'wasm',
                enabled: stores.wasm
            })
            break

        // Message 'set-wasm' is for enabling or disabling Wasm
        case 'set-wasm':
            // Enable or disable wasm
            enableWasm(!!(event.data && event.data.enabled))

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
            })
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
            stores.masterKey = event.data.masterKey
            stores.keyId = event.data.keyId
            stores.index = Prvt.getIndex(stores.masterKey)
            break

        // Do nothing otherwise
        default:
            break
    }
})

// Enable or disable Wasm
async function enableWasm(enable) {
    // Check if we are enabling or disabling Wasm
    if (enable && !stores.wasm) {
        // Initialize the Go object and load the Wasm file if this is the first time we're enabling Wasm
        if (!go) {
            go = new Go()

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
        stores.masterKey = null
        stores.keyId = null
        stores.index = null

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
async function fetchWasm() {
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
