<div class="flex justify-between content-center mt-8 lg:mt-0 ">
  <h1 class="font-bold break-normal text-text-300 px-2 text-xl md:text-2xl truncate" title={title} aria-label={title}>{title}</h1>
  <slot></slot>
</div>
<hr class="bg-shade-300 my-5">

<script>
export let title = ''
</script>
