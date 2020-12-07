<p class="text-text-300">
    <span class="cursor-pointer" on:click={themesMenuClick} >
        <i class="fa fa-cog text-sm" aria-hidden="true"></i>
        Settings and themes
    </span>
</p>

<script lang="ts">
// Components
import SettingsModal from './SettingsModal.svelte'

// Stores
import {modal} from '../stores'

// Open the modal on click
function themesMenuClick() {
    $modal = {
        component: SettingsModal
    }
}
</script>
