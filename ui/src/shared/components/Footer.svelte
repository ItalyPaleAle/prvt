<footer class="block p-3 mt-8 mb-2 leading-normal text-xs text-text-200">
    <p>
      Powered by <a href="https://github.com/ItalyPaleAle/prvt" class="text-text-300">prvt</a> {appVersion}
      {#if $readOnly} – <b class="text-accent-300">Read-only mode</b>{/if}
    </p>
    <p class="ml-2 mt-1">
      <ThemePicker />
    </p>
  </footer>
  
  <script>
  /* global env */
  import {readOnly} from '../stores'
  import ThemePicker from './ThemePicker.svelte'
  const appVersion = APP_VERSION
  </script>
  