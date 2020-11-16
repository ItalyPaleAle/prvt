<footer class="block p-3 mt-8 mb-2 text-xs text-text-200 space-y-2">
  <p>Powered by <a href="https://github.com/ItalyPaleAle/prvt" class="text-accent-300 hover:underline">prvt</a> {appVersion}</p>
  {#if $showStoreName && $AppInfo}
    {#if $AppInfo.storeType && $AppInfo.storeAccount}
      <p>
        Using store {$AppInfo.storeType} :: {$AppInfo.storeAccount} – <a href="#/repo" class="text-accent-300 hover:underline">change</a>
        {#if AppInfo.repoId}
          <br/>Repository ID: {$AppInfo.repoId}
        {/if}
      </p>
    {/if}
    {#if !$AppInfo.repoUnlocked}
      <p class="text-accent-300">Repository has not been unlocked</p>
    {/if}
    {#if $AppInfo.readOnly}
      <p class="text-accent-300">Read-only mode enabled</p>
    {/if}
  {/if}
  <p class="ml-2 pt-3">
    <SettingsLink />
  </p>
</footer>

<script>
/* global APP_VERSION */
import {showStoreName} from '../stores'
import AppInfo from '../lib/appinfo'
import SettingsLink from './SettingsLink.svelte'
const appVersion = APP_VERSION
</script>
