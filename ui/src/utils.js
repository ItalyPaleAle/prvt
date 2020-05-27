// Cleans the path from the URL
export function cleanPath(path) {
    path = path || ''
    if (path.charAt(0) == '/') {
        path = path.slice(1)
    }
    if (path.charAt(path.length - 1) == '/') {
        path = path.slice(0, -1)
    }
    return decodeURIComponent(path)
}

// Encodes the path so it can put in a URL for requests to the server
export function encodePath(path) {
    // Run "encodeURIComponent" and then revert back %2F to /
    return encodeURIComponent(path).replace(/\%2[Ff]/g, '/')
}

// Returns the icon for the given file mime type
export function fileTypeIcon(mimeType) {
    // Default is file-o
    if (!mimeType) {
        return 'fa-file-o'
    }

    // Specific types
    switch (mimeType) {
        case 'application/pdf':
        case 'application/x-pdf':
            return 'fa-file-pdf-o'

        case 'application/zip':
        case 'application/x-bzip':
        case 'application/x-bzip2':
        case 'application/gzip':
        case 'application/x-tar':
        case 'application/zip':
        case 'application/x-7z-compressed':
        case 'application/vnd.rar':
            return 'fa-file-archive-o'

        case 'text/plain':
        case 'application/rtf':
            return 'fa-file-text-o'

        case 'application/epub+zip':
            return 'fa-file-epub-o'

        case 'text/html':
        case 'text/javascript':
        case 'text/css':
        case 'text/xml':
        case 'application/json':
        case 'application/php':
        case 'application/x-sh':
        case 'application/x-csh':
        case 'application/xhtml+xml':
        case 'application/xml':
        case 'application/x-freearc':
            return 'fa-file-code-o'

        case 'application/vnd.ms-powerpoint':
        case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
        case 'application/vnd.oasis.opendocument.presentation':
            return 'fa-file-powerpoint-o'

        case 'text/csv':
        case 'application/vnd.ms-excel':
        case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
        case 'application/vnd.oasis.opendocument.spreadsheet':
            return 'fa-file-excel-o'

        case 'application/msword':
        case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
        case 'application/vnd.oasis.opendocument.text':
            return 'fa-file-word-o'

        default:
            if (mimeType.startsWith('image/')) {
                // All image types
                return 'fa-file-image-o'
            }
            else if (mimeType.startsWith('audio/')) {
                // All audio types
                return 'fa-file-audio-o'
            }
            else if (mimeType.startsWith('video/')) {
                // All video types
                return 'fa-file-video-o'
            }
            return 'fa-file-o'
    }
}
