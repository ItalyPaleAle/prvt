import {writable} from 'svelte/store'

// This store is a flag used to display a modal
export const modal = writable(null)

// This store is used to store the results of an operation
export const operationResult = writable(null)

// This stores the latest list loaded
export const fileList = writable(null)

// This stores is true when Wasm is enabled
export const wasm = writable(false)

// This stores controls whether to hide the store name from the footer
export const showStoreName = writable(true)
