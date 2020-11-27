/**
 * Key type: either a passphrase or GPG
 */
type RepoKeyType = 'passphrase' | 'gpg'

/**
 * Request object for POST /api/repo/unlock
 */
interface APIRepoUnlockRequest {
    type: RepoKeyType
    passphrase?: string
}

/**
 * Response object for POST /api/repo/unlock
 */
interface APIRepoUnlockResponse {
    type: RepoKeyType
    keyId: string
}
