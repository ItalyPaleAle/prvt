import {writable, readable, derived} from 'svelte/store'
import {Request} from '../shared/lib/request'

// This store is a flag used to display a modal
export const modal = writable(null)

// This store is used to store the results of an operation
export const operationResult = writable(null)

// This stores the latest list loaded
export const fileList = writable(null)

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
