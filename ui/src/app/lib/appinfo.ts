import type {Readable} from 'svelte/store'
import {Request} from '../../shared/request'

// From Svelte, here because they're not exported
declare type Subscriber<T> = (value: T) => void
declare type Unsubscriber = () => void

/**
 * Offers mechanisms to access info about the application and server
 * Implements the Svelte readable store contracts
 */
export class AppInfo implements Readable<APIRepoInfoResponse> {
    _cached?: APIRepoInfoResponse
    _requesting?: Promise<APIRepoInfoResponse|undefined>
    _subscriptions: Subscriber<APIRepoInfoResponse>[]

    /**
     * Initializes the object
     */
    constructor() {
        this._cached = undefined
        this._subscriptions = []
        this._requesting = undefined
    }

    /**
     * Returns the app info object, using the cached one if available
     * @returns App info
     */
    async get(): Promise<APIRepoInfoResponse> {
        // If there's a request pending, wait for it
        if (this._requesting) {
            await this._requesting
        }

        // If the data is already in cache
        if (this._cached) {
            // Clone the object before returning it
            return Object.assign({}, this._cached)
        }

        // Request data
        const info = await this.update()
        return info || {}
    }

    /**
     * Forces an update of the cached app info object and returns it
     * @returns App info
     */
    update(): Promise<APIRepoInfoResponse|undefined> {
        this._requesting = (async () => {
            let val: APIRepoInfoResponse|undefined
            try {
                val = await Request('/api/info') as APIRepoInfoResponse
            }
            catch (err) {
                // Log the error, but don't halt the execution
                // eslint-disable-next-line no-console
                console.error('Error while requesting app info', err)
                val = undefined
            }
            this._cached = val

            // Remove the pending request semaphore
            this._requesting = undefined

            // Notify al subscribers
            this._notify()

            return val
        })()

        return this._requesting
    }
    
    /**
     * Resets the cached app info object if present
     */
    reset() {
        // Append to the pending request
        // If there's no pending request, just add a promise
        this._requesting = (this._requesting || Promise.resolve(undefined))
            .then(() => {
                // Reset the cache
                this._cached = undefined
            
                // Notify al subscribers
                this._notify()

                return undefined
            })
    }

    /**
     * Returns true if the repository is open in read-only mode
     * @returns True if the repository is open in read-only mode
     */
    async isReadOnly(): Promise<boolean> {
        const info = await this.get()
        return !!(info?.readOnly)
    }

    /**
     * Subscription function for implementing the Svelte store contract
     * @param handler - Subscription function, as per the Svelte store contract
     * @returns Unsubscribe function, as per the Svelte store contract
     */
    subscribe(handler: Subscriber<APIRepoInfoResponse>): Unsubscriber {
        // Record the new handler as subscriber
        this._subscriptions.push(handler)

        // Call the handler with the current value
        // Note that the contract requires a synchronous response, so the value here is read from the cache;
        // if the data has not been requested yet, this will be empty
        // (but that's ok, because the data will be requested soon and, if needed, an update will be triggered)
        handler(this._cached || {})

        // Invoke the get() method that will cache data if needed, in background
        if (!this._cached) {
            this.get()
        }

        // Returns the unsubscribe function
        return () => {
            this._subscriptions = this._subscriptions.filter(el => el !== handler)
        }
    }

    // Internal method that sends a notification to every subscriber when a new value is available in the cache
    _notify() {
        // Notify every subscriber, synchronously
        this._subscriptions.forEach((el) => {
            el(this._cached || {})
        })
    }
}

const info = new AppInfo()
export default info
