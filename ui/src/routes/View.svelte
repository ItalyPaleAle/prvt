{#if path}
  <PageTitle>
    <span slot="side">
      <TitleBarButton name="Download" icon="fa-download" href={url + '?dl=1'} />
    </span>
    <span slot="title">
      <Path {path} />
    </span>
  </PageTitle>
{:else}
  <PageTitle {title}>
    <span slot="side">
      <TitleBarButton name="Download" icon="fa-download" href={url + '?dl=1'} />
    </span>
  </PageTitle>
{/if}

{#if type !== null}
  <div class="py-1 px-3">
    {#if type == 'text' || type == 'code'}
      {#await fetch(url).then((result) => result.text()) then text}
        <pre class="w-full text-sm whitespace-pre-wrap">{text}</pre>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox {url}/>
      {/await}
    {:else if type == 'image'}
      <img src={url} alt={title} class="w-full h-auto" />
    {:else if type == 'video'}
      {#await requesting then _}
        <!-- svelte-ignore a11y-media-has-caption -->
        <video autoplay controls class="w-full h-auto">
          <source src={url} type={mimeType} />
        </video>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox {url}/>
      {/await}
    {:else if type == 'audio'}
      {#await requesting then _}
        <!-- svelte-ignore a11y-media-has-caption -->
        <audio autoplay controls class="w-full">
          <source src={url} type={mimeType} />
        </audio>
      {:catch err}
        <ErrorBox title="Error requesting file" message={err || 'Unknwon error'} />
        <DownloadBox {url}/>
      {/await}
    {:else if type == 'pdf'}
      <object type="application/pdf"
        data={url}
        class="w-full max-h-screen" style="height: 36rem;"
        title={title} />
    {:else}
      <DownloadBox {url}/>
    {/if}
  </div>
{/if}

<script>
// Utils
import {fileType} from '../utils'
import {Request} from '../request'

// Stores
import {fileList} from '../stores'

// Components
import PageTitle from '../components/PageTitle.svelte'
import DownloadBox from '../components/DownloadBox.svelte'
import ErrorBox from '../components/ErrorBox.svelte'
import TitleBarButton from '../components/TitleBarButton.svelte'
import Path from '../components/Path.svelte'

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
        url = '/file/' + params.fileId
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
