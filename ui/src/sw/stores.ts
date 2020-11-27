interface SWStores {
    /** Flag indicating whether wasm mode is enabled */
    wasm: boolean
    /** Holds the RepoIndex singleton */
    index?: RepoIndex
    /** Master key for the repo, unwrapped */
    masterKey?: Uint8Array
    /** ID of the key used to unlock the repo */
    keyId?: string
}

// List of stores
export default <SWStores> {
    wasm: false,
    index: undefined,
    masterKey: undefined,
    keyId: undefined
}
