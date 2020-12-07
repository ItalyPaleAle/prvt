<footer class="block p-3 mt-8 mb-2 text-xs text-text-200 space-y-2">
  <p>Powered by <a href="https://github.com/ItalyPaleAle/prvt" class="text-accent-300 hover:underline">prvt</a> {appVersion}</p>
  {#if $showStoreName && $AppInfo}
    {#if $AppInfo.storeType && $AppInfo.storeAccount}
      <p>
        Using repository <span class="font-bold">{$AppInfo.storeType} :: {$AppInfo.storeAccount}</span>
        {#if !$AppInfo.repoUnlocked}
          (locked)
        {/if}
        – <a href="#/repo" class="text-accent-300 hover:underline">Switch repository</a>
        {#if $AppInfo.repoId}
          <br/>Repository ID: {$AppInfo.repoId}
        {/if}
      </p>
    {/if}
  {/if}
  {#if $wasm}
    <p class="text-accent-300">In-Browser E2EE enabled (read-only)</p>
  {:else if $AppInfo.readOnly}
    <p class="text-accent-300">Read-only mode</p>
  {/if}
  <p class="ml-2 pt-3">
    <SettingsLink />
  </p>
</footer>

<script lang="ts">
import {showStoreName, wasm} from '../stores'
import AppInfo from '../lib/appinfo'
import SettingsLink from './SettingsLink.svelte'
const appVersion = APP_VERSION
</script>
