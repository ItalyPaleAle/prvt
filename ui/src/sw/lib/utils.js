/**
 * Returns a JavaScript Response object containing the given data.
 *
 * @param {*} data - Data that will be included in the response
 */
export function JSONResponse(data) {
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    return new Response(
        JSON.stringify(data),
        {headers}
    )
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