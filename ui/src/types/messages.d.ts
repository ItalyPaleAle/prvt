/** Messages between service workers and clients */
interface ServiceWorkerMessage {
    /** Name of the message */
    message: string
    /** Other key-value pairs */
    [other: string]: any
}
