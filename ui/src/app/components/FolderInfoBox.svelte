<div class="break-all">
  <div class="flex text-2xl break-all mb-4 text-accent-300">
    <span class="flex-grow-0"><i class="fa fa-folder-open-o fa-fw" aria-hidden="true"></i></span>
    <span class="pl-2 flex-grow-1">{element.name}</span>
  </div>
  {#if list && list.length}
    <div class="mb-2 ml-4">
      <div class="mb-2 ml-4">
        <i class="fa fa-file-o fa-fw" aria-hidden="true"></i>
        <span class="ml-1">{list.filter(el => !el.isDir).length} files</span>
      </div>
      <div class="mb-2 ml-4">
        <i class="fa fa-folder-o fa-fw" aria-hidden="true"></i>
        <span class="ml-1">{list.filter(el => el.isDir).length} folders</span>
      </div>
    </div>
  {/if}
</div>

<script>
// Utils
import {Request} from '../../shared/request'

// Props
export let element = null
export let path = ''

// File list, which is requested from the server
let list = []
$: requestContents(element)

// Request the full list
function requestContents(el) {
    // Request the full list
    Request('/api/tree/' + ((path && path != '/') ? path + '/' : '') + el.path)
        .then((response) => {
            list = response
        })
}
</script>
