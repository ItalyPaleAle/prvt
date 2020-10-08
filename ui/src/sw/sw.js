/* global Go, Prvt, URL_PREFIX */

// Import the Go WebAssembly runtime
import './wasm_exec'

// Imports from workbox, useful for setting up the service worker
import {clientsClaim} from 'workbox-core'
import {precacheAndRoute} from 'workbox-precaching'

// Handler for fetch requests
import requestHandler from './requests'

// Stores
import stores from './stores'

// Enable skipWaiting and clientsClaim options
clientsClaim()

// Automatically pre-cache all assets from Webpack - this will contain auto-generated code
precacheAndRoute(self.__WB_MANIFEST)

// Handle the events to turn on and off in-browser E2EE via Wasm
let wasmEnabled = false
let go = null
self.addEventListener('message', async (event) => {
    if (!event || !event.data) {
        return
    }

    let list
    switch (event.data.message) {
        // Message 'wasm' is for enabling or disabling Wasm
        case 'wasm':
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

                // Unset the master key and related objects
                stores.masterKey = null
                stores.keyId = null
                stores.index = null
            }

            // Check if we have clients to notify
            list = await self.clients.matchAll()
            list.forEach(c => {
                c.postMessage({
                    message: 'wasm',
                    enabled: wasmEnabled
                })
            })
            break

        // Message 'masterKey' is for setting the master key
        case 'masterKey':
            stores.masterKey = event.data.masterKey
            stores.keyId = event.data.keyId
            stores.index = Prvt.getIndex(stores.masterKey)
            break

        // Do nothing otherwise
        default:
            break
    }
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
