import {writable, readable, derived} from 'svelte/store'
import {Request} from './lib/request'

// This stores the name of the app being used (e.g. "app", "repo", etc)
export const currentApp = writable(null)

// This store is a flag used to display a modal
export const modal = writable(null)

// This stores the info about the app/server
export const appInfo = readable({}, (set) => {
    Request('/api/info')
        .then(set)
        .catch((err) => {
            // Log the error, but don't halt the execution
            // eslint-disable-next-line no-console
            console.error('Error while requesting app info', err)
        })
})

// This returns true if we're in a read-only server
export const readOnly = derived(appInfo, ($appInfo) => $appInfo && $appInfo.readOnly)
