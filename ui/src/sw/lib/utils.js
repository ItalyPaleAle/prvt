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
