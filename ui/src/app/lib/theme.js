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
    const theme = localStorage.getItem('theme')
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
const subscriptions = []
const theme = {
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
