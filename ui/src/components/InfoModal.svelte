<div class="p-4">
  {#if metadata}
    <div class="flex flex-col justify-between">
      <div class="text-2xl truncate mb-4 text-accent-300">
        <i class="fa {fileTypeIcon(metadata.mimeType)} fa-fw" aria-hidden="true"></i>
        <span class="ml-1" title="{metadata.name}" aria-label="{metadata.name}">{metadata.name}</span>
      </div>
      {#if metadata.folder}
        <div class="truncate mb-2 ml-4">
          <i class="fa fa-folder-open-o fa-fw" aria-hidden="true"></i>
          <span class="ml-1" title="{metadata.folder}" aria-label="{metadata.folder}">{metadata.folder}</span>
        </div>
      {/if}
      {#if metadata.size}
        <div class="truncate mb-2 ml-4">
          <i class="fa fa-database fa-fw" aria-hidden="true"></i>
          <span class="ml-1" title="{metadata.size} bytes" aria-label="{size}">{size}</span>
        </div>
      {/if}
      <div class="truncate mb-2 ml-4">
        <i class="fa fa-clock-o fa-fw" aria-hidden="true"></i>
        <span class="ml-1">{format(date, 'PPpp')}</span>
      </div>
    </div>
  {:else}
    No file selected
  {/if}
</div>
<script>
// Utils
import {fileTypeIcon, formatSize} from '../utils'
import format from 'date-fns/format'

// Props
export let element = null
let requesting = null

// Metadata: this is pre-populated with data from the element prop, but then we request the fulll metadata
let metadata = null
$: requestMetadata(element)

// Date and size
$: size = (metadata && metadata.size) ? formatSize(metadata.size) : null
$: date = (metadata && metadata.date) ? new Date(metadata.date) : null

// Request the full metadata
function requestMetadata(el) {
    // While we request the metadata, pre-popuate it with what we already have
    metadata = el

    // Request the full metadata
    requesting = fetch('/api/metadata/' + el.fileId)
        // Get response as JSON
        .then((resp) => {
            return resp.json()
        })
        .then((obj) => {
            metadata = obj
        })
}
</script>
