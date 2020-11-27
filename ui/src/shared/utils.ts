/**
 * Returns a Promise that resolves after a certain amount of time, in ms
 *
 * @param time Time to wait in ms; if 0, this just executes the action on the next tick of the event loop
 * @returns Promise that resolves after a certain amount of time
 */
export function waitPromise(time: number): Promise<void> {
    return new Promise((resolve) => {
        setTimeout(resolve, time || 0)
    })
}

/**
 * Sets a timeout on a Promise, so it's automatically rejected if it doesn't resolve within a certain time.
 *
 * @param promise Promise to execute
 * @param timeout Timeout in ms
 * @returns Promise with a timeout
 */
export function timeoutPromise<T>(promise: Promise<T>, timeout: number): Promise<T> {
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
