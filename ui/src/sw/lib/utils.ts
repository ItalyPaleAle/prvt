// Globals
declare var self: ServiceWorkerGlobalScope

/**
 * Returns a JavaScript Response object containing the given data.
 *
 * @param data Data that will be included in the response
 * @param status Status code for the response (optional - defaults to 200)
 * @returns A Response object with the correct body and headers
 */
export function JSONResponse(data: any, status?: number): Response {
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
 * @param data Message to send
 */
export async function BroadcastMessage(data: ServiceWorkerMessage): Promise<void> {
    const list = await self.clients.matchAll()
    list.forEach(c => {
        c.postMessage(data)
    })
}

/**
 * Checks if a string represents a UUID (any version)
 * @param str String to check
 * @returns true if the string represents a UUID
 */
export function IsUUID(str: string): boolean {
    return !!str.match(/^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$/i)
}

/**
 * Convenience method that returns a Promise for an operation with IndexedDB
 * @param req - Request
 * @returns Promise that resolves with the result
 */
export function idbPromisify<T>(req: IDBRequest<T>): Promise<T> {
    return new Promise((resolve, reject) => {
        req.onerror = () => {
            // eslint-disable-next-line no-console
            console.error('IndexedDB error', req.error)
            reject(req.error)
        }
        req.onsuccess = () => {
            resolve(req.result)
        }
    })
}
