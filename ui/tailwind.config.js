module.exports = {
    theme: {
        extend: {
            colors: {
                'shade-neutral': 'var(--color-shade-neutral)',
                'shade-base': 'var(--color-shade-base)',
                'shade-100': 'var(--color-shade-100)',
                'shade-200': 'var(--color-shade-200)',
                'shade-300': 'var(--color-shade-300)',
                'text-base': 'var(--color-text-base)',
                'text-100': 'var(--color-text-100)',
                'text-200': 'var(--color-text-200)',
                'text-300': 'var(--color-text-300)',
                'accent-100': 'var(--color-accent-100)',
                'accent-200': 'var(--color-accent-200)',
                'accent-300': 'var(--color-accent-300)',
                'error-100': 'var(--color-error-100)',
                'error-200': 'var(--color-error-200)',
                'error-300': 'var(--color-error-300)',
                'success-100': 'var(--color-success-100)',
                'success-200': 'var(--color-success-200)',
                'success-300': 'var(--color-success-300)',
                overlay: 'var(--color-overlay)',
                alert: 'var(--color-alert)',
            },
            maxHeight: {
                '90vh': '90vh'
            }
        },
        container: {
            center: true,
        }
    },
    variants: {
        opacity: ['responsive', 'hover', 'focus', 'disabled'],
    },
    plugins: [],
    future: {
        removeDeprecatedGapUtilities: true,
        purgeLayersByDefault: true
    },
    purge: false
}
