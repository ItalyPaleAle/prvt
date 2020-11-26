<div class="break-all">
  <div class="flex text-2xl mb-4 text-accent-300">
    <span class="flex-grow-0"><i class="fa {fileTypeIcon(metadata.mimeType)} fa-fw" aria-hidden="true"></i></span>
    <span class="pl-2 flex-grow-1">{metadata.name}</span>
  </div>
  {#if metadata.folder}
    <div class="mb-2 ml-4">
      <i class="fa fa-folder-open-o fa-fw" aria-hidden="true"></i>
      <span class="ml-1">{metadata.folder}</span>
    </div>
  {/if}
  {#if metadata.size !== undefined && metadata.size !== null}
    <div class="mb-2 ml-4">
      <i class="fa fa-database fa-fw" aria-hidden="true"></i>
      <span class="ml-1" title="{metadata.size} bytes" aria-label="{size}">{size}</span>
    </div>
  {/if}
  {#if date}
    <div class="mb-2 ml-4">
      <i class="fa fa-clock-o fa-fw" aria-hidden="true"></i>
      <span class="ml-1">{date}</span>
    </div>
  {/if}
</div>

<script>
// Utils
import {fileTypeIcon, formatSize} from '../lib/utils'
import {Request} from '../../shared/request'
import format from 'date-fns/format'

// Props
export let element = null

// Metadata: this is pre-populated with data from the element prop, but then we request the fulll metadata
let metadata = {}
$: requestMetadata(element)

// Date and size
$: size = (metadata && metadata.size) ? formatSize(metadata.size) : '0 bytes'
$: date = (metadata && metadata.date) ? format(new Date(metadata.date), 'PPpp') : ''

// Request the full metadata
function requestMetadata(el) {
    // While we request the metadata, pre-popuate it with what we already have
    metadata = el

    // Request the full metadata
    Request('/api/metadata/' + el.fileId)
        .then((obj) => {
            if (obj.size === undefined || obj.size === null) {
                obj.size = 0
            }
            metadata = obj
        })
}
</script>
