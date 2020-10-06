/* global URL_PREFIX */

import fileHandler from './requests/file'

// List of fetch requests to intercept and their handlers
// Path can either be a string, which matches the pathname's prefix, or a regular expression matching the pathname
const requests = [
    {
        path: '/file',
        handler: fileHandler
    }
]

/**
 * Sets up all handlers for fetch requests we want to intercept
 *
 * @param {Event} event - Event object; only "fetch" events are handled
 */
export default function(event) {
    // Only handle fetch events
    if (event.type != 'fetch') {
        return
    }

    // Only capture requests to the API server
    if (URL_PREFIX) {
        if (!event.request.url.startsWith(URL_PREFIX)) {
            return
        }
    }

    // Get the URL
    const url = new URL(event.request.url)

    // Check if we have a match
    for (let i = 0; i < requests.length; i++) {
        const e = requests[i]
        // If path is a string, match the prefix
        // If it's a RegExp, match the entire pathname
        if (
            (typeof e.path == 'string' && url.pathname.startsWith(e.path)) ||
            (typeof e.path == 'object' && e.path instanceof RegExp && url.pathname.match(e.path))
        ) {
            // Intercept the request
            event.respondWith(e.handler(event.request))
            // We found a match, so abort the loop
            break
        }
    }
}
