declare var URL_PREFIX: string

import {timeoutPromise, TimeoutError} from './utils'

const requestTimeout = 5000 // 5s

interface RequestOptions {
    method?: any
    headers?: any
    body?: any
    postData?: any
    timeout?: any
}

interface ErrorResponse {
    error?: string
}

/**
 * Performs API requests.
 */
export async function Request<T>(url: string, options?: RequestOptions): Promise<T> {
    if (!options) {
        options = {}
    }

    // URL prefix
    if (URL_PREFIX) {
        url = URL_PREFIX + url
    }

    // Set the options
    const reqOptions: RequestInit = {
        method: 'GET',
        cache: 'no-store',
        credentials: 'omit',
    }
    const headers = new Headers()

    // HTTP method
    if (options.method) {
        reqOptions.method = options.method
    }

    // Headers
    if (options.headers && typeof options.headers == 'object') {
        for (const key in options.headers) {
            if (Object.prototype.hasOwnProperty.call(options.headers, key)) {
                headers.set(key, options.headers[key])
            }
        }
        reqOptions.headers = headers
    }

    // Request body
    // Disallow for GET and HEAD requests
    if (options.body && reqOptions.method != 'GET' && reqOptions.method != 'HEAD') {
        reqOptions.body = options.body
    }

    // POST data, if any
    if (options.postData) {
        // Ensure method is POST
        reqOptions.method = 'POST'
        reqOptions.body = JSON.stringify(options.postData)
        headers.set('Content-Type', 'application/json')
    }
    reqOptions.headers = headers

    // Make the request
    try {
        let p = fetch(url, reqOptions)
        if (options.timeout === undefined || options.timeout === null || options.timeout > 0) {
            p = timeoutPromise(p, options.timeout || requestTimeout)
        }
        const response = await p

        // We're expecting a JSON document
        const ct = response.headers.get('content-type')
        if (!ct?.match(/application\/json/i)) {
            throw Error('Response was not JSON')
        }
    
        // Get the JSON data from the response
        const body = await response.json()

        // Check if we have a response with status code 200-299
        if (!response.ok) {
            if (body?.error) {
                // eslint-disable-next-line no-console
                console.error('Invalid response status code')
                throw Error(body.error)
            }
            throw Error('Invalid response status code')
        }
    
        return body
    }
    catch (err) {
        if (err instanceof TimeoutError) {
            throw Error('Request has timed out')
        }
        throw err
    }
}
