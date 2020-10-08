/**
 * Returns a Promise that resolves after a certain amount of time, in ms
 * @returns {Promise<void>} Promise that resolves after a certain amount of time
 */
export function waitPromise(time) {
    return new Promise((resolve) => {
        setTimeout(resolve, time || 0)
    })
}

/**
 * Sets a timeout on a Promise, so it's automatically rejected if it doesn't resolve within a certain time.
 * @param {Promise<T>} promise - Promise to execute
 * @param {number} timeout - Timeout in ms
 * @returns {Promise<T>} Promise with a timeout
 */
export function timeoutPromise(promise, timeout) {
    return Promise.race([
        waitPromise(timeout).then(() => {
            throw new TimeoutError('Promise has timed out')
        }),
        promise
    ])
}

/**
 * Error returned by timed out Promises in timeoutPromise
 */
export class TimeoutError extends Error {}
