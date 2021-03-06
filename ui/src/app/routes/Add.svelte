<PageTitle title="Add" backButton={destination ? '#/tree' + destination : '#/tree/'} />

{#if error}
  <div class="ml-2 w-full max-w-md">
    <ErrorBox message={error} on:close={() => error = null} />
  </div>
{/if}

{#if running}
  <Spinner title="Uploading" />
{:else}
  <div class="w-full max-w-md bg-shade-neutral shadow p-4 ml-2 mb-6">
    <div class="sm:flex sm:items-center">
      <div class="sm:w-1/3">
        <label class="block text-text-300 sm:text-right mb-1 sm:mb-0 pr-4" for="destination">
          Destination folder
        </label>
      </div>
      <div class="sm:w-2/3">
        <input class="bg-shade-200 appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 text-text-300 leading-tight focus:outline-none focus:bg-shade-neutral focus:border-accent-200" id="destination" type="text" bind:value={destination} />
      </div>
    </div>
    <p class="text-xs sm:w-2/3 sm:mr-0 sm:ml-auto">Type the folder where the file should be uploaded. If it doesn't exist, it will be created.</p>
  </div>

  <div class="w-full max-w-md bg-shade-neutral shadow p-4 ml-2 mb-6">
    <ul class="flex border-b border-shade-300 mb-4">
      <li class="mr-3 cursor-pointer">
        <span class={addType == 'upload' ? activeTabStyle : idleTabStyle} on:click={() => addType = 'upload'}>Upload file</span>
      </li>
      <li class="mr-3 cursor-pointer">
        <span class={addType == 'local' ? activeTabStyle : idleTabStyle} on:click={() => addType = 'local'}>Add from the local disk</span>
      </li>
    </ul>
    {#if addType == 'upload'}
      <div class="sm:flex sm:items-center mb-6">
        <div class="sm:w-1/3">
          <label class="block text-text-300 sm:text-right mb-1 sm:mb-0 pr-4" for="uploadfile">
            File
          </label>
        </div>
        <div class="sm:w-2/3">
          <input class="bg-shade-200 appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 text-text-300 leading-tight focus:outline-none focus:bg-shade-neutral focus:border-accent-200" multiple id="uploadfile" type="file" />
        </div>
      </div>
      <div class="sm:flex sm:items-center">
        <div class="sm:w-1/3"></div>
        <div class="sm:w-2/3">
          <button class="shadow bg-accent-200 hover:bg-accent-100 focus:shadow-outline focus:outline-none text-shade-neutral font-bold py-2 px-4 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            type="button"
            disabled={running}
            on:click={uploadHandler}>
            Upload
          </button>
        </div>
      </div>
    {:else if addType == 'local'}
      <div class="sm:flex sm:items-center">
        <div class="sm:w-1/3">
          <label class="block text-text-300 sm:text-right mb-1 sm:mb-0 pr-4" for="localpath">
            Path
          </label>
        </div>
        <div class="sm:w-2/3">
          <input class="bg-shade-200 appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 text-text-300 leading-tight focus:outline-none focus:bg-shade-neutral focus:border-accent-200" id="localpath" type="text" />
        </div>
      </div>
      <p class="text-xs sm:w-2/3 sm:mr-0 sm:ml-auto  mb-6">Type the path to the file or folder in your local disk.</p>
      <div class="sm:flex sm:items-center">
        <div class="sm:w-1/3"></div>
        <div class="sm:w-2/3">
          <button class="shadow bg-accent-200 hover:bg-accent-100 focus:shadow-outline focus:outline-none text-shade-neutral font-bold py-2 px-4 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            type="button"
            disabled={running}
            on:click={addLocalHandler}>
            Add
          </button>
        </div>
      </div>
    {/if}
  </div>
{/if}

<script>
// Utils
import {encodePath, cleanPath} from '../lib/utils'
import {Request} from '../../shared/request'

// Components
import PageTitle from '../components/PageTitle.svelte'
import ErrorBox from '../components/ErrorBox.svelte'
import Spinner from '../components/Spinner.svelte'

// Libraries
import {push} from 'svelte-spa-router'

// Stores
import {operationResult, fileList} from '../stores'

// Props for this view
export let params = {}
let path = ''
let destination = ''
let error = null
let addType = 'upload'
let running = false

// Classes for the active and idle tab
const activeTabStyle = 'inline-block py-2 px-4 border rounded-t border-shade-200 bg-shade-200 text-text-200'
const idleTabStyle = 'inline-block py-2 px-4 border rounded-t border-shade-neutral text-text-100 hover:border-shade-100 hover:bg-shade-100'

// Clean the path
$: path = cleanPath(params && params.wild)

// Destination
$: destination = '/' + path

// Handler for all requests
function requestHandler(body) {
    // Only one request at a time
    if (running) {
        return
    }

    // Ensure the destination starts with /
    let dest = destination
    if (dest && dest.charAt(0) != '/') {
        dest = '/' + dest
    }

    error = null
    running = true

    // Upload the file
    // This request has no timeout, because we might be uploading a large file
    return Request('/api/tree' + encodePath(dest), {
            method: 'POST',
            body,
            timeout: 0
        })
        .then((list) => {
            if (!list || !Array.isArray(list) || !list.length) {
                if (list && list.error) {
                    return Promise.reject(list.error)
                }
                throw Error('Invalid response')
            }

            running = false

            $fileList = null
            $operationResult = {
                title: 'Added',
                message: 'File(s) have been added',
                list
            }
            push('/tree' + dest)
        })
        .catch((e) => {
            running = false
            error = e
        })
}

// Handler for file upload
function uploadHandler() {
    // Request body
    const file = document.getElementById('uploadfile')
    if (!file || !file.files || !file.files.length) {
        error = 'No file selected'
        return
    }
    const body = new FormData()
    for (let i = 0; i < file.files.length; i++) {
        body.append('file', file.files[i])  
    }

    // Send the request
    return requestHandler(body)
}

// Handler for adding files from local disk
function addLocalHandler() {
    // Request body
    const path = document.getElementById('localpath')
    if (!path || !path.value) {
        error = 'Value for path is empty'
        return
    }
    const body = new FormData()
    body.set('localpath', path.value)

    // Send the request
    return requestHandler(body)
}
</script>
