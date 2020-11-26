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
 * Checks if a string represents a UUID (any version)
 * @param {string} str - String to check
 * @returns {boolean} true if the string represents a UUID
 */
export function IsUUID(str) {
    return !!str.match(/^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$/i)
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

/** 
 * Encodes a Uint8Array to hex string
 * 
 * @param {Uint8Array} data - Data to encode
 * @returns {string} Hex representation of the data
 */
export function BytesToHex(data) {
    return [...data]
        .map(b => b.toString(16).padStart(2, '0'))
        .join('')
}

