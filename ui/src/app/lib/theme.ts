import type {Writable} from 'svelte/store'

// Default theme
const defaultTheme = 'auto-blue'

// List of supported themes
export const themes = [
    'auto-blue',
    'auto-orange',
    'dark-blue',
    'dark-orange',
    'light-blue',
    'light-orange',
    'midnight'
]

// Return the current theme
export function getTheme() {
    const theme = localStorage.getItem('theme')
    if (!theme || themes.indexOf(theme) == -1) {
        return defaultTheme
    }
    return theme
}

// Set a new theme
export function setTheme(theme: string) {
    localStorage.setItem('theme', theme)
}

// Copied from Svelte, as it's not exported
type Subscriber<T> = (value: T) => void

// Object that implements the svelte/store contract
const subscriptions: Subscriber<string>[] = []
const theme: Writable<string> = {
    subscribe: (sub) => {
        subscriptions.push(sub)
        if (sub) {
            const theme = getTheme()
            sub('theme-' + theme)
        }

        return () => {
            const index = subscriptions.indexOf(sub)
            if (index != -1) {
                subscriptions.splice(index, 1)
            }
        }
    },

    set: (val) => {
        // Set the default theme if empty
        if (!val || themes.indexOf(val) < 0) {
            val = defaultTheme
        }
        // Update in local storage
        setTheme(val)
        // Notify all subscribers
        for (let i = 0; i < subscriptions.length; i++) {
            if (subscriptions[i]) {
                subscriptions[i]('theme-' + val)
            }
        }
    },

    update: (_) => {
        throw Error('Method update is not supported in this store')
    }
}

// Set up the theme at runtime
theme.subscribe((theme) => {
    // Remove existing theme tags
    document.body.classList.forEach((val) => {
        if (val && val.indexOf('theme-') === 0) {
            document.body.classList.remove(val)
        }
    })

    // Add the new theme
    document.body.classList.add(theme)
})

// Export the theme store as default
export default theme
