import {writable} from 'svelte/store'

// These stores are flags used to close any dropdown menu or modal that might be open
export const dropdown = writable(null)
export const modal = writable(null)

// This store is used to store the results of an operation
export const operationResult = writable(null)
