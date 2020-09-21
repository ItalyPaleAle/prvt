import {replace} from 'svelte-spa-router'
import {wrap} from 'svelte-spa-router/wrap'

// Components
import ConnectionList from './routes/ConnectionList.svelte'
import UnlockRepo from './routes/UnlockRepo.svelte'

// Route definition object
export default {
    '/': wrap({
        route: ConnectionList,
        conditions: [() => {
            replace('/repo')
            return false
        }]
    }),
    '/repo': ConnectionList,
    '/unlock': UnlockRepo
}
