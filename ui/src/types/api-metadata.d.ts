/**
 * Response from GET /api/metadata/:file
 */
interface APIFileMetadataResponse {
    fileId: string
    folder: string
    name: string
    date?: string
    mimeType?: string
    size?: number
    digest?: string
}
