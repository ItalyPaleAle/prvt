// Utils
import {JSONResponse, IsUUID} from '../lib/utils'

// Stores
import stores from '../stores'

/**
 * Handler for the GET /api/metadata/:file request.
 * This returns the metadata for a file, either by its path or its ID (UUID)
 *
 * @param req Request object from the client
 * @returns Response object for the request
 */
export default async function(req: Request): Promise<Response> {
    // Only GET requests are supported
    const method = req.method
    if (method != 'GET') {
        throw Error('Invalid request method')
    }

    // Ensure we have the master key
    if (!stores.masterKey || !stores.index) {
        throw Error('Repository is not unlocked')
    }

    // Get the file path from the URL
    const reqPath = (new URL(req.url)).pathname
    if (!reqPath || !reqPath.startsWith('/api/metadata/')) {
        throw Error('Invalid request path')
    }
    let file = reqPath.substr(14) || ''

    // Get the file UUID and other data from the index
    let indexEl: ListItem | null = null

    // Check if we have a UUID
    if (IsUUID(file)) {
        // Ensure the file exists
        indexEl = await stores.index.getFileById(file)
        if (!indexEl || !indexEl.fileId) {
            throw Error('File not found')
        }
    }
    else {
        // Treat this like a path
        // Ensure path starts with /
        if (file.charAt(0) != '/') {
            file = '/' + file
        }

        // Get the file to get its ID
        indexEl = await stores.index.getFileByPath(file)
        if (!indexEl || !indexEl.fileId) {
            throw Error('File not found')
        }
    }

    // Request the metadata
    const metadata = await Prvt.getFileMetadata(stores.masterKey, indexEl.fileId)

    // Combine the metadata and the data from the index to get the same response as the APIs
    const pos = indexEl.path.lastIndexOf('/') + 1
    const res: APIFileMetadataResponse = {
        fileId: indexEl.fileId,
        folder: indexEl.path.substr(pos),
        name: metadata.name,
        date: indexEl.date,
        mimeType: metadata.mimeType,
        size: metadata.size,
        digest: indexEl.digest,
    }

    // Return a Response object with the list of files
    return JSONResponse(res)
}
