/**
 * Response from GET /api/fsoptions
 * This contains options for all FS
 */
type APIFSOptionsListResponse = Record<string,APIFSOptions>

/**
 * Response from GET /api/fsoptions/:fs
 * This contains options for a single FS
 */
type APIFSOptionsResponse = APIFSOptions

/**
 * Options for a FS
 */
type APIFSOptions = {
    label: string
    required: APIFSOptionsRule[]
    optional?: APIFSOptionsRule[]
}

/**
 * Rule for a FS
 */
interface APIFSOptionsRule {
    name: string
    type: 'string' | 'bool' | 'path'
    label: string
    description?: string
    default?: string
    private?: boolean

    // Extension that is not part of the API but used in the app
    validate?: string
    validateMessage?: string
}
