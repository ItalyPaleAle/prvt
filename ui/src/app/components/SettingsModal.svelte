<h1 class="flex text-2xl break-all mb-4 text-accent-300">Settings</h1>
<button on:click={toggleWasm}>{$wasm ? 'Disable' : 'Enable'} In-Browser End-to-End Encryption</button>

<h1 class="flex text-2xl break-all mb-4 mt-8 text-accent-300">Theme</h1>
<SettingsThemePicker />

<script>
// Components
import SettingsThemePicker from './SettingsThemePicker.svelte'

// Stores
import {wasm} from '../stores'

// Enable or disable wasm
function toggleWasm() {
    localStorage.setItem('useWasm', $wasm ? '0' : '1')
    location.replace('app.html#')
    location.reload()
}
</script>
