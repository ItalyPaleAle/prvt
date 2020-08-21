<PageTitle>
  <!-- See: https://css-tricks.com/flexbox-truncated-text/ -->
  <span class="flex-initial" style="min-width: 0" slot="title">
    <Path {path} />
  </span>
  <span class="flex-none" slot="right">
    <TitleBarButton name="Add files" icon="fa-cloud-upload" href={'#/add/' + path} />
  </span>
</PageTitle>
<FileList path={path} />

<script>
import PageTitle from '../components/PageTitle.svelte'
import FileList from '../components/FileList.svelte'
import TitleBarButton from '../components/TitleBarButton.svelte'
import Path from '../components/Path.svelte'

import {cleanPath} from '../utils'

// Props for this view
export let params = {}
let path = ''

// Clean the path
$: path = cleanPath(params && params.wild)
</script>