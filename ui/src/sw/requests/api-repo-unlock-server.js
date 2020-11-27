// Utils
import {BroadcastMessage} from '../lib/utils'

/**
 * Handler for the /api/repo/unlock requests, which unlocks a repo.
 * This is used for non-Wasm mode, in which the unlock happens in the server. We still need to intercept the request to notify all pages that the unlock happened
 *
 * @param {Request} req - Request object from the client
 * @returns {Response} Response object for the request
 */
export default async function(req) {
    // Submit the request as-is
    const res = await fetch(req)

    // If the status code is 2xx, then the unlock was successful
    // Note that we're not parsing the response to avoid consuming the body stream
    if (res.status >= 200 && res.status < 300) {
        // In the next tick, send a message to all clients that the repo was unlocked
        setTimeout(() => {
            BroadcastMessage({
                message: 'unlocked'
            })
        }, 0)
    }

    // Return the response as-is
    return res
}
