/** Interface for the Prvt object, as defined in the Wasm code */
interface Prvt {
    /** Set a new value for the base URL where the server is located */
    setBaseURL: (baseUrl: string) => void

    /**
     * Used by Service Workers when intercepting requests: it accepts a Request object and returns a modified Response that decrypts a file from the store.
     * Supports both requests for full files and requests with a Range header.
     */
    decryptRequest: (masterKey: Uint8Array, req: Request) => Promise<Response>

    /** Returns the RepoIndex object for the current repo */
    getIndex: (masterKey: Uint8Array) => RepoIndex

    /** Returns the metadata for a given file by its ID */
    getFileMetadata: (masterKey: Uint8Array, fileId: string) => Promise<FileMetadata>

    /**
     * Try unlocking the repo with a passphrase and returns the master key
     * Note that this doesn't support unlocking with a GPG key, because Wasm can't communicate with the user's GPG agent safely
     */
    unlock: (passphrase: string) => Promise<{masterKey: Uint8Array, keyId: string}>
}

// Export the Prvt singleton
declare const Prvt: Prvt

/** Object for the index file and methods to work on it */
interface RepoIndex {
    /**
     * Triggers a refresh of the index.
     * Returns a Promise that resolves (with no value) once the new index has been fetched.
     */
    refresh: (force: boolean) => Promise<void>

    /**
     * Adds a file to the index.
     * NOT YET IMPLEMENTED.
     */
    addFile: () => void

    /**
     * Returns stats about the repo
     */
    stat: () => Promise<RepoStats>

    /** Returns the list item object for a file, searching by its path */
    getFileByPath: (path: string) => Promise<ListItem>

    /** Returns the list item object for a file, searching by its id */
    getFileById: (fileId: string) => Promise<ListItem>

     /**
     * Removes a file from the index.
     * NOT YET IMPLEMENTED.
     */
    deleteFile: () => void

    /** Returns the list of elements in a folder */
    listFolder: (pat: string) => Promise<ListItem[]>
}

/** Stats for a repo */
interface RepoStats {
    /** Number of files in the repo */
    fileCount: number
}

/** Item in a folder, as returned by the RepoIndex methods */
interface ListItem {
    /** Item name */
    path: string
    /** If true, item is a directory */
    isDir?: boolean
    /** ID of the file (for files only) */
    fileId?: string
    /** Date the file was added to the repo (for files only) */
    date?: Date
    /** Mime type of the file (for files only) */
    mimeType?: string
    /** File size (for files only) */
    size?: number
    /** SHA-256 checksum (for files only */
    digest?: Uint8Array
}

/** File metadata, as returned by Prvt.getFileMetadata */
interface FileMetadata {
    /** File name */
    name: string
    /** Mime type of the file */
    mimeType?: string
    /** File size */
    size?: number
}
