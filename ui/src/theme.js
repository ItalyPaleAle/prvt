// Default theme
const defaultTheme = 'auto-blue'

// List of supported themes
export const themes = [
    'auto-blue',
    'auto-orange',
    'dark-blue',
    'dark-orange',
    'midnight',
    'light-blue',
    'light-orange'
]

// Return the current theme
export function getTheme() {
    let theme = localStorage.getItem('theme')
    if (!theme || themes.indexOf(theme) == -1) {
        return defaultTheme
    }
    return theme
}

// Set a new theme
export function setTheme(theme) {
    localStorage.setItem('theme', theme)
}

// Object that implements the svelte/store contract
let subscriptions = []
export default {
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
        setTheme(val)
        for (let i = 0; i < subscriptions.length; i++) {
            if (subscriptions[i]) {
                subscriptions[i]('theme-' + val)
            }
        }
    }
}
