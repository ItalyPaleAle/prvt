{#await requesting}
  Requesting…
{:then _}
  {#if !fs}
    <!-- Show fs selector -->
    <h1 class="flex text-2xl break-all mb-4 text-accent-300">New connection</h1>
    <p class="mb-4">Storage type:</p>
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
      {#each Object.keys(options) as el}  
        <button type="button"
          class="p-2 bg-shade-neutral rounded shadow text-accent-200 hover:bg-shade-100"
          on:click={() => {fs = el}}>
          {options[el].label}
        </button>
      {/each}
    </div>
  {:else}
    <!-- Show options -->
    <form
      name="connectionAddForm"
      on:submit|preventDefault={submit} 
      class="flex flex-col justify-between h-full">
      <div>
        <h1 class="flex text-2xl break-all mb-4 text-accent-300">New {options[fs].label} connection</h1>
        <ConnectionAddOptsForm options={options[fs]} />
      </div>
      <div class="mt-8 flex items-center justify-around flex-wrap">
        <button type="submit"
          class="w-11/12 sm:w-2/5 p-2 my-2 flex-grow-0 bg-shade-neutral rounded shadow text-accent-200 hover:bg-shade-100">
          <i class="fa fa-save fa-fw" aria-hidden="true"></i>
          Save
        </button>
      </div>
    </form>
  {/if}
{:catch err}
  Error {err}
{/await}

<script>
// Components
import ConnectionAddOptsForm from "./ConnectionAddOptsForm.svelte"

// Utils
import {Request} from '../../shared/request'

// Props
export let add = null

// Repository info, which is requested from the server
let requesting = requestFsOptions()
let options = null
let fs = null

// Request options for all fs
function requestFsOptions() {
    return Request('/api/fsoptions')
        .then((response) => {
            options = response
        })
}

// Submit the form
function submit(event) {
    event.preventDefault()
    const form = event.target
    if (form.tagName != 'FORM') {
        return false
    }

    // Collect all values
    const data = {
        type: fs
    }
    Object.keys(form.elements).forEach((key) => {
        const el = form.elements[key]
        if (el.type != "submit") {
            data[el.name] = el.value
        }
    })

    // Return the data to the parent component
    if (add) {
        add(data)
    }

    return false
}
</script>
