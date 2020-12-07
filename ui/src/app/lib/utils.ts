// Enables or disables Wasm
export function enableWasm(enabled: boolean) {
    if (!navigator.serviceWorker.controller) {
        // Should never happen
        return
    }

    // Request a change in wasm enablement
    navigator.serviceWorker.controller.postMessage({
        message: 'set-wasm',
        enabled
    })
}

// Cleans the path from the URL
export function cleanPath(path: string): string {
    path = path || ''
    // Ensure the path starts with a /
    if (path.charAt(0) == '/') {
        path = path.slice(1)
    }
    // Ensure the path does not end with a /
    if (path.charAt(path.length - 1) == '/') {
        path = path.slice(0, -1)
    }
    return path
}

// Encodes the path so it can put in a URL for requests to the server
export function encodePath(path: string): string {
    // Run "encodeURIComponent" and then revert back %2F to /
    return encodeURIComponent(path).replace(/%2[Ff]/g, '/')
}

// Performs a deep cloning of an object (as long as it can be serialized as JSON)
export function cloneObject<T>(obj: T): T {
    return JSON.parse(JSON.stringify(obj))
}

// Formats a size in bytes into human-readable
export function formatSize(sz: number): string {
    const units = ['bytes', 'KB', 'MB', 'GB', 'TB']
    let unit = 0
    while (sz > 1000 && unit < 4) {
        sz /= 1024
        unit++
    }
    let result = sz + ''
    if (unit > 0) {
        result = Number(Math.round(Number(sz + 'e2')) + 'e-2').toString()
    }
    result += ' ' + units[unit]

    return result
}

// Returns the type for a given file mime type
export function fileType(mimeType: string): string {
    if (!mimeType) {
        return ''
    }

    // Remove what's after the ;
    const pos = mimeType.indexOf(';')
    if (pos > 1) {
        mimeType = mimeType.slice(0, pos)
    }

    switch (mimeType) {
        case 'application/pdf':
        case 'application/x-pdf':
            return 'pdf'

        case 'application/zip':
        case 'application/x-bzip':
        case 'application/x-bzip2':
        case 'application/gzip':
        case 'application/x-tar':
        case 'application/x-7z-compressed':
        case 'application/vnd.rar':
            return 'archive'

        case 'text/plain':
            return 'text'

        case 'application/epub+zip':
            return 'book'

        case 'text/html':
        case 'text/javascript':
        case 'text/css':
        case 'text/json':
        case 'text/xml':
        case 'text/yaml':
        case 'application/json':
        case 'application/php':
        case 'application/x-sh':
        case 'application/x-csh':
        case 'application/xhtml+xml':
        case 'application/xml':
        case 'application/x-freearc':
            return 'code'

        case 'application/vnd.ms-powerpoint':
        case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
        case 'application/vnd.oasis.opendocument.presentation':
            return 'presentation'

        case 'text/csv':
        case 'application/vnd.ms-excel':
        case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
        case 'application/vnd.oasis.opendocument.spreadsheet':
            return 'spreadsheet'

        case 'application/msword':
        case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
        case 'application/vnd.oasis.opendocument.text':
        case 'application/rtf':
            return 'richtext'

        default:
            if (mimeType.startsWith('image/')) {
                // All image types
                return 'image'
            }
            else if (mimeType.startsWith('audio/')) {
                // All audio types
                return 'audio'
            }
            else if (mimeType.startsWith('video/')) {
                // All video types
                return 'video'
            }
            return ''
    }
}

// Returns the icon for the given file mime type
export function fileTypeIcon(mimeType: string): string {
    switch (fileType(mimeType)) {
        case 'pdf':
            return 'fa-file-pdf-o'
        case 'archive':
            return 'fa-file-archive-o'
        case 'text':
            return 'fa-file-text-o'
        case 'book':
            return 'fa-file-epub-o'
        case 'code':
            return 'fa-file-code-o'
        case 'presentation':
            return 'fa-file-powerpoint-o'
        case 'spreadsheet':
            return 'fa-file-excel-o'
        case 'richtext':
            return 'fa-file-word-o'
        case 'image':
            return 'fa-file-image-o'
        case 'audio':
            return 'fa-file-audio-o'
        case 'video':
            return 'fa-file-video-o'
        default:
            return 'fa-file-o'
    }
}
