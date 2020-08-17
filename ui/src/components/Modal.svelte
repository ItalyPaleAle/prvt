<div class="fixed bottom-0 inset-0 px-4 py-4 sm:flex sm:items-end sm:justify-end" style="z-index: 99;">
  <div class="fixed inset-0 transition-opacity">
    <div class="absolute inset-0 bg-overlay" on:click={close}></div>
  </div>
  <div class="bg-shade-base rounded-lg overflow-x-hidden overflow-y-scroll shadow-xl transform transition-all h-full w-full sm:w-2/3 md:w-1/2 lg:w-2/5 flex flex-col justify-start" role="dialog" aria-modal="true" aria-labelledby="modal-headline">
    <div class="w-full p-2 flex items-end justify-end border-b border-shade-200">
      <span class="mx-2 text-xl cursor-pointer"
        aria-label="Close dialog" title="Close dialog"
        on:click={close}>
        <i class="fa fa-times" aria-hidden="true"></i>
      </span>
    </div>
    {#if $modal && $modal.component}
        <svelte:component this={$modal.component} {...($modal.props || {})} />
    {/if}
  </div>
</div>

<script>
import {modal} from '../stores'

function close() {
    $modal = null
}
</script>
