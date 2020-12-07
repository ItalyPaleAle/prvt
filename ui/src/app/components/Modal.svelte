<div class="fixed top-0 inset-0 px-4 py-4 sm:flex sm:items-end sm:justify-end" style="z-index: 99;">
  <div class="fixed inset-0 transition-opacity">
    <div class="absolute inset-0 bg-overlay" on:click={close}></div>
  </div>
  <div class="modal-box bg-shade-base rounded-lg overflow-x-hidden overflow-y-scroll shadow-xl transform transition-all w-full sm:w-2/3 md:w-1/2 lg:w-2/5 flex flex-col justify-start"
    role="dialog" aria-modal="true">
    <div class="w-full p-2 flex items-end justify-end border-b border-shade-200 flex-none">
      <span class="mx-2 text-xl cursor-pointer"
        aria-label="Close dialog" title="Close dialog"
        on:click={close}>
        <i class="fa fa-times" aria-hidden="true"></i>
        <span class="sr-only">Close dialog</span>
      </span>
    </div>
    {#if $modal?.component}
      <div class="p-4 flex-1">
        <svelte:component this={$modal?.component} {...($modal?.props || {})} />
      </div>
    {/if}
  </div>
</div>

<style>
.modal-box {
  /* Fallback */
  height: 100%;
  height: calc(100vh - calc(100vh - 100%))
}
</style>

<script lang="ts">
import {modal} from '../stores'

function close() {
    $modal = null
}
</script>
