/**
 * Response from GET /api/info
 */
interface APIRepoInfoResponse {
    name?: string
    version?: string
    buildId?: string
    buildTime?: string
    commitHash?: string
    runtime?: string
    info?: string
    readOnly?: boolean
    repoSelected?: boolean
    repoUnlocked?: boolean
    storeType?: string
    storeAccount?: string
    repoId?: string
    repoVersion?: number
    files?: number
    gpgUnlock?: boolean
}
