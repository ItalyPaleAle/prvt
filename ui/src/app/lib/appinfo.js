import {Request} from '../../shared/request'

/**
 * @typedef {Object} AppInfoObject
 * @property {string} [name] - Server app name
 * @property {string} [version] - Server app version
 * @property {string} [buildId] - Server app build ID
 * @property {string} [buildTime] - Server app build time
 * @property {string} [commitHash] - Server app build source commit hash
 * @property {string} [runtime] - Server runtime (Go version)
 * @property {string} [info] - Info message
 * @property {boolean} [readOnly] - If true, the repository is in read-only mode
 * @property {boolean} [repoSelected] - If true, a repository has been selected (but not necessarily unlocked)
 * @property {boolean} [repoUnlocked] - If true, the selected repository has been unlocked
 * @property {string} [storeType] - Type of the store used
 * @property {string} [storeAccount] - Account of the store used
 * @property {boolean} [repoId] - ID of the repository selected
 * @property {number} [repoVersion] - Version of the repository
 * @property {number} [files] - Number of files
 * @property {boolean} [gpgUnlock] - If true, repository can be unlocked with a GPG key
 */

/**
 * Offers mechanisms to access info about the application and server
 * Implements the Svelte readable store contracts
 */
export class AppInfo {
    /**
     * Initializes the object
     */
    constructor() {
        this._cached = null
        this._subscriptions = []
        this._requesting = null
    }

    /**
     * Returns the app info object, using the cached one if available
     * @returns {AppInfoObject} App info
     * @async
     */
    async get() {
        // If there's a request pending, wait for it
        if (this._requesting) {
            await this._requesting
        }

        // If the data is already in cache
        if (this._cached) {
            // Clone the object before returning it
            return this._cached ? Object.assign({}, this._cached) : {}
        }

        // Request data
        const info = await this.update()
        return info || {}
    }

    /**
     * Forces an update of the cached app info object and returns it
     * @returns {Promise<AppInfoObject|null>} App info
     * @async
     */
    update() {
        this._requesting = (async () => {
            let val
            try {
                val = await Request('/api/info')
            }
            catch (err) {
                // Log the error, but don't halt the execution
                // eslint-disable-next-line no-console
                console.error('Error while requesting app info', err)
                val = null
            }
            this._cached = val

            // Remove the pending request semaphore
            this._requesting = null

            // Notify al subscribers
            this._notify()

            return val
        })()

        return this._requesting
    }

    /**
     * Returns true if the repository is open in read-only mode
     * @returns {boolean} True if the repository is open in read-only mode
     * @async
     */
    async isReadOnly() {
        const info = await this.get()
        return info && info.readOnly
    }

    /**
     * Subscription function for implementing the Svelte store contract
     * @param {Function} handler - Subscription function, as per the Svelte store contract
     * @returns {Function} Unsubscribe function, as per the Svelte store contract
     */
    subscribe(handler) {
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
