{#if path}
  <PageTitle backButton={'#/tree/' + path.slice(0, Math.max(path.lastIndexOf('/'), 0))}>
    <span class="flex-initial" style="min-width: 0" slot="title">
      <Path {path} />
    </span>
  </PageTitle>
{:else}
  <PageTitle {title} backButton="#/tree/" />
{/if}

{#if type !== null}
  <div class="py-1 px-3">
    {#if type == 'text' || type == 'code'}
      {#await fetch(url).then((result) => result.text())}
        <Spinner />
      {:then text}
        <pre class="w-full px-4 text-sm whitespace-pre-wrap">{text}</pre>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox fileId={params.fileId}/>
      {/await}
    {:else if type == 'image'}
      <img src={url} alt={title} class="w-auto h-auto max-w-full max-h-90vh mx-auto" />
    {:else if type == 'video'}
      {#await requesting}
        <Spinner />
      {:then}
        <!-- svelte-ignore a11y-media-has-caption -->
        <video autoplay controls class="w-auto h-auto max-w-full max-h-90vh mx-auto">
          <source src={url} type={mimeType} />
        </video>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox fileId={params.fileId}/>
      {/await}
    {:else if type == 'audio'}
      {#await requesting}
        <Spinner />
      {:then}
        <!-- svelte-ignore a11y-media-has-caption -->
        <audio autoplay controls class="w-full">
          <source src={url} type={mimeType} />
        </audio>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox fileId={params.fileId}/>
      {/await}
    {:else if type == 'pdf'}
      <object type="application/pdf"
        data={url}
        class="w-full max-h-screen" style="height: 36rem;"
        title={title} />
    {:else}
      <p>This file is in a format we can't preview</p>
    {/if}
  </div>
  <div class="my-10 pl-6">
    <DownloadBox fileId={params.fileId}/>
  </div>
{/if}

<script>
// Utils
import {fileType} from '../lib/utils'
import {Request} from '../../shared/request'

// Stores
import {fileList} from '../stores'

// Components
import PageTitle from '../components/PageTitle.svelte'
import DownloadBox from '../components/DownloadBox.svelte'
import ErrorBox from '../components/ErrorBox.svelte'
import Path from '../components/Path.svelte'
import Spinner from '../components/Spinner.svelte'

// Props
export let params = {}

// State
let title
let path
let mimeType
let type
let url
let requesting
reset()

$: {
    if (params && params.fileId) {
        url = (URL_PREFIX || '') + '/file/' + params.fileId
        requestMetadata(params.fileId)
    }
    else {
        reset()
    }
}

// Reset state
function reset() {
    title = 'View'
    path = null
    mimeType = {}
    type = null
    url = null
    requesting = new Promise(() => {})
}

// Request the file's metadata, which contains the type and name
function requestMetadata(fileId) {
    // Check if we have the metadata we need in cache
    const cache = $fileList
    if (cache && cache.list) {
        for (let i = 0; i < cache.list.length; i++) {
            const el = cache.list[i]
            if (el && el.fileId == fileId) {
                // If the cached data doesn't include a mime type, the file was likely added using an odler version of prvt, which did not add the metadata to the index
                // So, we need to skip using the cached data and request the metadata
                if (el.mimeType === undefined) {
                    break
                }

                title = el.path
                path = cache.folder + '/' + el.path
                mimeType = el.mimeType
                type = fileType(mimeType)
                if (path.charAt(0) == '/') {
                    // Remove the / from the beginning of the path, if present
                    path = path.slice(1)
                }

                // Set requesting to a Promise that is immediately resolved, then stop the function
                requesting = Promise.resolve()
                return
            }
        }
    }

    // Request the full metadata
    requesting = Request('/api/metadata/' + fileId)
        .then((obj) => {
            title = obj.name
            path = obj.folder + obj.name
            mimeType = obj.mimeType
            type = fileType(mimeType)
            if (path.charAt(0) == '/') {
                // Remove the / from the beginning of the path, if present
                path = path.slice(1)
            }
        })
}
</script>
