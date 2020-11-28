<div class="space-y-8">
    {#if loading}
        <div class="space-y-4">
            <Spinner title="Working" />
        </div>
    {:else}
        <div class="space-y-4">
            <h1 class="flex text-2xl break-all text-accent-300">Theme</h1>
            <div class="mx-2">
                <SettingsThemePicker />
            </div>
        </div>
        <div class="space-y-4">
            <h1 class="flex text-2xl break-all text-accent-300">Advanced</h1>
            <div class="mx-2">
                <button on:click={toggleWasm}>
                    {#if $wasm}
                        <i class="fa fa-check-square-o fa-lg mr-2" aria-hidden="true"></i>
                        <span class="sr-only">Currently enabled</span>
                    {:else}
                        <i class="fa fa-square-o fa-lg mr-2" aria-hidden="true"></i>
                        <span class="sr-only">Currently disabled</span>
                    {/if}
                    In-Browser End-to-End Encryption
                </button>
            </div>
        </div>
    {/if}
</div>

<script>
// Components
import SettingsThemePicker from './SettingsThemePicker.svelte'
import Spinner from './Spinner.svelte'

// Stores
import {wasm, modal} from '../stores'

// Utils
import {enableWasm} from '../lib/utils'

let loading = false

// Enable or disable wasm
async function toggleWasm() {
    // Request change
    enableWasm(!$wasm)
    
    // Wait for the operation to be completed and show the loading view in the meanwhile
    loading = true
    let unsubscribe
    await new Promise((resolve) => {
        unsubscribe = wasm.subscribe(resolve)
    })

    // All done
    loading = false
    $modal = null

    // Unsubscribe from the store
    unsubscribe()
}
</script>
