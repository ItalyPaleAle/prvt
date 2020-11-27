const production = process.env.NODE_ENV == 'production'

module.exports = {
    plugins: [
        require('postcss-import')(),
        require('postcss-url')(),
        require('tailwindcss'),
        ...(production ? [require('@fullhuman/postcss-purgecss')({

            // Specify the paths to all of the template files in your project
            content: [
                './dist/index.html',
                './src/**/*.svelte',
                './src/**/*.html',
                './src/**/*.js',
            ],

            // Include any special characters you're using in this regular expression
            defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || []
        })] : []),
        ...(production ? [require('cssnano')] : []),
    ],
}
