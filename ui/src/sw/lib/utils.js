/**
 * Returns a JavaScript Response object containing the given data.
 *
 * @param {*} data - Data that will be included in the response
 * @param {number} [status] - Status code for the response (optional - defaults to 200)
 */
export function JSONResponse(data, status) {
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    return new Response(
        JSON.stringify(data),
        {headers, status}
    )
}

/**
 * Sends a message to every client connected to this service worker
 *
 * @param {{message: string, [other: string]: unknown}} data - Message to send
 */
export async function BroadcastMessage(data) {
    const list = await self.clients.matchAll()
    list.forEach(c => {
        c.postMessage(data)
    })
}

/**
 * Convenience method that returns a Promise for an operation with IndexedDB
 * @template T
 * @param {IDBRequest<T>} req - Request
 * @returns {Promise<T>}
 */
export function idbPromisify(req) {
    return new Promise((resolve, reject) => {
        req.onerror = () => {
            // eslint-disable-next-line no-console
            console.error('IndexedDB error', req.error)
            reject(req.error)
        }
        req.onsuccess = (event) => {
            resolve(event.target.result)
        }
    })
}