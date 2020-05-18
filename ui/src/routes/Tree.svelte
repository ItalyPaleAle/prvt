<PageTitle title="{'/' + path}">
  <a class="py-2 px-4 bg-white shadow text-gray-700 hover:bg-gray-100" href="{'#/add/' + path}">
    <i class="fa fa-cloud-upload" aria-hidden="true" title="Add a file"></i>
    <span class="sr-only">Add a file</span>
  </a>
</PageTitle>
<FileList path={path} />

<script>
import PageTitle from '../components/PageTitle.svelte'
import FileList from '../components/FileList.svelte'

import {cleanPath} from '../utils'

// Props for this view
export let params = {}
let path = ''

// Clean the path
$: path = cleanPath(params && params.wild)
</script>