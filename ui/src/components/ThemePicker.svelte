<p on:click={themesMenuClick}>
    <span class="underline cursor-pointer">Theme</span>
    {#if expandDropdown}
        <div class="absolute text-sm">
            <ul class="py-1 my-2 mx-2 w-64 bg-shade-neutral rounded shadow">
                {#each themes as t}
                    <li class="cursor-pointer flex block px-4 py-2 text-text-base hover:bg-shade-200" on:click={() => $theme = t}>
                        <span class="color-circle theme-{t}" title="Theme: {t}" aria-hidden="true"></span>
                        <span class="ml-2">Theme: {t}</span>
                    </li>
                {/each}
            </ul>
        </div>
    {/if}
</p>

<style>
.color-circle {
    display: inline-block;
    width: 1.5em;
    height: 1.5em;

    box-shadow: 
        0 0 0 0.075em #edf2f7,
        0 0 0 0.15em #4a5568;
    border-radius: 50%;
    background-size:
        50% 100%,
        50% 100%;
    background-repeat: no-repeat;
    background-image:
        var(--picker-gradient-left),
        var(--picker-gradient-right);
    background-position: left top, right top;
}
</style>

<script>
// Themes
import {themes} from '../theme'
import theme from '../theme'

// Stores
import {dropdown} from '../stores'

// State
let expandDropdown = false

// There can only be one actions menu open in the entire application, so we use the $dropdown store as semaphore
$: expandDropdown = $dropdown === 'theme'

function themesMenuClick() {
    if (expandDropdown) {
        $dropdown = null
    }
    else {
        $dropdown = 'theme'
    }
}
</script>
