import {writable} from 'svelte/store'
import type {Writable} from 'svelte/store'
import type {SvelteComponent} from 'svelte'

export type Modal = Writable<
    {
        component: typeof SvelteComponent
        props?: Record<string,any>
    } | null
>

// This store is a flag used to display a modal
export const modal: Modal = writable(null)

export type OperationResult = Writable<
    {
        title: string
        message: string
        list: any[]
    } | null
>

// This store is used to store the results of an operation
export const operationResult: OperationResult = writable(null)

export type FileList = Writable<
    {
        folder: string
        list: any[]
    } | null
>

// This stores the latest list loaded
export const fileList: FileList = writable(null)

// This stores is true when Wasm is enabled
export const wasm: Writable<boolean> = writable(false)

// This stores controls whether to hide the store name from the footer
export const showStoreName: Writable<boolean> = writable(true)
