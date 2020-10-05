// Style
import '../css/style.css'

// WebAssembly app
import WasmApp from './WasmApp.svelte'

;(async function main() {
    // Register the service worker
    try {
        await navigator.serviceWorker.register('sw.js')
        console.log('Registration succeeded')
    }
    catch (err) {
        console.log('Registration failed with ' + err)
    }

    await navigator.serviceWorker.ready

    // Initialize the Svelte app and inject it in the DOM
    const app = new WasmApp({
        target: document.body,
    })
})()
