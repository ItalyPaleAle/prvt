type ServiceWorkerMessageTypes = 'connected' |
    'off' |
    'set-master-key' |
    'set-theme' |
    'set-wasm' |
    'theme' |
    'wasm' |
    'unlocked'

/** Messages between service workers and clients */
interface ServiceWorkerMessage {
    /** Name of the message */
    message: ServiceWorkerMessageTypes
    /** Other key-value pairs */
    [other: string]: any
}
