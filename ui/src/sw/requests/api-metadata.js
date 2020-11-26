/// <reference path="../Prvt.d.ts" />
/* global Prvt */

// Utils
import {JSONResponse, IsUUID, BytesToHex} from '../lib/utils'

// Stores
import stores from '../stores'

/**
 * Handler for the GET /api/metadata/:file request.
 * 
 * This returns the metadata for a file, either by its path or its ID (UUID)
 *
 * @param {Request} req - Request object from the client
 * @returns {Response} Response object for the request
 */
export default async function(req) {
    // Get the method of the request
    // Only GET requests are implemented for now
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
    let indexEl = null

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
    const res = {
        fileId: indexEl.fileId,
        folder: indexEl.path.substr(pos),
        name: metadata.name,
        date: indexEl.date,
        mimeType: metadata.mimeType,
        size: metadata.size,
        // Encode to hex because Uint8Array doesn't survive the JSON encoding
        digest: BytesToHex(indexEl.digest),
    }

    // Return a Response object with the list of files
    return JSONResponse(res)
}
