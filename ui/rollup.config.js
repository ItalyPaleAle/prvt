import svelte from 'rollup-plugin-svelte'
import resolve from 'rollup-plugin-node-resolve'
import commonjs from 'rollup-plugin-commonjs'
import livereload from 'rollup-plugin-livereload'
import postcss from 'rollup-plugin-postcss'
import {terser} from 'rollup-plugin-terser'
import autoPreprocess from 'svelte-preprocess'

const production = !process.env.ROLLUP_WATCH

export default {
    input: 'app/main.js',
    output: {
        sourcemap: true,
        format: 'iife',
        name: 'ui',
        file: 'dist/bundle.js'
    },
    plugins: [
        // Svelte
        svelte({
            // Enable run-time checks when not in production
            dev: !production,
            // PostCSS support
            preprocess: autoPreprocess({
                postcss: true
            }),
            // We'll extract any component CSS out into a separate file
            css: css => {
                css.write('dist/components.css')
            }
        }),

        // PostCSS
        postcss({
            minimize: production,
            extract: 'dist/bundle.css'
        }),

        // Support external dependencies from npm
        resolve({
            browser: true
        }),
        commonjs(),

        // Watch the `dist` directory and refresh the
        // browser on changes when not in production
        !production && livereload('dist'),

        // If we're building for production (npm run build
        // instead of npm run dev), minify
        production && terser()
    ],
    watch: {
        clearScreen: false
    }
}
