// Utils
import {BroadcastMessage, JSONResponse} from '../lib/utils'

// Stores
import stores from '../stores'

/**
 * Handler for the /api/repo/unlock requests, which unlocks a repo.
 * This is the handler for when in Wasm mode, in which the unlock happens in the Wasm code.
 *
 * @param req Request object from the client
 * @returns Response object for the request
 */
export default async function(req: Request): Promise<Response> {
    // Get the body from the request
    // We only support passphrases when in Wasm mode
    const data = await req.json() as APIRepoUnlockRequest
    if (!data || data.type != 'passphrase' || !data.passphrase) {
        throw Error('In wasm mode, unlock requests support passphrases only')
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

    // In the next tick, send a message to all clients that the repo was unlocked
    setTimeout(() => {
        BroadcastMessage({
            message: 'unlocked'
        })
    }, 0)

    // Return a Response object just like the API server would for /api/repo/unlock
    return JSONResponse({
        keyId: result.keyId,
        type: 'passphrase'
    })
}
