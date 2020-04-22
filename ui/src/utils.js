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
