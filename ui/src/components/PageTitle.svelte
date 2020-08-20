<div class="flex justify-between content-center mt-8 lg:mt-0 w-full">
  <!-- See: https://github.com/sveltejs/svelte/issues/4546 -->
  <slot name="title">
    <h1 title={title} aria-label={title}
      class="font-bold break-normal text-text-300 px-2 text-xl md:text-2xl truncate">
      {title}
    </h1>
  </slot>
  <slot name="side"></slot>
</div>
<hr class="bg-shade-300 my-5">

<script>
export let title = ''
</script>
