import {writable} from 'svelte/store'

// This store is a flag used to close any dropdown menu that might be open
export const dropdown = writable(false)
