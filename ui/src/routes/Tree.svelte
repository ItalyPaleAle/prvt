<PageTitle title="{'/' + path}">
  <span slot="side">
    <a  href="{'#/add/' + path}"
      class="py-2 px-4 bg-shade-neutral shadow text-accent-200 hover:bg-shade-100">
      <i class="fa fa-cloud-upload" aria-hidden="true" title="Add files"></i>
      <span class="sr-only">Add files</span>
    </a>
  </span>
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