<PageTitle title="Add" />

<div class="w-full max-w-md bg-white shadow p-4 ml-2 mb-6">
  <div class="sm:flex sm:items-center">
    <div class="sm:w-1/3">
      <label class="block text-gray-700 sm:text-right mb-1 sm:mb-0 pr-4" for="destination">
        Destination folder
      </label>
    </div>
    <div class="sm:w-2/3">
      <input class="bg-gray-200 appearance-none border-2 border-gray-200 rounded w-full py-2 px-4 text-gray-700 leading-tight focus:outline-none focus:bg-white focus:border-orange-500" id="destination" type="text" bind:value={destination} />
    </div>
  </div>
  <p class="text-xs sm:w-2/3 sm:mr-0 sm:ml-auto">Type the folder where the file should be uploaded. If it doesn't exist, it will be created.</p>
</div>

<div class="w-full max-w-md bg-white shadow p-4 ml-2 mb-6">
  <ul class="flex border-b border-gray-300 mb-4">
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
        <label class="block text-gray-700 sm:text-right mb-1 sm:mb-0 pr-4" for="uploadfile">
          File
        </label>
      </div>
      <div class="sm:w-2/3">
        <input class="bg-gray-200 appearance-none border-2 border-gray-200 rounded w-full py-2 px-4 text-gray-700 leading-tight focus:outline-none focus:bg-white focus:border-orange-500" id="uploadfile" type="file" />
      </div>
    </div>
    <div class="sm:flex sm:items-center">
      <div class="sm:w-1/3"></div>
      <div class="sm:w-2/3">
        <button class="shadow bg-orange-500 hover:bg-orange-400 focus:shadow-outline focus:outline-none text-white font-bold py-2 px-4 rounded" type="button" on:click={uploadHandler}>
          Upload
        </button>
      </div>
    </div>
  {:else if addType == 'local'}
    <div class="sm:flex sm:items-center">
      <div class="sm:w-1/3">
        <label class="block text-gray-700 sm:text-right mb-1 sm:mb-0 pr-4" for="localpath">
          Path
        </label>
      </div>
      <div class="sm:w-2/3">
        <input class="bg-gray-200 appearance-none border-2 border-gray-200 rounded w-full py-2 px-4 text-gray-700 leading-tight focus:outline-none focus:bg-white focus:border-orange-500" id="localpath" type="text" />
      </div>
    </div>
    <p class="text-xs sm:w-2/3 sm:mr-0 sm:ml-auto  mb-6">Type the path to the file or folder in your local disk.</p>
    <div class="sm:flex sm:items-center">
      <div class="sm:w-1/3"></div>
      <div class="sm:w-2/3">
        <button class="shadow bg-orange-500 hover:bg-orange-400 focus:shadow-outline focus:outline-none text-white font-bold py-2 px-4 rounded" type="button" on:click={addLocalHandler}>
          Add
        </button>
      </div>
    </div>
  {/if}
</div>

<script>
// Components
import PageTitle from '../components/PageTitle.svelte'

// Libraries
import {push} from 'svelte-spa-router'

// Utils
import {cleanPath} from '../utils'

// Stores
import {operationResult} from '../stores'

// Props for this view
export let params = {}
let path = ''
let destination = ''
let addType = 'upload'

// Classes for the active and idle tab
let activeTabStyle = 'inline-block py-2 px-4 border rounded-t border-gray-200 bg-gray-200 text-orange-500'
let idleTabStyle = 'inline-block py-2 px-4 border rounded-t border-white text-gray-500 hover:border-gray-100 hover:bg-gray-100'

// Clean the path
$: path = cleanPath(params && params.wild)

// Destination
$: destination = '/' + path

// Handler for all requests
function requestHandler(body) {
    // Ensure the destination starts with /
    let dest = destination
    if (dest.charAt(0) != '/') {
        dest = '/' + dest
    }

    // Upload the file
    return fetch('/api/tree' + encodeURIComponent(dest), {
            method: 'POST',
            body
        })
        // Get response as JSON
        .then((resp) => {
            return resp.json()
        })
        .then((list) => {
            if (!list || !Array.isArray(list) || !list.length) {
                throw Error('Invalid response')
            }
            
            $operationResult = {
                title: 'Added',
                message: 'File(s) have been added',
                list
            }
            push('/tree' + dest)
        })
}

// Handler for file upload
function uploadHandler() {
    // Request body
    const file = document.getElementById('uploadfile')
    const body = new FormData()
    body.set('file', file.files[0])

    // Send the request
    return requestHandler(body)
}

// Handler for adding files from local disk
function addLocalHandler() {
    // Request body
    const path = document.getElementById('localpath')
    const body = new FormData()
    body.set('localpath', path.value)

    // Send the request
    return requestHandler(body)
}
</script>
