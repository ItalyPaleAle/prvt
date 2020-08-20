<PageTitle {title}>
  <span slot="side">
    <a class="py-2 px-4 bg-shade-neutral shadow text-accent-200 hover:bg-shade-100" href="{url + '?dl=1'}">
      <i class="fa fa-download" aria-hidden="true" title="Download"></i>
      <span class="sr-only">Download</span>
    </a>
  </span>
</PageTitle>

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

// Props
export let params = {}

// State
let title
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
                console.log(el)
                title = el.path
                mimeType = el.mimeType
                type = fileType(mimeType)
                return Promise.resolve()
            }
        }
    }

    // Request the full metadata
    requesting = Request('/api/metadata/' + fileId)
        .then((obj) => {
            title = obj.name
            mimeType = obj.mimeType
            type = fileType(mimeType)
        })
}
</script>
