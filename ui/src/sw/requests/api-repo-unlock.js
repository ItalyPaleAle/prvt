/* global Prvt */

// Stores
import stores from '../stores'

/**
 * Handler for the /api/repo/unlock requests, which unlocks a repo.
 *
 * @param {Request} req - Request object from the client
 * @returns {Response} Response object for the request
 */
export default async function(req) {
    // Get the body from the request
    const data = await req.json()
    if (!data || data.type != 'passphrase' || !data.passphrase) {
        return
    }

    // Unlock the repo
    const result = await Prvt.unlock(data.passphrase)
    if (!result || !result.masterKey || !result.keyId) {
        throw Error('Invalid response')
    }
    stores.masterKey = result.masterKey
    stores.keyId = result.keyId

    // Get the index object
    stores.index = Prvt.getIndex(stores.masterKey)

    // Return a Response object just like the API server would for /api/repo/unlock
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    return new Response(
        JSON.stringify({
            keyId: result.keyId,
            type: 'passphrase'
        }),
        {headers}
    )
}
