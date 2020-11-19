<PageTitle title="Unlock repository" />

{#await requesting}
  <Spinner title="Unlocking" />
{:catch err}
  <div class="mx-4">
      <ErrorBox message={(err && err.message) || 'An undefined error eccorrued'} noClose={true} />
  </div>
{/await}

{#if showForm}
  <div class="md:flex md:flex-row">
    <form class="md:w-1/2 flex-grow p-4 mx-3 mb-4 bg-shade-neutral flex flex-col justify-between" on:submit|preventDefault={unlockPassphrase}>
      <h2 class="text-2xl mb-3">Passphrase</h2>
      <input type="password" name="passphrase" placeholder="Passphrase" class="bg-shade-neutral appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 mb-3 text-text-300 leading-tight focus:outline-none focus:bg-shade-neutral focus:border-accent-200" bind:value={passphrase} autocomplete="off" />
      <button type="submit" class="shadow bg-shade-100 hover:bg-shade-200 focus:shadow-outline focus:outline-none text-text-200 font-bold py-2 px-4 rounded">Unlock with passphrase</button>
    </form>
    {#if gpgUnlock && !$wasm}
      <div class="md:w-1/2 p-4 mx-3 mb-4 bg-shade-neutral flex flex-col justify-between">
        <h2 class="text-2xl mb-3">GPG Key</h2>
        <button type="button" class="shadow bg-shade-100 hover:bg-shade-200 focus:shadow-outline focus:outline-none text-text-200 font-bold py-2 px-4 rounded" on:click={unlockGPG}>Unlock with a GPG key</button>
      </div>
    {/if}
  </div>
{/if}

<script>
import {Request} from '../../shared/request'
import {querystring} from 'svelte-spa-router'

// Components
import ErrorBox from '../components/ErrorBox.svelte'
import PageTitle from '../components/PageTitle.svelte'
import Spinner from '../components/Spinner.svelte'

// Stores
import {wasm, fileList, operationResult} from '../stores'

// State
let requesting = null
let passphrase = ''
let showForm = true
let gpgUnlock = undefined

// Enable unlock with a GPG key if the repo supports it
$: getGpgUnlock($querystring)
async function getGpgUnlock(qs) {
    // If there's an explicit parameter in the querystring, rely on that
    if (qs == 'gpg=1') {
        gpgUnlock = true
    } else if (qs == 'gpg=0') {
        gpgUnlock = false
    } else {
        // Need to request if the server supports GPG
        gpgUnlock = undefined
        try {
            // Check if one of the keys is a GPG key
            const res = await Request('/api/repo/key')
            gpgUnlock = false
            if (res && res.keys) {
                for (let i = 0; i < res.keys.length; i++) {
                    if (res.keys[i] && res.keys[i].type == 'gpg') {
                        gpgUnlock = true
                        break
                    }
                }
            }
        }
        catch (err) {
            console.error('Caught exception:', err)
        }
    }
}

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
    // Hide the form
    showForm = false

    try {
        // Make the unlock request
        await Request('/api/repo/unlock', {postData})
    
        // Hide the passphrase
        passphrase = ''
    
        // Reset the cache
        $fileList = null
        $operationResult = null
    
        // On success, the app will receive an "unlocked" message from the SW
        // The handler for that message will also refresh the app info cache and go to the / route
    }
    catch (err) {
        // Show the form again
        showForm = true
        passphrase = ''

        // Re-throw the error
        throw err
    }
}
</script>
