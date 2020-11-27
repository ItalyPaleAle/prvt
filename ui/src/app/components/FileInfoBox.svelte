<div class="break-all">
  <div class="flex text-2xl mb-4 text-accent-300">
    <span class="flex-grow-0">
      <i class="fa {fileTypeIcon(metadata.mimeType)} fa-fw" aria-hidden="true" title={'File type: ' + type}></i>
      <span class="sr-only">{'File type: ' + type}</span>
    </span>
    <span class="pl-2 flex-grow-1">{metadata.name}</span>
  </div>
  {#if metadata.folder}
    <div class="flex mb-2 ml-4">
      <span class="flex-grow-0">
        <i class="fa fa-folder-open-o fa-fw" aria-hidden="true" title="Folder"></i>
        <span class="sr-only">Folder</span>
      </span>
      <span class="ml-1 flex-grow-1">{metadata.folder}</span>
    </div>
  {/if}
  {#if metadata.size}
    <div class="flex mb-2 ml-4">
      <span class="flex-grow-0">
        <i class="fa fa-database fa-fw" aria-hidden="true" title="Size"></i>
        <span class="sr-only">Size</span>
      </span>
      <span class="ml-1 flex-grow-1" title="{metadata.size} bytes" aria-label="{size}">{size}</span>
    </div>
  {/if}
  {#if date}
    <div class="flex mb-2 ml-4">
      <span class="flex-grow-0">
        <i class="fa fa-clock-o fa-fw" aria-hidden="true" title="Date added"></i>
        <span class="sr-only">Date added</span>
      </span>
      <span class="ml-1 flex-grow-1">{date}</span>
    </div>
  {/if}
  {#if metadata.digest}
    <div class="flex mb-2 ml-4">
      <span class="flex-grow-0">
        <i class="fa fa-biometric fa-fw" aria-hidden="true" title="SHA-256 checksum"></i>
        <span class="sr-only">SHA-256 checksum</span>
      </span>
      <span class="ml-1 p-1 flex-grow-1 text-xs font-mono overflow-x-scroll whitespace-no-wrap">
        {metadata.digest}
      </span>
    </div>
  {/if}
</div>

<script>
// Utils
import {fileType, fileTypeIcon, formatSize} from '../lib/utils'
import {Request} from '../../shared/request'
import format from 'date-fns/format'

// Props
export let element = null

// Metadata: this is pre-populated with data from the element prop, but then we request the fulll metadata
let metadata = {}
$: requestMetadata(element)

// Size, date, type
$: size = (metadata && metadata.size) ? formatSize(metadata.size) : '0 bytes'
$: date = (metadata && metadata.date) ? format(new Date(metadata.date), 'PPpp') : ''
$: type = (metadata && metadata.mimeType) ? fileType(metadata.mimeType) : 'unknown'

// Request the full metadata
function requestMetadata(el) {
    // While we request the metadata, pre-popuate it with what we already have
    metadata = el

    // Request the full metadata
    Request('/api/metadata/' + el.fileId)
        .then((obj) => metadata = obj)
}
</script>
