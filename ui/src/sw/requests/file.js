/* global Prvt */

// Stores
import stores from '../stores'

/**
 * Handler for the /file requests, which fetches a file and decrypts it.
 * This supports Range requests too.
 *
 * @param {Request} req - Request object from the client
 * @returns {Response} Response object for the request
 */
export default async function(req) {
    const response = await Prvt.decryptRequest(
        new Uint8Array(stores.masterKey),
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
