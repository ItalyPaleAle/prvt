<div class="flex flex-col justify-between h-full">
  <div class="break-all">
    <div class="flex text-2xl break-all mb-4 text-accent-300">
      <span class="flex-grow-0"><i class="fa fa-chevron-right fa-fw" aria-hidden="true"></i></span>
      <span class="pl-2 flex-grow-1">{name}</span>
    </div>
    {#await requesting}
      <p class="ml-4">Requesting</p>
    {:then info}
      <div class="ml-4">
        <div class="mb-3 ml-4 flex flex-row items-start">
          <i class="fa fa-link fa-fw flex-grow-0 mt-1" aria-hidden="true"></i>
          <span class="ml-3 flex-1">
            Location:<br />
            <span class="text-text-300">{info.storeType} :: {info.storeAccount}</span>
          </span>
        </div>
        <div class="mb-3 ml-4 flex flex-row items-start">
          <i class="fa fa-tag fa-fw flex-grow-0 mt-1" aria-hidden="true"></i>
          <span class="ml-3 flex-1">
            Repository ID:<br />
            <pre class="text-text-300 text-sm">{info.repoId}</pre>
          </span>
        </div>
        <div class="mb-3 ml-4 flex flex-row items-start">
          <i class="fa fa-key fa-fw flex-grow-0 mt-1" aria-hidden="true"></i>
          <span class="ml-3 flex-1">
            Repository version:<br />
            <pre class="text-text-300">{info.repoVersion}</pre>
          </span>
        </div>
      </div>
    {:catch err}
      <p class="ml-4">{err}</p>
    {/await}
  </div>
  <div class="mt-8 flex items-center justify-around flex-wrap">
      <button type="button"
        class="w-11/12 sm:w-3/5 p-2 my-2 flex-grow-0 bg-shade-neutral rounded shadow text-alert hover:bg-shade-100"
        on:click={() => remove(name)}>
          <i class="fa fa-trash fa-fw" aria-hidden="true"></i>
          Remove connection
        </button>
  </div>
</div>

<script>
// Utils
import {Request} from '../../shared/request'

// Props
export let name = null
export let remove = null

// Repository info, which is requested from the server
let requesting = null
$: requestInfo(name)

// Request the info
function requestInfo(name) {
    requesting = Request('/api/connection/' + name)
}
</script>
