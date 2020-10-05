<footer class="block p-3 mt-8 mb-2 text-xs text-text-200 space-y-2">
    <p>Powered by <a href="https://github.com/ItalyPaleAle/prvt" class="text-accent-300">prvt</a> {appVersion}</p>
    {#if $appInfo && $appInfo.repoId}
      <p>
        Using store {$appInfo.storeType} :: {$appInfo.storeAccount} <br/>
        Repository ID: {$appInfo.repoId}
      </p>
    {/if}
    {#if $appInfo && !$appInfo.repoUnlocked}
      <p class="text-accent-300">Repository has not been unlocked</p>
    {/if}
    {#if $readOnly}
      <p class="text-accent-300">Read-only mode enabled</p>
    {/if}
    <p class="ml-2 pt-3">
      <ThemePicker />
    </p>
  </footer>
  
  <script>
  /* global env */
  import {readOnly, appInfo} from '../stores'
  import ThemePicker from './ThemePicker.svelte'
  const appVersion = APP_VERSION
  </script>
  