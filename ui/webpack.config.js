const webpack = require('webpack')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const {CleanWebpackPlugin} = require('clean-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const SriPlugin = require('webpack-subresource-integrity')
const {InjectManifest} = require('workbox-webpack-plugin')
const path = require('path')
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin

const mode = process.env.NODE_ENV || 'development'
const prod = mode == 'production'
const analyze = process.env.ANALYZE == '1'
const appVersion = process.env.APP_VERSION || 'dev'

const htmlMinifyOptions = {
    collapseWhitespace: true,
    conservativeCollapse: true,
    removeComments: true,
    collapseBooleanAttributes: true,
    decodeEntities: true,
    html5: true,
    keepClosingSlash: false,
    processConditionalComments: true,
    removeEmptyAttributes: true
}

// List of pages
const pageList = {
    index: {
        dist: 'index.html',
        chunks: ['index'],
        html: path.resolve(__dirname, 'src/index/index.html'),
        entry: [path.resolve(__dirname, 'src/index/main.js')]
    },
    app: {
        dist: 'app.html',
        chunks: ['shared', 'vendor', 'app'],
        html: path.resolve(__dirname, 'src/app/index.html'),
        entry: [path.resolve(__dirname, 'src/app/main.ts')]
    },
}

// Entry points
const entry = {}
Object.keys(pageList).map((key) => {
    entry[key] = pageList[key].entry
})

// Plugins: html-webpack-plugin
const addPlugins = Object.keys(pageList).map((key) => {
    if (!pageList[key].html) {
        return null
    }
    return new HtmlWebpackPlugin({
        chunks: pageList[key].chunks,
        filename: pageList[key].dist,
        template: pageList[key].html,
        minify: prod ? htmlMinifyOptions : false,
    })
}).filter((val) => val !== null)

module.exports = {
    entry,
    resolve: {
        mainFields: ['svelte', 'browser', 'style', 'module', 'main'],
        extensions: ['.ts', '.mjs', '.js', '.svelte']
    },
    output: {
        path: path.resolve(__dirname, 'dist/'),
        publicPath: '/',
        filename: prod ? '[name].[contenthash:8].js' : '[name].js',
        chunkFilename: prod ? '[name].[contenthash:8].js' : '[name].js',
        crossOriginLoading: 'anonymous'
    },
    optimization: {
        usedExports: true,
        moduleIds: 'deterministic',
        runtimeChunk: false,
        splitChunks: {
            cacheGroups: {
                default: false,
                defaultVendors: false,
                // Contains all CSS (for all pages) and the shared code
                shared: {
                    name: 'shared',
                    test: /\.css$|src[\\/]shared[\\/]/,
                    chunks: 'all',
                    enforce: true,
                    priority: 30
                },
                // Contains all libraries, which are less likely to change as frequently as the rest of the code
                vendor: {
                    test: /[\\/]node_modules[\\/]/,
                    name: 'vendor',
                    chunks: 'all',
                    enforce: true,
                    priority: 20
                }
            }
        }
    },
    module: {
        // Do not parse wasm files and two other files which contain "require" calls for Node.js compatibility (Go WebAssembly runtime)
        noParse: /\.wasm$|wasm_exec\.js$/,
        rules: [
            {
                test: /\.(svelte)$/,
                exclude: [],
                use: {
                    loader: 'svelte-loader',
                    options: {
                        hotReload: true,
                        dev: !prod,
                    }
                }
            },
            {
                test: /\.ts$/,
                exclude: /node_modules/,
                use: {
                    loader: 'ts-loader'
                }
            },
            {
                test: /\.css$/,
                use: [
                    prod ? MiniCssExtractPlugin.loader : 'style-loader',
                    {loader: 'css-loader', options: {importLoaders: 1}},
                    'postcss-loader'
                ]
            },
            {
                test: /\.wasm$/,
                use: {
                    loader: 'base64-loader'
                },
                type: 'javascript/auto'
            },
            {
                test: /\.(eot|svg|ttf|woff|woff2)$/,
                use: {
                    loader: 'file-loader',
                    options: {
                        name: 'fonts/[name].[ext]'
                    }
                }
            }
        ]
    },
    plugins: [
        // Cleanup dist folder
        new CleanWebpackPlugin({
            cleanOnceBeforeBuildPatterns: ['**/*', '!assets', '!assets/*']
        }),

        // Extract CSS
        new MiniCssExtractPlugin({
            filename: '[name].[contenthash:8].css'
        }),

        // Definitions
        new webpack.DefinePlugin({
            PRODUCTION: prod,
            APP_VERSION: JSON.stringify(appVersion), 
            URL_PREFIX: process.env.URL_PREFIX ? JSON.stringify(process.env.URL_PREFIX) : null,
        }),

        // Enable subresource integrity check
        new SriPlugin({
            hashFuncNames: ['sha384'],
            enabled: prod,
        }),

        // Include the bundle analyzer only when mode is "analyze"
        ...(analyze ? [
            new BundleAnalyzerPlugin()
        ] : []),

        // Service worker
        new InjectManifest({
            swSrc: './src/sw/sw.ts',
            // In dev mode, include fonts only
            // Note that this never pre-caches the WASM file, as that is not part of the output of Webpack
            include: prod ? undefined : [/^fonts/],
        }),
    ].concat(addPlugins), // Add other plugins
    mode,
    devServer: {
        contentBase: path.join(__dirname, 'dist'),
        port: 3000
    },
    devtool: prod ? false : 'source-map'
}
