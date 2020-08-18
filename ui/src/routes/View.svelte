<PageTitle {title}>
  <a class="py-2 px-4 bg-shade-neutral shadow text-accent-200 hover:bg-shade-100" href="{url + '?dl=1'}">
    <i class="fa fa-cloud-download" aria-hidden="true" title="Download"></i>
    <span class="sr-only">Download</span>
  </a>
</PageTitle>

{#if type !== null}
  {#if type == 'text' || type == 'code'}
    {#await fetch(url).then((result) => result.text()) then text}
        <pre class="w-full text-sm whitespace-pre-wrap">{text}</pre>
    {/await}
  {:else if type == 'image'}
    <img src={url} alt={title} class="w-full h-auto" />
  {:else if type == 'video'}
    {#await requesting then _}
      <!-- svelte-ignore a11y-media-has-caption -->
      <video autoplay controls class="w-full h-auto">
        <source src={url} type={metadata.mimeType} />
      </video>
    {/await}
  {:else if type == 'audio'}
    {#await requesting then _}
      <!-- svelte-ignore a11y-media-has-caption -->
      <audio autoplay controls class="w-full">
        <source src={url} type={metadata.mimeType} />
      </audio>
    {/await}
  {:else if type == 'pdf'}
    <object type="application/pdf"
      data={url}
      class="w-full max-h-screen" style="height: 36rem;"
      title={title} />
  {:else}
    <a href={url + '?dl=1'}>Download</a>
  {/if}
{/if}

<script>
// Utils
import {fileType} from '../utils'
import {Request} from '../request'

// Components
import PageTitle from '../components/PageTitle.svelte'

// Props
export let params = {}

// State
let title
let metadata
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
    metadata = {}
    type = null
    url = null
    requesting = new Promise(() => {})
}

// Request the file's metadata, which contains the type and name
function requestMetadata(fileId) {
    // Request the full metadata
    requesting = Request('/api/metadata/' + fileId)
        .then((obj) => {
            metadata = obj
            title = metadata.name
            type = fileType(metadata.mimeType)
        })
}
</script>
