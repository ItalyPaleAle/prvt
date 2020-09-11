<footer class="block p-3 mt-8 mb-2 leading-normal text-xs">
  <p>
    Powered by <a href="https://github.com/ItalyPaleAle/prvt">prvt</a> {appVersion}
    {#if $readOnly} – <b class="text-accent-300">Read-only mode</b>{/if}
  </p>
  <ThemePicker />
</footer>

<script>
/* global env */
import {readOnly} from '../stores'
import ThemePicker from './ThemePicker.svelte'
const appVersion = env.APP_VERSION
</script>
