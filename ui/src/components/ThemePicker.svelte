{#each themes as t}
    <span class="color-picker cursor-pointer theme-{t}" title="Theme: {t}" on:click={() => $theme = t}>
        <span class="color-circle" aria-hidden="true"></span>
        <span class="sr-only">Theme: {t}</span>
    </span>
{/each}

<style>
.color-picker {
    display: inline-block;
    padding: 0.5em;
}
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
        linear-gradient(360deg, var(--picker-color-left) 100%, lightgrey 100%),
        linear-gradient(360deg, var(--picker-color-right) 100%, lightgrey 100%);
    background-position: left top, right top;
}
</style>

<script>
import {themes} from '../theme'
import theme from '../theme'
</script>
