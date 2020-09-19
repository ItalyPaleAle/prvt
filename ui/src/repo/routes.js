import {wrap, replace} from 'svelte-spa-router'

// Components
import ConnectionList from './routes/ConnectionList.svelte'
import UnlockRepo from './routes/UnlockRepo.svelte'

// Route definition object
export default {
    '/': wrap(ConnectionList, null, () => {
        replace('/repo')
        return false
    }),
    '/repo': ConnectionList,
    '/unlock': UnlockRepo
}
