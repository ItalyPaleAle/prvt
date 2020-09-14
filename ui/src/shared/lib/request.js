import {timeoutPromise, TimeoutError} from './utils'

const requestTimeout = 5000 // 5s

/* global URL_PREFIX */

/**
 * Performs API requests.
 */
export function Request(url, options) {
    if (!options) {
        options = {}
    }

    // URL prefix
    if (URL_PREFIX) {
        url = URL_PREFIX + url
    }

    // Set the options
    const reqOptions = {
        method: 'GET',
        cache: 'no-store',
        credentials: 'omit',
        headers: new Headers()
    }

    // HTTP method
    if (options.method) {
        reqOptions.method = options.method
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

        const body = new FormData()
        for (const key in options.postData) {
            if (!Object.prototype.hasOwnProperty.call(options.postData, key)) {
                continue
            }
            body.append(key, options.postData[key])
        }

        reqOptions.body = body
    }

    // Headers
    if (options.headers && typeof options.headers == 'object') {
        for (const key in options.headers) {
            if (Object.prototype.hasOwnProperty.call(options.headers, key)) {
                reqOptions.headers.set(key, options.headers[key])
            }
        }
    }

    // Make the request
    let p = fetch(url, reqOptions)
    if (options.timeout === undefined || options.timeout === null || options.timeout > 0) {
        p = timeoutPromise(p, options.timeout || requestTimeout)
    }
    return p
        .then((response) => {
            // Read the response stream and get the data
            if (options.rawResponse) {
                return response
            }
            
            // We're expecting a JSON document
            if (!response.headers.get('content-type').match(/application\/json/i)) {
                throw Error('Response was not JSON')
            }

            // Get the JSON data from the response
            return response.json()
                .then((body) => {
                    // Check if we have a response with status code 200-299
                    if (!response || !response.ok) {
                        if (body && body.error) {
                            // eslint-disable-next-line no-console
                            console.error('Invalid response status code')
                            throw Error(body.error)
                        }
                        throw Error('Invalid response status code')
                    }

                    return body
                })
        })
        .catch((err) => {
            if (err instanceof TimeoutError) {
                throw Error('Request has timed out')
            }
            throw err
        })
}
