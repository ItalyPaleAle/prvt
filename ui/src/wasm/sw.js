// Import from workbox, useful for setting up the service worker
//import {skipWaiting, clientsClaim} from 'workbox-core'
import {precacheAndRoute} from 'workbox-precaching'

// Enable skipWaiting and clientsClaim options
//skipWaiting()
//clientsClaim()

// Automatically pre-cache all assets from Webpack - this will contain auto-generated code
precacheAndRoute(self.__WB_MANIFEST)

// Import the Go WebAssembly runtime
import './wasm_exec'
/* global Go, Prvt */

// Utilities
import {DecodeArrayBuffer, RemovePaddingChars} from './lib/Base64Utils'

// Initialize the Go object
const go = new Go()

self.addEventListener('fetch', (event) => {
    console.log(event.request.method, event.request.url, event.request.headers.get('range'))
    const dest = new URL(event.request.url)
    if (dest.pathname.startsWith('/rawfile')) {
        event.respondWith(decryptRequest(event.request))
    }
})

const masterKey = '05qzELK0Nyg9yKaByRHzzsptSXY6SmP/4+VjZLaUYoM='

/**
 * @param {Request} req - Request object
 */
async function decryptRequest(req) {
    const response = await Prvt.decryptRequest(
        new Uint8Array(DecodeArrayBuffer(RemovePaddingChars(masterKey))),
        req
    )
    if (!response) {
        throw Error('Response from decryptPackages is empty')
    }
    else if (typeof response == 'object' && response instanceof Error) {
        throw response
    }

    return response
}

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
})()

/*
Test:

await (await fetch('http://localhost:3129/rawfile/015d4c16-2d95-4059-9f99-d91055c7a955', {
    headers: {
        Range: 'bytes=600000-'
    }
})).text()
*/
