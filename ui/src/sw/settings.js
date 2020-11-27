import {idbPromisify} from './lib/utils'

// Database connection
let _db = null

/**
 * Returns the DB connection
 * @returns {Promise<IDBDatabase>} DB connection
 */
export async function DB() {
    // If the database is open, return it
    if (_db) {
        return _db
    }

    // Open a connection to the IndexedDB db
    const openReq = indexedDB.open('prvt-db', 1)
    openReq.onupgradeneeded = () => {
        // Create the settings store on the first run
        openReq.result.createObjectStore('settings')
    }
    _db = await idbPromisify(openReq)
    return _db
}

/**
 * Gets a setting by its key
 * @param {string} key - Key name
 * @returns {Promise<any>} Value for the setting
 */
export async function Get(key) {
    // Get a connection to the IndexedDB
    const db = await DB()

    // Get the value
    const tx = db.transaction('settings', 'readonly')
    const settingsStore = tx.objectStore('settings')
    return idbPromisify(settingsStore.get(key))
}

/**
 * Sets a new value for the settings
 * @param {string} key - Key name
 * @param {any} value - Value to set
 */
export async function Set(key, value) {
    // Get a connection to the IndexedDB
    const db = await DB()

    // Get the value
    const tx = db.transaction('settings', 'readwrite')
    const settingsStore = tx.objectStore('settings')
    return idbPromisify(settingsStore.put(value, key))
}
