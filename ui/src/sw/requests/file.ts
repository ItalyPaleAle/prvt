// Stores
import stores from '../stores'

/**
 * Handler for the /file requests, which fetches a file and decrypts it.
 * This supports Range requests too.
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

    // Request and decrypt the file
    const response = await Prvt.decryptRequest(
        stores.masterKey,
        req
    )
    if (!response) {
        throw Error('Response from decryptPackages is empty')
    }
    else if (typeof response == 'object' && response instanceof Error) {
        throw response
    }

    return response
}
