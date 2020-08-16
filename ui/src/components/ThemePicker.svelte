<p on:click={themesMenuClick}>
    <span class="underline cursor-pointer">Theme</span>
</p>

<script>
// Components
import ThemePickerModal from './ThemePickerModal.svelte'

// Stores
import {modal} from '../stores'

// Open the modal on click
function themesMenuClick() {
    $modal = {
        component: ThemePickerModal
    }
}
</script>
