
<h1 class="font-sans font-bold break-normal text-gray-700 px-2 text-xl mt-12 lg:mt-0 md:text-2xl">/{path}</h1>
<hr class="bg-gray-300 my-5">
<FileList path={path} />

<script>
import FileList from '../components/FileList.svelte'

// Props for this view
export let params = {}
let path = ''
$: path = decodeURIComponent(params && params.wild || '')
</script>