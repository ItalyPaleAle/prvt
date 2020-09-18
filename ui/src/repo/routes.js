import {wrap, replace} from 'svelte-spa-router'

// Components
import ConnectionList from './routes/ConnectionList.svelte'

// Route definition object
export default {
    '/': wrap(null, null, () => {
        replace('/repo')
        return false
    }),
    '/repo': ConnectionList,
}
