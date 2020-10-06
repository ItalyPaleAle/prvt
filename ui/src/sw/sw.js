/* global Go, Prvt, URL_PREFIX */

// Import the Go WebAssembly runtime
import './wasm_exec'

// Imports from workbox, useful for setting up the service worker
import {clientsClaim} from 'workbox-core'
import {precacheAndRoute} from 'workbox-precaching'

// Handler for fetch requests
import requestHandler from './requests'

// Enable skipWaiting and clientsClaim options
clientsClaim()

// Automatically pre-cache all assets from Webpack - this will contain auto-generated code
precacheAndRoute(self.__WB_MANIFEST)

// Utilities
import {DecodeArrayBuffer, RemovePaddingChars} from './lib/Base64Utils'

// TODO: DEV ONLY
const masterKey = DecodeArrayBuffer(RemovePaddingChars('05qzELK0Nyg9yKaByRHzzsptSXY6SmP/4+VjZLaUYoM='))
self.masterKey = masterKey

// Handle the events to turn on and off in-browser E2EE via Wasm
let wasmEnabled = false
let go = null
self.addEventListener('message', async (event) => {
    // Only listen to wasm event
    if (!event || !event.data || event.data.message != 'wasm') {
        return
    }

    // Check if we are enabling or disabling Wasm
    if (event.data.enable && !wasmEnabled) {
        // Initialize the Go object and load the Wasm file if this is the first time we're enabling Wasm
        if (!go) {
            go = new Go()

            // Fetch the Wasm code
            const result = await WebAssembly.instantiateStreaming(fetch((URL_PREFIX || '') + '/ui/app.wasm'), go.importObject)
            go.run(result.instance)
            
            // Set the base URL
            if (URL_PREFIX) {
                Prvt.setBaseURL(URL_PREFIX)
            }
        }

        // Add the event listener that captures fetch requests
        self.addEventListener('fetch', requestHandler)

        // Set the flag
        wasmEnabled = true
    }
    else if (!event.data.enable && wasmEnabled) {
        // Remove the event listener, which effectively turns off the Wasm functionality
        self.removeEventListener('fetch', requestHandler)

        // Set the flag
        wasmEnabled = false
    }

    // Check if we have clients to notify
    const list = await self.clients.matchAll()
    list.forEach(c => {
        c.postMessage({
            message: 'wasm',
            enabled: wasmEnabled
        })
    })
})
/*
;(async function main() {
    // Load the WebAssembly
    const result = await WebAssembly.instantiateStreaming(fetch('http://localhost:3129/ui/app.wasm'), go.importObject)
    go.run(result.instance)

    // Get the index
    const index = Prvt.getIndex('http://localhost:3129', new Uint8Array(DecodeArrayBuffer(RemovePaddingChars(masterKey))))
    console.log(index)
    console.log(await index.listFolder('/'))
    console.log(await index.stat())
    console.log(await index.getFileByPath('/bensound-energy.mp3'))
    console.log(await index.getFileById('f7a82545-fe70-4681-a5b2-ab593ebfab37'))
})()*/

/*
Test:

await (await fetch('http://localhost:3129/rawfile/015d4c16-2d95-4059-9f99-d91055c7a955', {
    headers: {
        Range: 'bytes=600000-'
    }
})).text()
*/
