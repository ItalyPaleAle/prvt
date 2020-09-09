import svelte from 'rollup-plugin-svelte'
import replace from '@rollup/plugin-replace'
import resolve from 'rollup-plugin-node-resolve'
import commonjs from 'rollup-plugin-commonjs'
import livereload from 'rollup-plugin-livereload'
import postcss from 'rollup-plugin-postcss'
import {terser} from 'rollup-plugin-terser'
import copy from 'rollup-plugin-copy'
import autoPreprocess from 'svelte-preprocess'

const production = !process.env.ROLLUP_WATCH

export default {
    input: 'src/main.js',
    output: {
        sourcemap: true,
        format: 'iife',
        name: 'ui',
        file: 'dist/bundle.js'
    },
    plugins: [
        // Replace
        replace({
            'env.APP_VERSION': JSON.stringify(process.env.APP_VERSION || 'canary')
        }),

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
                css.write('css/components.css')
            }
        }),

        // PostCSS
        postcss({
            minimize: production,
            extract: 'css/bundle.css'
        }),

        // Copy static files
        copy({
            targets: [
                // Index file
                {src: 'src/index.html', dest: 'dist'},
                {src: 'node_modules/fork-awesome/fonts/', dest: 'dist'}
            ]
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
