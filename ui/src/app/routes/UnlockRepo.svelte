<PageTitle title="Unlock repository" backButton={"#/repo"} />

{#await requesting}
  <p>Unlocking…</p>
{:then _}
  <div class="md:flex md:flex-row">
    <div class="md:w-1/2 flex-grow p-4 mx-3 mb-4 bg-shade-neutral flex flex-col justify-between">
      <h2 class="text-2xl mb-3">Passphrase</h2>
      <input type="password" name="passphrase" placeholder="Passphrase" class="bg-shade-neutral appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 mb-3 text-text-300 leading-tight focus:outline-none focus:bg-shade-neutral focus:border-accent-200" bind:value={passphrase} />
      <button type="button" class="shadow bg-accent-200 hover:bg-accent-100 focus:shadow-outline focus:outline-none text-text-200 font-bold py-2 px-4 rounded" on:click={unlockPassphrase}>Unlock with passphrase</button>
    </div>
    {#if gpgUnlock}
      <div class="md:w-1/2 p-4 mx-3 mb-4 bg-shade-neutral flex flex-col justify-between">
        <h2 class="text-2xl mb-3">GPG Key</h2>
        <button type="button" class="shadow bg-accent-200 hover:bg-accent-100 focus:shadow-outline focus:outline-none text-text-200 font-bold py-2 px-4 rounded" on:click={unlockGPG}>Unlock with a GPG key</button>
      </div>
    {/if}
  </div>
{:catch err}
  <p>Error: {err}</p>
{/await}

<script>
import {Request} from '../lib/request'
import AppInfo from '../lib/appinfo'
import {querystring, push} from 'svelte-spa-router'

// Components
import PageTitle from '../components/PageTitle.svelte'

// Enable unlock with a GPG key if the repo supports it
$: gpgUnlock = $querystring == 'gpg=1'

let requesting = null
let passphrase = ''

// Unlock with passphrase
function unlockPassphrase() {
    if (!passphrase) {
        return
    }

    requesting = doUnlock({type: 'passphrase', passphrase})
}

// Unlock with a GPG key
function unlockGPG() {
    requesting = doUnlock({type: 'gpg'})
}

async function doUnlock(postData) {
    // Make the unlock request
    await Request('/api/repo/unlock', {postData})

    // On success, refresh AppInfo and go back to the app
    await AppInfo.update()
    push('/')
}
</script>
