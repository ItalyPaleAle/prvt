const webpack = require('webpack')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const {CleanWebpackPlugin} = require('clean-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const SriPlugin = require('webpack-subresource-integrity')
const path = require('path')
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin

const mode = process.env.NODE_ENV || 'development'
const prod = mode == 'production'
const analyze = process.env.ANALYZE == '1'

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

// Entry points
const entry = {
    prvt: [path.resolve(__dirname, 'src/main.js')],
}

module.exports = {
    entry,
    resolve: {
        mainFields: ['svelte', 'browser', 'style', 'module', 'main'],
        extensions: ['.mjs', '.js', '.svelte']
    },
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: '[name].[hash].js',
        chunkFilename: '[name].[id].[hash].js',
        crossOriginLoading: 'anonymous'
    },
    module: {
        // Do not parse wasm files
        noParse: /\.wasm$/,
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
                test: /\.css$/,
                use: [
                    prod ? MiniCssExtractPlugin.loader : 'style-loader',
                    {loader: 'css-loader', options: {importLoaders: 1}},
                    'postcss-loader'
                ]
            },
            {
                test: /\.wasm$/,
                loaders: ['base64-loader'],
                type: 'javascript/auto'
            },
            {
                test: /\.(eot|svg|ttf|woff|woff2)$/,
                loader: 'file-loader?name=fonts/[name].[ext]'
            }
        ]
    },
    node: {
        __dirname: false,
        fs: 'empty',
        Buffer: false,
        process: false
    },
    plugins: [
        // Cleanup dist folder
        new CleanWebpackPlugin({
            cleanOnceBeforeBuildPatterns: ['**/*', '!assets', '!assets/*']
        }),

        // Extract CSS
        new MiniCssExtractPlugin({
            filename: '[name].[hash].css'
        }),

        // Definitions
        new webpack.DefinePlugin({
            PRODUCTION: prod,
            APP_VERSION: process.env.APP_VERSION ? JSON.stringify(process.env.APP_VERSION) : false, 
            URL_PREFIX: process.env.URL_PREFIX ? JSON.stringify(process.env.URL_PREFIX) : false,
        }),

        // Generate the index.html file
        new HtmlWebpackPlugin({
            filename: 'index.html',
            template: path.resolve(__dirname, 'src/main.html'),
            chunks: ['prvt'],
            minify: prod ? htmlMinifyOptions : false
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
    ],
    mode,
    devServer: {
        contentBase: path.join(__dirname, 'dist'),
        port: 3000
    },
    devtool: prod ? false : 'source-map'
}
