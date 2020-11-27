<div class="flex {justify} items-center w-full mb-10 lg:mb-12">
  <!-- See https://github.com/sveltejs/svelte/issues/4546 for warning at runtime when in development mode -->
  {#if backButton}
    <TitleBarButton name="Back" icon="fa-chevron-left" href={backButton} />
  {/if}
  <slot name="title">
    <h1 title={title} aria-label={title}
      class="font-bold break-normal text-text-300 px-2 text-2xl md:text-3xl truncate">
      {title}
    </h1>
  </slot>
  <slot name="right"></slot>
</div>

<script>
// Components
import TitleBarButton from '../components/TitleBarButton.svelte'

export let title = ''
export let backButton = false

$: justify = backButton ? 'justify-start' : 'justify-between'
</script>
