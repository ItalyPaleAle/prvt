<a class="mx-3 py-3 px-4 bg-shade-neutral shadow text-accent-200 hover:bg-shade-100 md:text-lg" {href}>
  <i class="fa {icon}" aria-hidden="true" title="{name}"></i>
  <span class="sr-only">{name}</span>
</a>

<script lang="ts">
export let href: string
export let icon: string
export let name: string
</script>
