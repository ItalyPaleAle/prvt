<p class="text-text-300">
    <span class="cursor-pointer" on:click={themesMenuClick} >
        <i class="fa fa-paint-brush text-sm" aria-hidden="true"></i>
        Theme
    </span>
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
