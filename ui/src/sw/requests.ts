// Handlers
import fileHandler from './requests/file'
import apiRepoUnlockHandlerWasm from './requests/api-repo-unlock-wasm'
import apiRepoUnlockHandlerServer from './requests/api-repo-unlock-server'
import apiInfoHandler from './requests/api-info'
import apiTreeHandler from './requests/api-tree'
import apiMetadataHandler from './requests/api-metadata'

// Utils
import {JSONResponse} from './lib/utils'

// Stores
import stores from './stores'

// Types

/** Handler function */
type HandlerFunction = ((req: Request) => Promise<Response>)

/** Object containing the definition of a request handler */
interface RequestDefinition {
    /** Path to match */
    path: string | RegExp

    /** Handler */
    handler: HandlerFunction
}

// List of fetch requests to intercept when in Wasm mode and their handlers
// Path can either be a string, which matches the pathname's prefix, or a regular expression matching the pathname
const requestsWasm: RequestDefinition[] = wrapDefinitionList([
    {
        path: '/api/info',
        handler: apiInfoHandler
    },
    {
        path: '/api/repo/unlock',
        handler: apiRepoUnlockHandlerWasm
    },
    {
        path: '/api/tree',
        handler: apiTreeHandler
    },
    {
        path: '/api/metadata',
        handler: apiMetadataHandler
    },
    {
        path: '/file',
        handler: fileHandler
    },
])

// List of fetch requests to intercept when not in Wasm mode and their handlers
const requestsServer: RequestDefinition[] = wrapDefinitionList([
    {
        path: '/api/repo/unlock',
        handler: apiRepoUnlockHandlerServer
    }
])

/**
 * Sets up all handlers for fetch requests we want to intercept
 *
 * @param event Event object; only "fetch" events are handled
 */
export default function(event: FetchEvent) {
    // Only handle fetch events
    if (event?.type != 'fetch' || !(event?.request?.url)) {
        return
    }

    // Waiting on: https://github.com/w3c/ServiceWorker/issues/1544
    /*self.addEventListener('fetch', (event) => {
        console.log('fetch', event)
        event.request.signal.addEventListener('abort', (event) => {
            console.log('aborted', event)
        })
    })*/

    // Only capture requests to the API server
    if (URL_PREFIX) {
        if (!event.request.url.startsWith(URL_PREFIX)) {
            return
        }
    }

    // Get the URL
    const url = new URL(event.request.url)

    // List of requests to match, depending on whether wasm is enabled
    const list = stores.wasm ? requestsWasm : requestsServer

    // Check if we have a match
    for (let i = 0; i < list.length; i++) {
        const e = list[i]
        // If path is a string, match the prefix
        // If it's a RegExp, match the entire pathname
        if (
            (typeof e.path == 'string' && url.pathname.startsWith(e.path)) ||
            (typeof e.path == 'object' && e.path instanceof RegExp && url.pathname.match(e.path))
        ) {
            // Intercept the request
            event.respondWith(e.handler(event.request))
            return
        }
    }
}

/**
 * Catches all errors/exceptions from the handlers and converts them to a Response with an error
 *
 * @param handler Handler function
 */
function catchErrors(handler: HandlerFunction): HandlerFunction {
    return async function(request) {
        try {
            // Do not just do "return handler()" because we want to catch exceptions here
            const res = await handler(request)
            return res
        }
        catch (err) {
            // Convert to a Response object
            return JSONResponse({
                error: err?.message
            }, 400)
        }
    }
}

/**
 * Wraps each request definition with `catchErrors`
 *
 * @param list List of definitions
 * @returns Modified list
 */
function wrapDefinitionList(list: RequestDefinition[]): RequestDefinition[] {
    return list.map((e) => {
        // Wrap every handler in "catchErrors"
        e.handler = catchErrors(e.handler)
        return e
    })
}
