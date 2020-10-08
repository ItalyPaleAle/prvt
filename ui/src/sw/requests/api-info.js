// Stores
import stores from '../stores'

/**
 * Handler for the /api/info requests.
 * This intercepts the response and if the repo is unlocked via Wasm, sets the correct "repoUnlocked" flag.
 *
 * @param {Request} req - Request object from the client
 * @returns {Response} Response object for the request
 */
export default async function(req) {
    // Submit the request as-is
    let res = await fetch(req)

    // Check if the repo is unlocked
    if (stores.masterKey) {
        // Read the response
        const data = await res.json()
        if (!data) {
            throw Error('Response is empty')
        }

        // Set the repoUnlocked flag
        data.repoUnlocked = true

        // Get repo stats to set file count
        const stats = await stores.index.stat()
        data.fileCount = stats.fileCount

        // While in Wasm mode (at least for now), we are always in read-only mode
        data.readOnly = true

        // Rebuild the Response object
        const headers = new Headers()
        headers.set('Content-Type', 'application/json')
        res = new Response(JSON.stringify(data), {headers})
    }

    return res
}
