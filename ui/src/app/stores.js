import {writable} from 'svelte/store'

// This store is used to store the results of an operation
export const operationResult = writable(null)

// This stores the latest list loaded
export const fileList = writable(null)
