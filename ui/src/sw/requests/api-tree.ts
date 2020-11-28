// Utils
import {JSONResponse} from '../lib/utils'

// Stores
import stores from '../stores'

/**
 * Handler for the /api/tree/:path requests.
 *
 * The GET method returns the list of files in the folder.
 * The POST method is not yet implemented; eventually, it will allow uploading files.
 *
 * @param req Request object from the client
 * @returns Response object for the request
 */
export default async function(req: Request): Promise<Response> {
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
    if (!reqPath || !reqPath.startsWith('/api/tree')) {
        throw Error('Invalid request path')
    }
    let path = decodeURIComponent(reqPath.substr(9) || '/')

    // Ensure path starts with /
    if (path.charAt(0) != '/') {
        path = '/' + path
    }

    // Get the list of files
    const list = await stores.index.listFolder(path) || []
    
    // Return a Response object with the list of files
    return JSONResponse(list)
}
