<ul class="p-4">
  {#each themes as t}
    <li class="cursor-pointer flex px-4 py-2 text-text-base hover:bg-shade-200" on:click={() => setTheme(t)}>
      <span class="color-circle theme-{t}" title="Theme: {t}" aria-hidden="true"></span>
      <span class="ml-2">Theme: {t}</span>
    </li>
  {/each}
</ul>

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
// Theme data
import {themes} from '../theme'
import theme from '../theme'

// Stores
import {modal} from '../stores'

// Set the theme
function setTheme(t) {
  $theme = t
  // Hide the modal
  $modal = null
}
</script>
