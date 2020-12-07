{#if opt.type == 'string' || opt.type == 'path'}
  <!-- Eventually, path should be changed to a file picker -->
  <label class="mb-4 block">
    <span class="mb-1 inline-block">{opt.label}:</span>
    <input
      class="bg-shade-neutral appearance-none border-2 border-shade-200 rounded w-full py-2 px-4 text-text-300 leading-tight focus:outline-none focus:bg-shade-100 focus:border-accent-200"
      name="{opt.name}"
      type={opt.private ? 'password' : 'text'}
      value={opt.default || ''}
      placeholder={opt.default || ''}
      {required}
      on:change={validate}
      on:keyup={validate}
    />
    {#if opt.description}
      <span class="text-xs leading-snug mt-1 mx-1 inline-block">{opt.description}</span>
    {/if}
  </label>
{:else if opt.type == 'bool'}
  <label class="mb-4 block">
    <input type="checkbox" name="{opt.name}" checked={opt.default == '1'} />
    {opt.label}
  </label>
{/if}

<script lang="ts">
// Props
export let opt: APIFSOptionsRule
export let required = false

$: validateRegex = (opt?.validate) ? new RegExp(opt.validate) : null

function validate(event: Event) {
    if (!validateRegex) {
        return
    }
    const el = event.target as HTMLInputElement
    el.setCustomValidity(
        validateRegex.test(el.value)
        ? ''
        : (opt.validateMessage || 'Please match the requested format')
    )
}
</script>
