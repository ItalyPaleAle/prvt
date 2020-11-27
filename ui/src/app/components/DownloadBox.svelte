<a class="py-3 px-6 my-3 mx-6 text-xl bg-shade-neutral shadow text-accent-200 hover:bg-shade-100" href="{(urlPrefix || '') + '/file/' + fileId + '?dl=1'}">
    <i class="fa fa-download" aria-hidden="true"></i>
    Download
</a>

<script>
export let fileId

// URL prefix - expose the constant here so it can be used in the Svelte markup above
const urlPrefix = URL_PREFIX
</script>
